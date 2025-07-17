package entity

type TenantCustomClaims struct {
	TenantID          string              `json:"tenantId,omitempty"`
	Role              []string            `json:"role,omitempty"`
	PermissionLevel   TenantGroupType     `json:"permissionLevel,omitempty"`
	GroupMemberships  []GroupMember       `json:"group_memberships,omitempty"`  // Optional: List of group IDs the user belongs to
	TenantMemberships []TenantMemberships `json:"tenant_memberships,omitempty"` // Optional: List of tenant IDs the user belongs to
}

func NewTenantCustomClaims(tenantID *string, role []string, owner *Owner) *TenantCustomClaims {
	if tenantID == nil || *tenantID == "" {
		return nil
	}
	if owner == nil {
		return nil
	}

	tenant := &TenantMemberships{
		TenantID: *tenantID,
		Role:     "owner",
	}

	if owner.TenantMemberships != nil {
		for _, m := range owner.TenantMemberships {
			if m.TenantID == *tenantID {
				tenant = &TenantMemberships{
					TenantID: *tenantID,
					Role:     m.Role,
				}
			}
		}
	}

	tenants := make([]TenantMemberships, 0, len(owner.TenantMemberships))
	tenants = append(tenants, owner.TenantMemberships...)
	tenants = append(tenants, *tenant)

	return &TenantCustomClaims{
		TenantID:          *tenantID,
		Role:              role,
		PermissionLevel:   TenantGroupOwner,
		GroupMemberships:  owner.GroupMemberships,
		TenantMemberships: tenants,
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
