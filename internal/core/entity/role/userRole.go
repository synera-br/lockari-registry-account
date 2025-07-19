package entity

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	dto "lockari-api-application/internal/core/dto/role"
)

type UserRoleRepository interface {
	Create(ctx context.Context, tenantid *string, userRole *UserRole) (*UserRole, error)
	Update(ctx context.Context, tenantID, id *string, userRole *UserRole) (*UserRole, error)
	Delete(ctx context.Context, tenantID, id *string) error
	Get(ctx context.Context) (*UserRoleResponse, error)
	GetByFilter(ctx context.Context, tenantID *string, filter map[string]interface{}) (*UserRoleResponse, error)
}

type UserRoleService interface {
	Create(ctx context.Context, userRole *UserRole) (*UserRole, error)
	GetByID(ctx context.Context, id *string) (*UserRole, error)
	Update(ctx context.Context, tenantID, id *string, userRole *UserRole) (*UserRole, error)
	Delete(ctx context.Context, tenantID, id *string) error
	ListByUser(ctx context.Context, userID *string) ([]UserRole, error)
	Get(ctx context.Context) ([]UserRole, error)
	GetByFilter(ctx context.Context, filter *dto.RoleFilter) (*UserRoleResponse, error)
}

type UserRoleResponse struct {
	UserRole  *UserRole  `json:"user_role,omitempty"`
	UserRoles []UserRole `json:"user_roles,omitempty"`
}

type UserRole struct {
	ID        string    `json:"id,omitempty"`         // Unique identifier for the user role
	User      string    `json:"user,omitempty"`       // Unique identifier for the user
	Roles     []Role    `json:"role,omitempty"`       // Role assigned to the user
	CreatedAt time.Time `json:"created_at,omitempty"` // Timestamp when the role was created
	UpdatedAt time.Time `json:"updated_at,omitempty"` // Timestamp when the role was last updated
}

func (r *UserRole) IsValid() error {
	if r.User == "" {
		return errors.New("user ID cannot be empty")
	}
	if len(r.Roles) == 0 {
		return errors.New("user must have at least one role")
	}

	if r.UpdatedAt.IsZero() {
		return errors.New("user role updated at timestamp cannot be empty")
	}
	for _, role := range r.Roles {
		if err := role.IsValid(); err != nil {
			return err
		}
	}
	return nil
}

func NewUserRole(user string, roles []Role) (*UserRole, error) {
	if user == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if len(roles) == 0 {
		return nil, errors.New("user must have at least one role")
	}

	id, _ := uuid.NewV7()

	r := &UserRole{
		ID:        id.String(),
		User:      user,
		Roles:     roles,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	r.appendRoles(roles)

	if err := r.IsValid(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *UserRole) GetID() string {
	return r.ID
}

func (r *UserRole) GetObjectID() string {
	return r.ID
}

func (r *UserRole) GetUser() string {
	return r.User
}

func (r *UserRole) GetRoles() []Role {
	return r.Roles
}

func (r *UserRole) SetUser(user string) error {
	if user == "" {
		return errors.New("user ID cannot be empty")
	}

	if r.User != "" && r.User != user {
		return errors.New("user ID cannot be changed once set")
	}

	r.User = user
	r.UpdatedAt = time.Now()
	return nil
}

func (r *UserRole) SetRoles(roles []Role) error {
	if len(roles) == 0 {
		return errors.New("user must have at least one role")
	}

	r.appendRoles(roles)

	return nil
}

func (r *UserRole) appendRoles(roles []Role) error {
	if len(roles) == 0 {
		return errors.New("user must have at least one role")
	}

	for _, role := range roles {
		if err := role.IsValid(); err != nil {
			return err
		}
		if role.UpdatedBy == "" {
			return errors.New("role updated by cannot be empty")
		}
	}
	r.Roles = append(r.Roles, roles...)
	r.UpdatedAt = time.Now()

	return nil
}
