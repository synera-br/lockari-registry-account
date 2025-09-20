package entity_registry_request

import (
	"context"
	"errors"
)

type ServiceRegistryRequest interface {
	RegistryNewAccount(ctx context.Context, user *RegistryRequest) (*RegistryResponse, error)
}

type UserFilter struct {
	Email   *string
	ID      *string
	UUID    *string
	Tenants *[]string
}

func (f *UserFilter) Validate() error {
	hasFilter := false
	if f == nil {
		return nil
	}

	if f.Email != nil && *f.Email != "" {
		hasFilter = true
	}

	if f.ID != nil && *f.ID != "" {
		hasFilter = true
	}

	if f.UUID != nil && *f.UUID != "" {
		hasFilter = true
	}

	if f.Tenants != nil && len(*f.Tenants) > 0 {
		hasFilter = true
	}

	if !hasFilter {
		return errors.New("no filter provided")
	}

	return nil
}
