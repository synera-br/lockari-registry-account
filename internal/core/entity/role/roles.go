package entity

import (
	"errors"
	"time"
)

type Role struct {
	ObjectID    string    `json:"object,omitempty"`      // ID of the role object
	ObjectName  string    `json:"object_name,omitempty"` // Name of the role object
	ObjectType  string    `json:"object_type,omitempty"` // Type of the object (e.g., "tenant", "group")
	Permissions []string  `json:"permissions,omitempty"` // List of permissions associated with the role
	Description string    `json:"description,omitempty"` // Description of the role
	CreatedAt   time.Time `json:"created_at,omitempty"`  // Timestamp when the role was created
	UpdatedAt   time.Time `json:"updated_at,omitempty"`  // Timestamp when the role was last updated
	CreatedBy   string    `json:"created_by,omitempty"`  // User who created the role
	UpdatedBy   string    `json:"updated_by,omitempty"`  // User who last updated the role
}

// NewRole creates a new role with the specified parameters.
// It returns an error if any of the required fields are empty.
func NewRole(objectID, objectName, objectType string, permissions []string, createdBy string) (*Role, error) {
	if objectID == "" || objectName == "" || objectType == "" {
		return nil, errors.New("object ID, name, and type cannot be empty")
	}
	if len(permissions) == 0 {
		return nil, errors.New("permissions cannot be empty")
	}

	role := &Role{
		ObjectID:    objectID,
		ObjectName:  objectName,
		ObjectType:  objectType,
		Permissions: permissions,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	role.CreatedAt = time.Now()
	role.UpdatedAt = role.CreatedAt

	return role, nil
}

func (r *Role) IsValid() error {
	if r.ObjectID == "" {
		return errors.New("role object ID cannot be empty")
	}
	if r.ObjectName == "" {
		return errors.New("role object name cannot be empty")
	}
	if r.ObjectType == "" {
		return errors.New("role object type cannot be empty")
	}
	if len(r.Permissions) == 0 {
		return errors.New("role must have at least one permission")
	}
	if r.UpdatedAt.IsZero() {
		return errors.New("role updated at timestamp cannot be empty")
	}
	return nil
}

func (r *Role) GetObjectID() string {
	return r.ObjectID
}

func (r *Role) GetObjectName() string {
	return r.ObjectName
}

func (r *Role) GetObjectType() string {
	return r.ObjectType
}

func (r *Role) GetPermissions() []string {
	return r.Permissions
}

func (r *Role) SetPermissions(permissions []string) error {
	if len(permissions) == 0 {
		return errors.New("permissions cannot be empty")
	}
	r.Permissions = permissions
	return nil
}

func (r *Role) SetUpdatedBy(user *string) error {
	if user == nil || *user == "" {
		return errors.New("updated by user cannot be empty")
	}
	r.UpdatedBy = *user
	return nil
}
