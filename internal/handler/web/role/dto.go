package webhandler

import (
	"errors"
	"time"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
)

type TenantDetailsResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	OwnerEmail string    `json:"ownerEmail"`
	CreatedAt  time.Time `json:"createdAt"` // Using time.Time for better handling
	Plan       PlanInfo  `json:"plan"`
}

// PlanInfo contains details about the subscription plan.
type PlanInfo struct {
	Name     string   `json:"name"`
	Features []string `json:"features"`
}

func NewTenantDetailsResponse(tenant *entity.Tenant) (*TenantDetailsResponse, error) {

	if tenant == nil {
		return nil, errors.New("tenant is nil")
	}

	if tenant.ID == "" {
		return nil, errors.New("tenant.Tenant is nil")
	}

	features := []string{"No features available"}
	if len(tenant.Tenant.Plan.GetFeatures()) != 0 {
		features = []string{}
		for _, feature := range tenant.Tenant.Plan.GetFeatures() {
			features = append(features, feature.String())
		}
	}

	return &TenantDetailsResponse{
		ID:         tenant.ID,
		Name:       tenant.Tenant.Name,
		OwnerEmail: tenant.Owner.GetEmail(),
		CreatedAt:  tenant.CreatedAt,
		Plan: PlanInfo{
			Name:     string(tenant.Tenant.Plan),
			Features: features,
		},
	}, nil
}
