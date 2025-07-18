package authorization

import "fmt"

// TenantOperation representa uma operação de tupla no contexto de um tenant
// Ela inclui informações sobre o usuário, a relação, o objeto e a ação a ser realizada
type TenantOperation struct {
	UserType   TenantObject
	UserID     string               // O ID específico do User (ex: Firebase UID, ID do grupo, ID do tenant)
	Relation   TenantRolePermission // A relação (ex: "member", "owner", "tenant_owner_group", "parent")
	ObjectType string
	ObjectID   string // O ID específico do Object (ex: tenantID, vaultID, secretID)
}

type TenantObject string

const (
	TenantObjectUser  TenantObject = "user"
	TenantObjectGroup TenantObject = "group"
)

func (t TenantObject) IsValid() bool {
	validObjects := []TenantObject{
		TenantObjectUser, TenantObjectGroup,
	}

	for _, valid := range validObjects {
		if t == valid {
			return true
		}
	}
	return false
}
func (t TenantObject) String() (string, error) {
	if !t.IsValid() {
		return "", fmt.Errorf("invalid tenant object: %s", t)
	}
	return string(t), nil
}

func (t *TenantOperation) IsValid() error {
	if t == nil {
		return fmt.Errorf("tenant operation cannot be nil")
	}
	if !t.UserType.IsValid() {
		return fmt.Errorf("invalid user type: %s %s", t.UserType, "must be one of: owner, admin, member, guest, manager")
	}
	if t.UserID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if !t.Relation.IsValid() {
		return fmt.Errorf("invalid relation: %s", t.Relation)
	}
	if t.ObjectType == "" {
		return fmt.Errorf("object type cannot be empty")
	}

	if t.ObjectID == "" {
		return fmt.Errorf("object ID cannot be empty")
	}

	return nil
}

type TenantRolePermission string

const (
	TenantRolePermissionOwner        TenantRolePermission = "owner"
	TenantRolePermissionOwnerGroup   TenantRolePermission = "tenant_owner_group"
	TenantRolePermissionManagerGroup TenantRolePermission = "tenant_manager_group"
	TenantRolePermissionViewerGroup  TenantRolePermission = "tenant_viewer_group"
)

func (t *TenantRolePermission) IsValid() bool {
	validRoles := []TenantRolePermission{
		TenantRolePermissionOwner, TenantRolePermissionOwnerGroup,
		TenantRolePermissionManagerGroup, TenantRolePermissionViewerGroup,
	}

	for _, valid := range validRoles {
		if *t == valid {
			return true
		}
	}
	return false
}

func (t *TenantRolePermission) String() (string, error) {
	if !t.IsValid() {
		return "", fmt.Errorf("invalid tenant role permission: %s", *t)
	}
	return string(*t), nil
}
