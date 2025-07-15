package tenant

import (
	"context"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	corev1 "github.com/synera-br/lockari-backend-app/pkg/core/v1"
)

func (s *tenantEventService) SetCustomClaims(ctx context.Context, owner *entity.Owner, claims *entity.TenantCustomClaims) error {

	var claimsToMap map[string]interface{}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if owner == nil {
		return corev1.ErrGenericError("owner cannot be nil")
	}
	if claims != nil {
		claimsToMap = claims.ToMap()
	} else {
		claimsToMap = make(map[string]interface{})
	}
	if err := s.authenticator.SetCustomClaims(ctx, owner.Uid, claimsToMap); err != nil {
		return err
	}
	return nil
}
