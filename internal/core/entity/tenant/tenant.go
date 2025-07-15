package entity

import (
	"errors"
	"strings"
	"time"

	corev1 "github.com/synera-br/lockari-backend-app/pkg/core/v1"
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
	Name string      `json:"name,omitempty"` // Name of the tenant
	Plan corev1.Plan `json:"plan,omitempty"` // Subscription plan of the tenant
}

type Tenant struct {
	ID         string           `json:"id,omitempty"`        // Unique identifier for the tenant
	EventType  corev1.EventType `json:"eventType,omitempty"` // Event type, e.g., SIGNUP_SUCCESS
	Owner      Owner            `json:"owner" binding:"required"`
	ClientInfo Client           `json:"clientInfo" binding:"required"`
	Timestamp  time.Time        `json:"timestamp" binding:"required"`
	CreatedAt  time.Time        `json:"createdAt,omitempty"` // Creation timestamp
	UpdatedAt  time.Time        `json:"updatedAt,omitempty"` // Last update timestamp
	Tenant     TenantInfo       `json:"tenant,omitempty"`    // Information about the tenant
}

func (t *Tenant) IsValid() error {
	if t == nil {
		return errors.New(ErrInvalidTenant)
	}

	if t.Owner.IsValid() != nil {
		return t.Owner.IsValid()
	}
	if t.ClientInfo.IsValid() != nil {
		return t.ClientInfo.IsValid()
	}
	if t.Timestamp.IsZero() {
		return errors.New("invalid tenant: timestamp is required")
	}
	if t.Tenant.IsValid() != nil {
		return t.Tenant.IsValid()
	}
	if t.EventType.IsValid() != nil {
		return errors.New("invalid tenant: eventType is required")
	}

	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now()
	}

	if t.Tenant.Name == "" {
		t.Tenant.SetTenantName(t.Owner.GetUsername())
	}

	if t.ID != "" {
		if len(t.ID) != 36 {
			return errors.New(ErrInvalidTenantID)
		}
	}

	return nil
}

func (t *Tenant) GetTenant() Tenant {
	if t == nil {
		return Tenant{}
	}
	return *t
}

func (t *Tenant) GetPlan() string {
	if t == nil {
		return ""
	}
	return t.Tenant.Plan.String()
}

func (t *Tenant) GetOwner() Owner {
	if t == nil {
		return Owner{}
	}
	return t.Owner
}

func (t *Tenant) GetID() (string, bool) {
	isValid := false
	if t.ID != "" {
		isValid = true
	}
	return t.ID, isValid
}

func (t *Tenant) SetTenantID(tenantID *string) error {
	if t == nil {
		return errors.New(ErrInvalidTenantInfo)
	}

	if tenantID == nil {
		return errors.New(ErrTenantRequired)
	}

	if *tenantID == "" || len(*tenantID) != 36 {
		return errors.New(ErrInvalidTenantID)
	}

	if t.ID != "" {
		return errors.New(ErrTenantAlreadySet)
	}

	t.ID = *tenantID

	if t.ID == "" {
		return errors.New(ErrInvalidTenantID)
	}

	return nil
}

func (ti *TenantInfo) IsValid() error {
	if ti == nil {
		return errors.New(ErrInvalidTenantInfo)
	}

	if ti.Plan == "" {
		return errors.New(ErrInvalidTenantInfoPlan)
	}

	ti.toLower()
	return nil
}

func (t *TenantInfo) SetTenantName(name string) error {
	if t == nil {
		return errors.New(ErrInvalidTenant)
	}
	t.Name = name

	return nil
}

func (t *TenantInfo) toLower() {
	if t == nil {
		return
	}
	t.Name = strings.ToLower(t.Name)
}
