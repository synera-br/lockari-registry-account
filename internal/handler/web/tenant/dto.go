package webhandler

import "time"

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
