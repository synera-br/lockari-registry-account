package tenant

import (
	"context"
	"fmt"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	corev1 "github.com/synera-br/lockari-backend-app/pkg/core/v1"
)

func (s *tenantEventService) createTenantInDB(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	// Check if tenant already exists
	existingTenants, err := s.repo.List(ctx, nil)
	if err != nil {
		if err.Error() == corev1.TenantAlreadyExists {
			return nil, corev1.ErrGenericError(fmt.Sprintf("Tenant with name %s already exists", tenant.Tenant.Name))
		}
		return nil, fmt.Errorf("failed to check existing tenant: %w", err)
	}

	if len(existingTenants) > 0 {
		for _, existingTenant := range existingTenants {
			if existingTenant.Owner.Email == tenant.Owner.Email {
				return nil, corev1.ErrGenericError(fmt.Sprintf("Tenant with email %s already exists", tenant.Owner.Email))
			}
			if existingTenant.ID == tenant.ID {
				return nil, corev1.ErrGenericError(fmt.Sprintf("Tenant with ID %s already exists", tenant.ID))
			}
			if tenant.Tenant.Name != "" && (existingTenant.Tenant.Name == tenant.Tenant.Name) {
				return nil, corev1.ErrGenericError(fmt.Sprintf("Tenant with name %s already exists", tenant.Tenant.Name))
			}
		}
	}
	result, err := s.repo.Create(ctx, tenant)
	if err != nil {
		// Salvar o erro original antes de tentar rollback
		// Sempre retorna o erro original, não o erro do rollback
		return nil, fmt.Errorf("failed to create a new tenant in database: %w", err)
	}
	if result == nil {
		return nil, corev1.ErrGenericError("Failed to create tenant event")
	}
	return result, nil
}

func (s *tenantEventService) createUserInDB(ctx context.Context, user *entity.Owner, tenantID *string) error {

	if user == nil {
		return corev1.ErrGenericError("user cannot be nil")
	}

	if user.Uid == "" || user.Email == "" {
		return corev1.ErrGenericError("Invalid User")
	}

	if tenantID == nil || *tenantID == "" {
		return corev1.ErrGenericError("Tenant ID is required")
	}

	err := s.repo.CreateUser(ctx, user, tenantID)
	if err != nil {
		return fmt.Errorf("failed to create user in database: %w", err)
	}

	return nil
}

func (s *tenantEventService) createGroupInDB(ctx context.Context, group *entity.UserGroup, tenantID *string) error {

	if group == nil {
		return corev1.ErrGenericError("UserGroup cannot be nil")
	}

	if group.ID == "" || group.Name == "" {
		return corev1.ErrGenericError("Invalid UserGroup")
	}

	if tenantID == nil || *tenantID == "" {
		return corev1.ErrGenericError("Tenant ID is required")
	}

	err := s.repo.CreateGroup(ctx, group, tenantID)
	if err != nil {
		return fmt.Errorf("failed to create group in database: %w", err)
	}

	return nil
}

func (s *tenantEventService) createVaultInDB(ctx context.Context, vault *entity.Vault, userID, tenantID *string) error {
	if vault == nil {
		return corev1.ErrGenericError("Vault cannot be nil")
	}

	if vault.ID == "" || vault.Name == "" {
		return corev1.ErrGenericError("Invalid Vault")
	}

	if userID == nil || *userID == "" {
		return corev1.ErrGenericError("User ID is required")
	}

	if tenantID == nil || *tenantID == "" {
		return corev1.ErrGenericError("Tenant ID is required")
	}

	err := s.repo.CreateVault(ctx, vault, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to create vault in database: %w", err)
	}

	return nil
}
