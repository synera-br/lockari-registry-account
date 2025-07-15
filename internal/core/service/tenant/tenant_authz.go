package tenant

import (
	"context"
	"fmt"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	"github.com/synera-br/lockari-backend-app/pkg/authorization"
)

func (s *tenantEventService) createTenantInAuthorization(ctx context.Context, tenant *entity.Tenant) error {

	features := make([]authorization.PlanFeature, 0)

	relations := []string{"owner", fmt.Sprintf("%s", entity.TenantGroupOwner)}

	err := s.authorizer.SetupTenant(ctx, tenant.ID, tenant.Owner.GetEmail(), features, relations)
	if err != nil {
		if err := s.authenticator.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.ID); err != nil {
			return fmt.Errorf("failed to set tenant rollback: %w", err)
		}
		return authorization.NewAuthorizationError("SetupTenant", "Failed to setup tenant in authorization service", err)
	}

	return nil
}

func (s *tenantEventService) AssociateGroupToTenant(ctx context.Context, groupID, tenantID, ownerID, relation *string) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if groupID == nil || tenantID == nil || ownerID == nil || relation == nil {
		return fmt.Errorf("all parameters are required")
	}

	if *relation == "" {
		return fmt.Errorf("relation is required")
	}

	if *groupID == "" {
		return fmt.Errorf("groupID is required")
	}

	if *tenantID == "" {
		return fmt.Errorf("tenantID is required")
	}

	if *ownerID == "" {
		return fmt.Errorf("ownerID is required")
	}

	err := s.authorizer.AssociateGroupToTenant(ctx, *groupID, *tenantID, *ownerID, *relation)
	if err != nil {
		return fmt.Errorf("failed to associate group to tenant: %w", err)
	}

	err = s.authorizer.AssociateUserToGroupAsMember(ctx, *ownerID, *groupID, *tenantID)
	if err != nil {
		return fmt.Errorf("failed to associate user to group as member: %w", err)
	}

	return nil
}

func (s *tenantEventService) AssociateUserToVault(ctx context.Context, vaultID, tenantID, userID *string) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if vaultID == nil || tenantID == nil || userID == nil {
		return fmt.Errorf("all parameters are required")
	}

	if *vaultID == "" {
		return fmt.Errorf("vaultID is required")
	}

	if *tenantID == "" {
		return fmt.Errorf("tenantID is required")
	}

	if *userID == "" {
		return fmt.Errorf("userID is required")
	}

	err := s.authorizer.SetupVault(ctx, *vaultID, *tenantID, *userID)
	if err != nil {
		return fmt.Errorf("failed to setup vault: %w", err)
	}

	return nil
}
