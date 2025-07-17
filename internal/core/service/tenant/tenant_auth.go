package tenant

import (
	"context"
	"fmt"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	corev1 "github.com/synera-br/lockari-backend-app/pkg/core/v1"
)

func (s *tenantEventService) SetCustomClaims(ctx context.Context, owner *entity.Owner, claims *entity.TenantCustomClaims) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}
	if owner == nil {
		return corev1.ErrGenericError("owner cannot be nil")
	}
	if claims == nil {
		return corev1.ErrGenericError("claims cannot be nil")
	}

	if claims.CustomClaims.TenantID == "" {
		return corev1.ErrGenericError("tenantID cannot be empty")
	}

	fmt.Println("\nSetting custom claims for user:", claims.ToMap())
	fmt.Println("\nSettings of owner:", owner)
	if err := s.authenticator.SetCustomClaims(ctx, owner.Uid, claims.ToMap()); err != nil {
		return err
	}

	return nil
}
