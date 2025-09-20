package entity_user

import (
	"context"
	"errors"
)

type UserRequestData interface {
	Create(ctx context.Context, user *User) (*UserResponse, error)
	Get(ctx context.Context, filter *UserFilter) ([]*UserResponse, error)
}

type UserResponseData interface {
	Create(ctx context.Context, user *User) (*UserResponse, error)
	Get(ctx context.Context, filter *UserFilter) ([]*UserResponse, error)
}

type UserFilter struct {
	Email   *string   `bson:"email,omitempty" json:"email"`
	ID      *string   `bson:"_id,omitempty" json:"_id"`
	UUID    *string   `bson:"uuid,omitempty" json:"uuid"`
	Tenants *[]string `bson:"tenants,omitempty" json:"tenants"`
}

func (f *UserFilter) Validate() error {
	hasFilter := false
	if f == nil {
		return errors.New("no filter provided")
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
