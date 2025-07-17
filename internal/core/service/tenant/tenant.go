package tenant

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	entity_audit "github.com/synera-br/lockari-backend-app/internal/core/entity/audit"
	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	"github.com/synera-br/lockari-backend-app/pkg/authenticator"
	"github.com/synera-br/lockari-backend-app/pkg/authorization"
	corev1 "github.com/synera-br/lockari-backend-app/pkg/core/v1"
	"github.com/synera-br/lockari-backend-app/pkg/tokengen"
	"github.com/synera-br/lockari-backend-app/pkg/utils"
)

type tenantEventService struct {
	repo          entity.TenantRepository
	authenticator authenticator.Authenticator
	tokenJWT      tokengen.TokenGenerator
	authorizer    authorization.LockariAuthorizationService
	audit         entity_audit.AuditSystemEventService
	mu            sync.Mutex
	wg            sync.WaitGroup
}

func InitializeTenantEventService(repo entity.TenantRepository, auth authenticator.Authenticator, tokenJWT tokengen.TokenGenerator, authz authorization.LockariAuthorizationService, audit entity_audit.AuditSystemEventService) (entity.TenantService, error) {

	if repo == nil {
		return nil, corev1.ErrRepositoryNotFound("TenantEventRepository")
	}

	if auth == nil {
		return nil, corev1.ErrRepositoryNotFound("Authenticator")
	}

	if tokenJWT == nil {
		return nil, corev1.ErrRepositoryNotFound("TokenGenerator")
	}

	if authz == nil {
		return nil, corev1.ErrRepositoryNotFound("AuthorizationService")
	}

	if audit == nil {
		return nil, corev1.ErrRepositoryNotFound("AuditSystemEventService")
	}

	return &tenantEventService{
		repo:          repo,
		authenticator: auth,
		tokenJWT:      tokenJWT,
		authorizer:    authz,
		audit:         audit,
	}, nil
}

// Create creates a new tenant and performs necessary operations such as creating a default group, user, and vault.
// It also sets custom claims for the user in the authentication service.
func (s *tenantEventService) Create(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {

	// Check context
	if ctx.Err() != nil {
		return nil, fmt.Errorf(corev1.ContextCancelled, ctx.Err())
	}

	s.wg.Add(1)
	defer s.wg.Wait()
	go func() {
		defer s.wg.Done()
		s.audit.Create(ctx, &entity_audit.AuditSystemEvent{
			EventType: entity_audit.EventType(tenant.EventType),
			User: entity_audit.User{
				Email: tenant.Owner.GetEmail(),
				Name:  tenant.Owner.GetUsername(),
				Uid:   tenant.Owner.Uid,
			},
			ClientInfo: entity_audit.Client{
				IpAddress: tenant.ClientInfo.IpAddress,
				UserAgent: tenant.ClientInfo.UserAgent,
			},
			Timestamp: tenant.Timestamp.Format(time.RFC3339),
		})
	}()

	// Check tenant is valid
	if tenant == nil {
		return nil, corev1.ErrGenericError("Tenant is required")
	}

	if err := tenant.IsValid(); err != nil {
		if err.Error() != entity.ErrInvalidTenantInfoID {
			return nil, corev1.ErrGenericError("Invalid tenant: " + err.Error())
		}
	}

	// Get token of application (JWT)
	appToken := utils.GetTokenFromContext(ctx)

	_, err := s.tokenJWT.Validate(appToken)
	if err != nil {
		return nil, fmt.Errorf(corev1.GenericError, err)
	}

	userToken, err := utils.GetAuthorizationFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf(corev1.GenericError, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate tenant ID if not provided
	tenantID := utils.GenerateTenant()

	// Set tenant ID at tenant struct
	tenant.SetTenantID(&tenantID)
	if tenant.ID == "" {
		return nil, corev1.ErrGenericError("Failed to generate tenant ID")
	}

	if tenant.Tenant.Name == "" {
		tenant.Tenant.SetTenantName(tenant.Owner.GetUsername())
	}

	// ##### FIRESTORE #####
	// create tenant in database
	result, err := s.createTenantInDB(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant in database: %w", err)
	}

	if result == nil {
		return nil, corev1.ErrGenericError("Failed to create tenant event")
	}

	if result.ID != tenantID {
		return nil, corev1.ErrGenericError("Tenant ID mismatch after creation")
	}

	// Set a default group
	defaultGroup := entity.NewDefaultUserGroup()
	err = s.createGroupInDB(ctx, defaultGroup, &tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create default group in database: %w", err)
	}

	// Set a default user
	member := entity.GroupMember{
		ID:   defaultGroup.ID,
		Name: defaultGroup.Name,
	}

	err = tenant.Owner.SetDefaultUser(&tenantID, &member)
	if err != nil {
		return nil, fmt.Errorf("failed to set default user in tenant: %w", err)
	}

	defaultUser := tenant.GetOwner()
	if defaultUser.Uid == "" || defaultUser.Email == "" {
		return nil, corev1.ErrGenericError("Default user is nil")
	}

	err = s.createUserInDB(ctx, &defaultUser, &tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create default user in database: %w", err)
	}

	// Define a default vault
	defaultVault := entity.NewDefaultVault(&tenantID, &defaultUser.Uid)
	err = s.createVaultInDB(ctx, defaultVault, &defaultUser.Uid, &tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create default vault in database: %w", err)
	}

	// ##### AUTHORIZATION #####
	// create tenant in authorization service
	err = s.createTenantInAuthorization(ctx, result)
	if err != nil {
		// Rollback tenant creation in database if authorization fails
		originalErr := err
		if rollbackErr := s.authenticator.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.ID); rollbackErr != nil {
			return nil, fmt.Errorf("failed to rollback tenant creation in database: %w", rollbackErr)
		}
		return nil, fmt.Errorf("failed to create tenant in authorization service: %w", originalErr)
	}

	relation := fmt.Sprintf("%s", entity.TenantGroupOwner)
	err = s.AssociateGroupToTenant(ctx, &defaultGroup.ID, &tenantID, &defaultUser.Uid, &relation)
	if err != nil {
		return nil, fmt.Errorf("failed to associate group to tenant: %w", err)
	}

	err = s.AssociateUserToVault(ctx, &defaultVault.ID, &tenantID, &defaultUser.Uid)
	if err != nil {
		return nil, fmt.Errorf("failed to associate user to vault: %w", err)
	}

	features := make([]authorization.PlanFeature, 0)
	realations := []string{"owner", fmt.Sprintf("%s", entity.TenantGroupOwner)}

	err = s.authorizer.SetupTenant(ctx, tenantID, tenant.Owner.GetEmail(), features, realations)
	if err != nil {
		if err := s.authenticator.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.ID); err != nil {
			return nil, fmt.Errorf("failed to set tenant rollback: %w", err)
		}
		return nil, authorization.NewAuthorizationError("SetupTenant", "Failed to setup tenant in authorization service", err)
	}

	// ##### ATUALIZAR CUSTOM CLAIMS DO FIREBASE AUTH #####
	claim := entity.NewTenantCustomClaims(&tenantID, nil, &defaultUser)
	err = s.SetCustomClaims(ctx, &defaultUser, claim)
	if err != nil {
		return nil, fmt.Errorf("failed to set custom claims for user: %w", err)
	}

	return result, nil
}

func (s *tenantEventService) Get(ctx context.Context, filters entity.TenantFilter) (*entity.Tenant, error) {

	// CONTEXT
	if ctx.Err() != nil {
		return nil, errors.New(utils.ContextCancelled)
	}

	token, err := utils.GetAuthorizationFromContext(ctx) // Ensure user ID is retrieved from context
	if err != nil {
		return nil, fmt.Errorf(utils.ContextError, err.Error())
	}

	claims, err := s.authenticator.GetClaimsFromToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if claims == nil {
		return nil, errors.New("claims cannot be nil")
	}

	tenantID, ok := claims.CustomClaims["tenant_id"].(string)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in claims")
	}

	response, err := s.repo.List(ctx, []entity.TenantFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if len(response) == 0 {
		return nil, corev1.ErrGenericError("Tenant not found")
	}

	// tenant_memberships
	var responseTenant entity.Tenant
	for _, tenant := range response {
		if tenant.ID == tenantID {
			responseTenant = tenant
			break
		}
	}

	return &responseTenant, nil
}

func (s *tenantEventService) List(ctx context.Context, filters []entity.TenantFilter) ([]entity.Tenant, error) {
	return nil, nil
}

func (s *tenantEventService) Update(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	return nil, nil
}

func (s *tenantEventService) Delete(ctx context.Context, tenantID string) error {
	return nil
}
