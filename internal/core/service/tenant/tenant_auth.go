package tenant

import (
	"context"
	"fmt"

	entity "lockari-api-application/internal/core/entity/tenant"
	corev1 "lockari-api-application/pkg/core/v1"
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

	getCustomClaims := claims.ToMap()
	if getCustomClaims == nil {
		return corev1.ErrGenericError("failed to convert claims to map")
	}

	var customClaims map[string]interface{}

	// Verificar se já existe a estrutura custom_claims
	if getCustomClaims["custom_claims"] != nil {
		// Já tem a estrutura correta, usar como está
		customClaims = getCustomClaims
		fmt.Println("Using existing custom_claims structure")
	} else {
		// Precisa criar a estrutura custom_claims e mover tudo para dentro
		customClaims = make(map[string]interface{})
		customClaims["custom_claims"] = getCustomClaims // TUDO vai para dentro de custom_claims
		fmt.Println("Created new custom_claims structure")
	}

	fmt.Printf("Final customClaims structure: %+v\n", customClaims)

	if err := s.authenticator.SetCustomClaims(ctx, owner.Uid, customClaims); err != nil {
		return err
	}

	return nil
}
