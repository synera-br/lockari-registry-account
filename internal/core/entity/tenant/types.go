package entity

import (
	"context"
	"errors"
	"time"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) (*Tenant, error)
	Get(ctx context.Context, filters TenantFilter) (*Tenant, error)
	List(ctx context.Context, filters []TenantFilter) ([]Tenant, error)
	Update(ctx context.Context, tenant *Tenant) (*Tenant, error)
	Delete(ctx context.Context, tenantID string) error
}

type TenantService interface {
	Create(ctx context.Context, tenant *Tenant) (*Tenant, error)
	Get(ctx context.Context, filters TenantFilter) (*Tenant, error)
	List(ctx context.Context, filters []TenantFilter) ([]Tenant, error)
	Update(ctx context.Context, tenant *Tenant) (*Tenant, error)
	Delete(ctx context.Context, tenantID string) error
}

type TenantFilter struct {
	TenantID  string    `json:"tenant_id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Plan      string    `json:"plan,omitempty"`
	Email     string    `json:"email,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
}

func (f *TenantFilter) IsValid() error {
	if f == nil {
		return errors.New("invalid tenant filter: filter cannot be nil")
	}

	if f.TenantID == "" && f.Name == "" && f.Plan == "" && f.Email == "" {
		return errors.New("invalid tenant filter: at least one field must be provided")
	}

	if !f.StartDate.IsZero() && !f.EndDate.IsZero() && f.StartDate.After(f.EndDate) {
		return errors.New("invalid tenant filter: start date cannot be after end date")
	}

	return nil
}
