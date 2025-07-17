package service

import (
	"context"
	"fmt"
	"sync"

	dto "github.com/synera-br/lockari-backend-app/internal/core/dto/role"
	entity_audit "github.com/synera-br/lockari-backend-app/internal/core/entity/audit"
	profile "github.com/synera-br/lockari-backend-app/internal/core/entity/profile"
	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/role"
	tenant "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	"github.com/synera-br/lockari-backend-app/pkg/authenticator"
	"github.com/synera-br/lockari-backend-app/pkg/authorization"
	corev1 "github.com/synera-br/lockari-backend-app/pkg/core/v1"
	"github.com/synera-br/lockari-backend-app/pkg/tokengen"
	"github.com/synera-br/lockari-backend-app/pkg/utils"
)

type userRoleService struct {
	repo    entity.UserRoleRepository
	tenant  tenant.TenantService
	profile profile.UserProfileService

	authenticator authenticator.Authenticator
	tokenJWT      tokengen.TokenGenerator
	authorizer    authorization.LockariAuthorizationService
	audit         entity_audit.AuditSystemEventService
	mu            sync.Mutex
	// wg            sync.WaitGroup
}

func InitializeUserRoleService(repo entity.UserRoleRepository, tenant tenant.TenantService, profile profile.UserProfileService, auth authenticator.Authenticator, tokenJWT tokengen.TokenGenerator, authz authorization.LockariAuthorizationService, audit entity_audit.AuditSystemEventService) (entity.UserRoleService, error) {
	if repo == nil {
		return nil, corev1.ErrRepositoryNotFound("UserRoleRepository")
	}

	if tenant == nil {
		return nil, corev1.ErrRepositoryNotFound("TenantService")
	}

	// if profile == nil {
	// 	return nil, corev1.ErrRepositoryNotFound("UserProfileService")
	// }

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

	return &userRoleService{
		repo:          repo,
		tenant:        tenant,
		profile:       profile,
		authenticator: auth,
		tokenJWT:      tokenJWT,
		authorizer:    authz,
		audit:         audit,
	}, nil
}

func (s *userRoleService) Create(ctx context.Context, userRole *entity.UserRole) (*entity.UserRole, error) {

	userClaim, err := s.validateToken(ctx)
	if err != nil {
		return nil, err
	}

	if userRole == nil {
		return nil, corev1.ErrGenericError("user role cannot be nil")
	}

	if err := userRole.IsValid(); err != nil {
		return nil, err
	}

	// Check if user exists in tenant
	if ok, err := s.UserExistsInTenant(ctx, &userClaim.UID, &userClaim.TenantID); err != nil {
		return nil, fmt.Errorf("error checking user existence in tenant: %w", err)
	} else if !ok {
		return nil, corev1.ErrGenericError("user does not exist in tenant")
	}

	// Check permission on authorization service
	if ok, err := s.authorizer.CanAssignRoleFromTenant(ctx, &userClaim.UID, &userClaim.TenantID, authorization.TenantRoleManager); err != nil {
		return nil, fmt.Errorf("error checking tenant permission: %w", err)
	} else if !ok {
		return nil, corev1.ErrPermissionDenied("user does not have permission to create user role")
	}

	// Preciso pegar o usuário que está fazendo a alteração
	// Pegar o tenant do usuário
	// Pegar a permissão do usuário referente ao tenant
	// Verificar se o usuário tem permissão para criar o user role
	// A permissão deve ser validada no AuthorizationService (OpenFGA)
	// Se não tiver, retornar erro de permissão

	// FALTA:
	// Validar se o usuário alvo existe no tenant
	// Validar se a role não existe já
	// Registrar evento de auditoria
	// Criar tupla no OpenFGA após criação no banco

	s.mu.Lock()
	defer s.mu.Unlock()

	createdRole, err := s.repo.Create(ctx, &userClaim.TenantID, userRole)
	if err != nil {
		return nil, err
	}

	return createdRole, nil
}

func (s *userRoleService) GetByID(ctx context.Context, id *string) (*entity.UserRole, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if id == nil {
		return nil, corev1.ErrGenericError("user role ID cannot be nil")
	}

	return nil, nil
}

func (s *userRoleService) Update(ctx context.Context, tenantID, id *string, userRole *entity.UserRole) (*entity.UserRole, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if tenantID == nil {
		return nil, corev1.ErrGenericError("tenant ID cannot be nil")
	}

	if id == nil {
		return nil, corev1.ErrGenericError("user role ID cannot be nil")
	}

	if userRole == nil {
		return nil, corev1.ErrGenericError("user role cannot be nil")
	}

	if err := userRole.IsValid(); err != nil {
		return nil, err
	}

	updatedRole, err := s.repo.Update(ctx, tenantID, id, userRole)
	if err != nil {
		return nil, err
	}

	return updatedRole, nil
}

func (s *userRoleService) Delete(ctx context.Context, tenantID, id *string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if id == nil {
		return corev1.ErrGenericError("user role ID cannot be nil")
	}

	if tenantID == nil {
		return corev1.ErrGenericError("tenant ID cannot be nil")
	}

	return s.repo.Delete(ctx, tenantID, id)
}

func (s *userRoleService) ListByUser(ctx context.Context, userID *string) ([]entity.UserRole, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if userID == nil {
		return nil, corev1.ErrGenericError("user ID cannot be nil")
	}

	return nil, nil
}

func (s *userRoleService) Get(ctx context.Context) ([]entity.UserRole, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return nil, nil
}

func (s *userRoleService) GetByFilter(ctx context.Context, filter *dto.RoleFilter) (*entity.UserRoleResponse, error) {
	userClaim, err := s.validateToken(ctx)
	if err != nil {
		return nil, err
	}

	if filter == nil {
		return nil, corev1.ErrGenericError("filter cannot be nil")
	}

	if err := filter.IsValid(); err != nil {
		return nil, err
	}

	// Check if user exists in tenant
	if ok, err := s.UserExistsInTenant(ctx, &userClaim.UID, &userClaim.TenantID); err != nil {
		return nil, fmt.Errorf("error checking user existence in tenant: %w", err)
	} else if !ok {
		return nil, corev1.ErrGenericError("user does not exist in tenant")
	}

	// Check permission on authorization service
	if ok, err := s.authorizer.CanAssignRoleFromTenant(ctx, &userClaim.UID, &userClaim.TenantID, authorization.TenantRoleManager); err != nil {
		return nil, fmt.Errorf("error checking tenant permission: %w", err)
	} else if !ok {
		return nil, corev1.ErrPermissionDenied("user does not have permission to create user role")
	}

	roles, err := s.repo.GetByFilter(ctx, &userClaim.TenantID, filter.ToMap())
	if err != nil {
		return nil, err
	}

	if roles == nil {
		return nil, corev1.ErrGenericError("no user roles found")
	}

	if roles.UserRole == nil && len(roles.UserRoles) == 0 {
		return nil, corev1.ErrGenericError("no user roles found")
	}

	return roles, nil
}

func (s *userRoleService) UserExistsInTenant(ctx context.Context, userID, tenantID *string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if userID == nil || *userID == "" {
		return false, corev1.ErrGenericError("user ID cannot be nil or empty")
	}

	if tenantID == nil || *tenantID == "" {
		return false, corev1.ErrGenericError("tenant ID cannot be nil or empty")
	}
	// profile, err := s.profile.GetProfile(ctx, &profile_dto.ProfileFilter{
	// 	TenantID: tenantID,
	// 	Uid:      userID,
	// })

	// if err != nil {
	// 	return false, fmt.Errorf("error getting user profile: %w", err)
	// }

	// if profile == nil || profile.Profile == nil {
	// 	return false, corev1.ErrGenericError("user profile not found")
	// }

	// if !profile.Profile.IsActive {
	// 	return false, corev1.ErrGenericError("user profile is not active")
	// }

	return tenantID != nil, nil
}

func (s *userRoleService) validateToken(ctx context.Context) (authenticator.UserCustomClaims, error) {
	if ctx == nil {
		return nil, corev1.ErrGenericError("context cannot be nil")
	}
	if ctx.Err() != nil {
		return nil, corev1.ErrGenericError("context is cancelled")
	}

	// check token firebase

	token, err := utils.GetAuthorizationFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf(corev1.ContextError, err.Error())
	}

	userClaim, err := s.authenticator.GetUserClaim(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user claim from context: %s", err.Error())
	}

	if userClaim == nil || userClaim.UID == "" {
		return nil, corev1.ErrGenericError("user claim is nil")
	}

	if userClaim.TenantID == "" {
		return nil, corev1.ErrGenericError("user claim tenant ID is empty")
	}

	return userClaim, nil
}
