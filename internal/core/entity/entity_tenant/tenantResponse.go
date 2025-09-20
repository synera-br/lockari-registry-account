package entitytenant

import (
	"errors"
	"registry-account/pkg/utils"
	"time"
)

type TenantResponse struct {
	*Tenant   `bson:",inline"`
	ID        string    `bson:"_id,omitempty"  json:"_id" binding:"required"`
	CreatedAt time.Time `bson:"created_at" json:"-"`
	UpdatedAt time.Time `bson:"updated_at" json:"-"`
}

func NewTenantResponse(tenant *Tenant) (*TenantResponse, error) {
	u := &TenantResponse{
		Tenant:    tenant,
		ID:        utils.NewID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := u.Validate()
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (t *TenantResponse) Validate() error {

	if t == nil {
		return errors.New("invalid tenant response")
	}

	if !t.IsID() {
		return utils.ErrInvalidTenantID
	}

	if err := t.Tenant.Validate(); err != nil {
		return err
	}

	return nil
}

func (t *TenantResponse) IsID() bool {
	if t == nil {
		return false
	}

	return t.ID != ""
}

func (t *TenantResponse) GetID() string {
	if t == nil {
		return ""
	}
	return t.ID
}

func (t *TenantResponse) GetTenant() *Tenant {
	if t == nil {
		return nil
	}

	return t.Tenant
}
