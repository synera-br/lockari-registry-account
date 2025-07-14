package tenant

import (
	"context"
	"fmt"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	"github.com/synera-br/lockari-backend-app/pkg/authenticator"
	"github.com/synera-br/lockari-backend-app/pkg/authorization"
	corev1 "github.com/synera-br/lockari-backend-app/pkg/core/v1"
	"github.com/synera-br/lockari-backend-app/pkg/tokengen"
	"github.com/synera-br/lockari-backend-app/pkg/utils"
)

type tenantEventService struct {
	repo     entity.TenantRepository
	auth     authenticator.Authenticator
	tokenJWT tokengen.TokenGenerator
	authz    authorization.LockariAuthorizationService
}

func InitializeTenantEventService(repo entity.TenantRepository, auth authenticator.Authenticator, tokenJWT tokengen.TokenGenerator, authz authorization.LockariAuthorizationService) (entity.TenantService, error) {

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

	return &tenantEventService{
		repo:     repo,
		auth:     auth,
		tokenJWT: tokenJWT,
		authz:    authz,
	}, nil
}

func (s *tenantEventService) Create(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	// Check tenant is valid
	if tenant == nil {
		return nil, corev1.ErrGenericError("Tenant is required")
	}
	if err := tenant.IsValid(); err != nil {
		if err.Error() != entity.ErrInvalidTenantInfoID {
			return nil, corev1.ErrGenericError("Invalid tenant: " + err.Error())
		}
	}

	// Check context
	if ctx.Err() != nil {
		return nil, fmt.Errorf(corev1.ContextCancelled, ctx.Err())
	}

	// check token firebase
	token := utils.GetTokenFromContext(ctx)

	// check tokenJWT
	_, err := s.tokenJWT.Validate(token)
	if err != nil {
		return nil, fmt.Errorf(corev1.GenericError, err)
	}

	// Check if tenant already exists
	existingTenants, err := s.repo.List(ctx, nil)

	if err != nil {
		if err.Error() == corev1.TenantAlreadyExists {
			return nil, corev1.ErrGenericError(fmt.Sprintf("Tenant with name %s already exists", tenant.Tenant.Name))
		}
		return nil, fmt.Errorf("failed to check existing tenant: %w", err)
	}

	if existingTenants != nil && existingTenants[0].Tenant.TenantID != "" {
		return nil, corev1.ErrGenericError("Tenant already exists with the provided details")
	}

	// Generate tenant ID if not provided
	tenantID := utils.GenerateTenant()
	tenant.Tenant.SetTenantID(&tenantID)
	if tenant.Tenant.TenantID == "" {
		return nil, corev1.ErrGenericError("Failed to generate tenant ID")
	}

	features := make([]authorization.PlanFeature, 0)
	for _, feature := range tenant.Tenant.Plan.GetFeatures() {
		features = append(features, authorization.PlanFeature(feature))
	}

	err = s.authz.SetupTenant(ctx, tenant.Tenant.TenantID, tenant.Owner.GetEmail(), features)
	if err != nil {
		if err := s.auth.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.Tenant.TenantID); err != nil {
			return nil, fmt.Errorf("failed to set tenant rollback: %w", err)
		}
		return nil, authorization.NewAuthorizationError("SetupTenant", "Failed to setup tenant in authorization service", err)
	}

	// Set tenant's owner
	err = s.authz.AddUserToTenant(ctx, tenant.Owner.GetEmail(), tenant.Tenant.TenantID, authorization.TenantRoleOwner)
	if err != nil {
		if err := s.auth.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.Tenant.TenantID); err != nil {
			return nil, fmt.Errorf("failed to set tenant rollback: %w", err)
		}
		return nil, authorization.NewAuthorizationError("AddUserToTenant", "Failed to add user to tenant in authorization service", err)
	}

	result, err := s.repo.Create(ctx, tenant)
	if err != nil {
		// Salvar o erro original antes de tentar rollback
		originalErr := err
		if err := s.auth.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.Tenant.TenantID); err != nil {
			return nil, fmt.Errorf("failed to set tenant rollback: %w", err)
		}
		// Tentar rollback do tenant
		if err := s.auth.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.Tenant.TenantID); err != nil {
			return nil, fmt.Errorf("failed to set tenant rollback: %w", err)
		}
		// Sempre retorna o erro original, não o erro do rollback
		return nil, originalErr
	}
	if result == nil {
		return nil, corev1.ErrGenericError("Failed to create tenant event")
	}

	return result, nil
}

func (s *tenantEventService) Get(ctx context.Context, filters entity.TenantFilter) (*entity.Tenant, error) {
	return nil, nil
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
