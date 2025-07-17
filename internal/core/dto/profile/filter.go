package dto

import (
	"errors"
	"time"
)

// Sugestão melhorada
type ProfileFilter struct {
	// Campos básicos
	Uid      *string `json:"uid,omitempty"`
	Email    *string `json:"email,omitempty"`
	Username *string `json:"username,omitempty"`
	TenantID *string `json:"tenant_id,omitempty"`
	FullName *string `json:"full_name,omitempty"`
	Role     *string `json:"role,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`

	// Campos para paginação
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`

	// Campos para ordenação
	OrderBy   *string `json:"order_by,omitempty"`   // "created_at", "email", etc.
	SortOrder *string `json:"sort_order,omitempty"` // "asc", "desc"

	// Campos para busca
	Search *string `json:"search,omitempty"` // Busca em múltiplos campos

	// Campos para filtro por data
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
}

func (f *ProfileFilter) IsValid() error {
	if f == nil {
		return errors.New("filter cannot be nil")
	}

	// Validar se pelo menos um campo está presente
	if f.isEmpty() {
		return errors.New("at least one filter field must be provided")
	}

	// Validar limit
	if f.Limit != nil && *f.Limit <= 0 {
		return errors.New("limit must be greater than 0")
	}

	// Validar offset
	if f.Offset != nil && *f.Offset < 0 {
		return errors.New("offset must be greater than or equal to 0")
	}

	// Validar order by
	if f.OrderBy != nil {
		validFields := []string{"uid", "email", "username", "created_at", "full_name"}
		if !contains(validFields, *f.OrderBy) {
			return errors.New("invalid order_by field")
		}
	}

	// Validar sort order
	if f.SortOrder != nil {
		if *f.SortOrder != "asc" && *f.SortOrder != "desc" {
			return errors.New("sort_order must be 'asc' or 'desc'")
		}
	}

	return nil
}

func (f *ProfileFilter) isEmpty() bool {
	return f.Uid == nil && f.Email == nil && f.Username == nil &&
		f.TenantID == nil && f.FullName == nil && f.Role == nil &&
		f.IsActive == nil && f.Search == nil
}

func (f *ProfileFilter) ToMap() map[string]interface{} {
	if f == nil {
		return map[string]interface{}{}
	}

	filter := make(map[string]interface{})

	// Campos de filtro
	if f.Uid != nil {
		filter["uid"] = *f.Uid
	}
	if f.Email != nil {
		filter["email"] = *f.Email
	}
	if f.Username != nil {
		filter["username"] = *f.Username
	}
	if f.TenantID != nil {
		filter["tenant_id"] = *f.TenantID
	}
	if f.FullName != nil {
		filter["full_name"] = *f.FullName
	}
	if f.Role != nil {
		filter["role"] = *f.Role
	}
	if f.IsActive != nil {
		filter["is_active"] = *f.IsActive
	}

	// Campos de busca
	if f.Search != nil {
		filter["search"] = *f.Search
	}

	// Campos de data
	if f.CreatedAfter != nil {
		filter["created_after"] = *f.CreatedAfter
	}
	if f.CreatedBefore != nil {
		filter["created_before"] = *f.CreatedBefore
	}

	return filter
}

// Método separado para opções de query
func (f *ProfileFilter) GetQueryOptions() map[string]interface{} {
	options := make(map[string]interface{})

	if f.Limit != nil {
		options["limit"] = *f.Limit
	}
	if f.Offset != nil {
		options["offset"] = *f.Offset
	}
	if f.OrderBy != nil {
		options["order_by"] = *f.OrderBy
	}
	if f.SortOrder != nil {
		options["sort_order"] = *f.SortOrder
	}

	return options
}

// Helper function to check if a string slice contains a specific value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
