package entitytenant

import (
	"errors"
	"registry-account/pkg/utils"
	"strings"
)

// "_id": "01995c8b-9236-7b1b-b8f2-ca11ea32615c",
//   "name": "my-tenant", // único
//   "displayName": "Minha Empresa",
//   "slug": "my-tenant", // para URLs amigáveis
//   "domain": "my-tenant.suaplataforma.com",
//   "plan": "premium",
//   "status": "active",
//   "settings": {},

type Tenant struct {
	Name        string `bson:"name" json:"name" binding:"required"`
	Description string `bson:"description" json:"description,omitempty"`
	DisplayName string `bson:"displayName" json:"displayName,omitempty"`
	Slug        string `bson:"slug" json:"slug,omitempty"`
	Owner       string `bson:"owner" json:"owner" binding:"required,email"`
}

type TenantFilter struct {
	ID          *string `bson:"_id,omitempty" json:"_id"`
	Name        *string `bson:"name" json:"name,omitempty"`
	Description *string `bson:"description" json:"description,omitempty"`
	DisplayName *string `bson:"displayName" json:"displayName,omitempty"`
	Slug        *string `bson:"slug" json:"slug,omitempty"`
	Owner       *string `bson:"owner"json:"owner,omitempty"`
}

func (t *TenantFilter) Validate() error {
	if t == nil {
		return errors.New("tenant filter is nil")
	}

	filterExists := false
	if t.Name != nil && *t.Name != "" {
		filterExists = true
	}
	if t.Description != nil && *t.Description != "" {
		filterExists = true
	}
	if t.DisplayName != nil && *t.DisplayName != "" {
		filterExists = true
	}
	if t.Slug != nil && *t.Slug != "" {
		filterExists = true
	}
	if t.Owner != nil && *t.Owner != "" {
		filterExists = true
	}

	if !filterExists {
		return errors.New("no filter found")
	}

	return nil
}

func NewTenant(name, slug, owner, displayName, description string) (*Tenant, error) {

	t := &Tenant{
		Name:        name,
		Slug:        slug,
		Owner:       owner,
		DisplayName: displayName,
		Description: description,
	}

	err := t.Validate()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Tenant) Validate() error {
	if t == nil {
		return errors.New("tenant is nil")
	}

	if !t.IsValidName() {
		return utils.ErrInvalidTenantName
	}

	if !t.IsValidOwner() {
		return utils.ErrInvalidTenantOwner
	}

	if t.Slug == "" {
		t.Slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
	}

	if t.DisplayName == "" {
		t.DisplayName = strings.ToUpper(t.Name)
	}

	return nil
}

func (t *Tenant) IsValidName() bool {
	if t == nil {
		return false
	}
	return t.Name != ""
}

func (t *Tenant) IsValidOwner() bool {
	if t == nil {
		return false
	}

	if err := utils.IsEmail(&t.Owner); err != nil {
		return false
	}

	return t.Owner != ""
}

func (t *Tenant) GetName() string {
	if t == nil {
		return ""
	}

	return t.Name
}

func (t *Tenant) GetSlug() string {
	if t == nil {
		return ""
	}
	return t.Slug
}

func (t *Tenant) GetOwner() string {
	if t == nil {
		return ""
	}

	return t.Owner
}

func (t *Tenant) GetDisplayName() string {
	if t == nil {
		return ""
	}
	return t.DisplayName
}

func (t *Tenant) GetDescription() string {
	if t == nil {
		return ""
	}
	return t.Description
}
