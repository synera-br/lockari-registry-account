package dto_registry_request

import "fmt"

type (
	UserDataParams struct {
		Name       string `json:"name" binding:"required"`
		Email      string `json:"email" binding:"required"`
		TenantName string `json:"tenantName" binding:"required"`
	}

	RegistryRequestParams struct {
		Token    string         `json:"token" binding:"required"`
		UserData UserDataParams `json:"userData" binding:"required"`
	}
)

func (d *UserDataParams) Validate() error {
	if d == nil {
		return fmt.Errorf("user data is nil")
	}

	if d.Name == "" {
		return fmt.Errorf("name is required")
	}

	if d.Email == "" {
		return fmt.Errorf("email is required")
	}

	if d.TenantName == "" {
		return fmt.Errorf("tenant name is required")
	}

	return nil
}

func (d *RegistryRequestParams) Validate() error {
	if d == nil {
		return fmt.Errorf("user data is nil")
	}

	if d.Token == "" {
		return fmt.Errorf("token is required")
	}

	if err := d.UserData.Validate(); err != nil {
		return fmt.Errorf("user data is invalid: %w", err)
	}

	return nil
}
