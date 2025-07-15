package entity

type TenantCustomClaims struct {
	TenantID          string              `json:"tenant_id,omitempty"`
	Role              []string            `json:"role,omitempty"`
	PermissionLevel   TenantGroupType     `json:"permissionLevel,omitempty"`
	GroupMemberships  []GroupMember       `json:"group_memberships,omitempty"`  // Optional: List of group IDs the user belongs to
	TenantMemberships []TenantMemberships `json:"tenant_memberships,omitempty"` // Optional: List of tenant IDs the user belongs to
}

func NewTenantCustomClaims(tenantID *string, role []string, owner *Owner) *TenantCustomClaims {
	return &TenantCustomClaims{
		TenantID:          *tenantID,
		Role:              role,
		PermissionLevel:   TenantGroupOwner,
		GroupMemberships:  owner.GroupMemberships,
		TenantMemberships: owner.TenantMemberships,
	}
}

func (c *TenantCustomClaims) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"tenant_id":          c.TenantID,
		"role":               c.Role,
		"permissionLevel":    c.PermissionLevel,
		"group_memberships":  c.GroupMemberships,
		"tenant_memberships": c.TenantMemberships,
	}
}
