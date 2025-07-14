package entity

import (
	"errors"
	"time"

	eventtype "github.com/synera-br/lockari-backend-app/pkg/event_type"
)

const (
	ErrInvalidTenant         = "invalid tenant: tenant cannot be empty"
	ErrInvalidTenantID       = "invalid tenant: tenant must be a valid UUID"
	ErrTenantAlreadySet      = "invalid tenant: tenant is already set"
	ErrTenantRequired        = "invalid tenant: tenant is required"
	ErrInvalidTenantInfo     = "invalid tenant info: tenant info cannot be nil"
	ErrInvalidTenantInfoName = "invalid tenant info: name is required"
	ErrInvalidTenantInfoID   = "invalid tenant info: tenant_id is required"
	ErrInvalidTenantInfoPlan = "invalid tenant info: plan is required"
)

type TenantInfo struct {
	Name     string `json:"name,omitempty"`      // Name of the tenant
	TenantID string `json:"tenant_id,omitempty"` // Unique identifier for the tenant
	Plan     string `json:"plan,omitempty"`      // Subscription plan of the tenant
}

type Tenant struct {
	ID         string              `json:"id,omitempty"`        // Unique identifier for the tenant
	EventType  eventtype.EventType `json:"eventType,omitempty"` // Event type, e.g., SIGNUP_SUCCESS
	Owner      Owner               `json:"owner" binding:"required"`
	ClientInfo Client              `json:"clientInfo" binding:"required"`
	Timestamp  time.Time           `json:"timestamp" binding:"required"`
	CreatedAt  time.Time           `json:"createdAt,omitempty"` // Creation timestamp
	UpdatedAt  time.Time           `json:"updatedAt,omitempty"` // Last update timestamp
	Tenant     TenantInfo          `json:"tenant,omitempty"`    // Information about the tenant
}

func (ti *TenantInfo) IsValid() error {
	if ti == nil {
		return errors.New(ErrInvalidTenantInfo)
	}

	if ti.Name == "" {
		return errors.New(ErrInvalidTenantInfoName)
	}

	if ti.TenantID == "" {
		return errors.New(ErrInvalidTenantInfoID)
	}

	if ti.Plan == "" {
		return errors.New(ErrInvalidTenantInfoPlan)
	}

	return nil
}

func (ti *TenantInfo) SetTenantID(tenantID *string) error {
	if ti == nil {
		return errors.New(ErrInvalidTenantInfo)
	}

	if len(*tenantID) != 36 {
		return errors.New(ErrInvalidTenantID)
	}

	if ti.TenantID != "" {
		return errors.New(ErrTenantAlreadySet)
	}

	if ti.TenantID == *tenantID {
		return errors.New("invalid signup: tenant is already set to the same value")
	}

	ti.TenantID = *tenantID
	if ti.TenantID == "" {
		return errors.New(ErrTenantRequired)
	}

	return nil
}
