package entity

import (
	"context"
	"errors"
	"strings"
	"time"

	dto "lockari-api-application/internal/core/dto/profile"
	entity_tenant "lockari-api-application/internal/core/entity/tenant"
)

type UserProfileRepository interface {
	CreateProfile(ctx context.Context, profile *Profile) (*Profile, error)
	UpdateProfile(ctx context.Context, id *string, profile *Profile) (*Profile, error)
	DeleteProfile(ctx context.Context, id *string) error
	GetProfile(ctx context.Context, filter map[string]interface{}) (*ProfileResponse, error)
}

type UserProfileService interface {
	CreateProfile(ctx context.Context, profile *Profile) (*Profile, error)
	UpdateProfile(ctx context.Context, id *string, profile *Profile) (*Profile, error)
	DeleteProfile(ctx context.Context, id *string) error
	GetProfile(ctx context.Context, filter *dto.ProfileFilter) (*ProfileResponse, error)
}

type TenantMemberships struct {
	TenantID string `json:"tenant_id,omitempty"` // ID of the tenant the user belongs to
	Role     string `json:"role,omitempty"`      // Role of the user in the tenant
}

type ProfileResponse struct {
	Profile  *Profile  `json:"profile,omitempty"`
	Profiles []Profile `json:"profiles,omitempty"`
}

// Profile
// This event is triggered when a user performs an action that requires authentication, such as logging in or signing up.
type Profile struct {
	Uid               string                        `json:"uid" binding:"required"`       // Unique identifier for the user in Firebase Authentication
	Email             string                        `json:"email" binding:"required"`     // Email address of the user
	Username          string                        `json:"username,omitempty"`           // Optional: Username of the tenant
	TenantID          string                        `json:"tenant_id,omitempty"`          // Optional: Tenant ID associated with the user
	FullName          string                        `json:"full_name,omitempty"`          // Optional: Full name of the user
	PermissionLevel   entity_tenant.TenantGroupType `json:"permission_level,omitempty"`   // Optional: Permission level of the user in the tenant
	CreatedAt         time.Time                     `json:"created_at,omitempty"`         // Optional: Timestamp when the user was created
	GroupMemberships  []entity_tenant.GroupMember   `json:"group_memberships,omitempty"`  // Optional: List of group IDs the user belongs to
	TenantMemberships []TenantMemberships           `json:"tenant_memberships,omitempty"` // Optional: List of tenant IDs the user belongs to
	IsActive          bool                          `json:"is_active,omitempty"`          // Optional: Indicates if the user is active
}

func (u *Profile) SetDefaultUser(tenant *string, group *entity_tenant.GroupMember) error {
	if err := u.IsValid(); err != nil {
		return err
	}

	if u.Username == "" {
		u.Username = "default-username"
	}
	if u.FullName == "" {
		name := strings.Split(u.Email, "@")
		endName := strings.ReplaceAll(name[0], ".", " ")
		u.FullName = strings.TrimSpace(strings.ToLower(endName))
	}

	if u.PermissionLevel == "" {
		u.PermissionLevel = entity_tenant.TenantGroupOwner
	}

	if u.TenantID == "" && tenant != nil {
		u.TenantID = *tenant
	}

	if u.GroupMemberships == nil && group != nil {
		u.GroupMemberships = []entity_tenant.GroupMember{*group}
	}

	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}

	u.TenantMemberships = []TenantMemberships{
		{TenantID: "default-tenant-id", Role: "member"},
	}

	u.IsActive = true

	return nil
}

func (u *Profile) IsValid() error {
	if u == nil {
		return errors.New("invalid user: user event cannot be nil")
	}

	if u.Uid == "" {
		return errors.New("invalid user: uid is required")
	}

	if u.Email == "" {
		return errors.New("invalid user: email is required")
	}

	u.toLower()

	return nil
}

func (u *Profile) GetUid() string {
	if u == nil {
		return ""
	}
	return u.Uid
}

func (u *Profile) GetEmail() string {
	if u == nil {
		return ""
	}
	return u.Email
}

func (u *Profile) GetUsername() string {
	if u == nil {
		return ""
	}
	return u.Username
}

func (u *Profile) SetUsername(username string) {
	if u == nil {
		return
	}
	u.Username = username
}

func (u *Profile) GetUser() Profile {
	if u == nil {
		return Profile{}
	}
	return *u
}

func (u *Profile) toLower() {
	if u == nil {
		return
	}

	u.Email = strings.ToLower(u.Email)
	u.Username = strings.ToLower(u.Username)
}
