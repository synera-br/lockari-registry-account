package entity_user

import (
	"errors"
	"fmt"
	"registry-account/pkg/utils"
	"slices"
)

// {
// 	"_id": "01995c8b-9236-7b1b-b8f2-ca11ea32615b",
// 	"email": "joao25@example.com",
// 	"uid": "vbcPA3bkfZV9t2uB7RxBkz3O4mJ2",
// 	"displayName": "João Silva",
// 	"tenants": [
// 	  {
// 		"tenantId": "tenant-uuid-here",
// 		"role": "admin", // ou "user", "viewer", etc.
// 		"permissions": ["read", "write", "delete"]
// 	  }
// 	]
//   }

type User struct {
	UID           string                 `bson:"uid" json:"uid" binding:"required"`
	Email         string                 `bson:"email"  json:"email" binding:"required"`
	DisplayName   string                 `bson:"displayName" json:"displayName,omitempty"`
	EmailVerified bool                   `bson:"emailVerified,omitempty" json:"emailVerified,omitempty"`
	Token         string                 `bson:"token" json:"token" `
	Claims        map[string]interface{} `bson:"claims" json:"claims,omitempty"`
	Tenants       []string               `bson:"tenants" json:"tenants,omitempty"`
}

func NewUser(uid, email, displayName string, emailVerified bool, token string, claims map[string]interface{}, tenants []string) (*User, error) {
	u := &User{
		UID:           uid,
		Email:         email,
		DisplayName:   displayName,
		EmailVerified: emailVerified,
		Token:         token,
		Claims:        claims,
		Tenants:       tenants,
	}

	err := u.Validate()
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (u *User) Validate() error {
	if u == nil {
		return errors.New("user is nil")
	}
	if u.UID == "" {
		return errors.New("invalid uid: cannot be empty")
	}
	if err := utils.IsEmail(&u.Email); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	// if u.Token == "" {
	// 	return errors.New("invalid token: cannot be empty")
	// }
	return nil
}

func (u *User) IsEmail() bool {
	if u == nil {
		return false
	}
	if err := utils.IsEmail(&u.Email); err != nil {
		return false
	}
	return true
}

func (u *User) IsUID() bool {
	if u == nil {
		return false
	}
	return u.UID != ""
}

func (u *User) IsToken() bool {
	if u == nil {
		return false
	}
	return u.Token != ""
}

func (u *User) GetUID() string {
	if u == nil {
		return ""
	}
	return u.UID
}

func (u *User) GetEmail() string {
	if u == nil {
		return ""
	}
	return u.Email
}

func (u *User) GetDisplayName() string {
	if u == nil {
		return ""
	}
	return u.DisplayName
}

func (u *User) GetEmailVerified() bool {
	if u == nil {
		return false
	}
	return u.EmailVerified
}

func (u *User) GetToken() string {
	if u == nil {
		return ""
	}
	return u.Token
}

func (u *User) AppendTenant(tenantID string) error {
	if u == nil {
		return errors.New("user is nil")
	}
	if tenantID == "" {
		return errors.New("invalid tenant id")
	}

	// Evita duplicatas
	if slices.Contains(u.Tenants, tenantID) {
		return errors.New("tenant already exists for this user")
	}

	u.Tenants = append(u.Tenants, tenantID)
	return nil
}

func (u *User) RemoveTenant(tenantID string) error {
	if u == nil {
		return errors.New("user is nil")
	}
	if tenantID == "" {
		return errors.New("invalid tenant id")
	}

	found := false
	// 'j' será o índice para o novo slice sem o tenant a ser removido
	j := 0
	for _, t := range u.Tenants {
		if t == tenantID {
			found = true
		} else {
			u.Tenants[j] = t
			j++
		}
	}

	if !found {
		return errors.New("tenant not found")
	}

	// Trunca o slice para o novo tamanho
	u.Tenants = u.Tenants[:j]
	return nil
}

func (u *User) AppendClaim(key string, value interface{}) error {
	if u == nil {
		return errors.New("user is nil")
	}
	if key == "" {
		return errors.New("invalid key")
	}
	if value == nil {
		return errors.New("invalid value")
	}

	if u.Claims == nil {
		u.Claims = make(map[string]interface{})
	}

	u.Claims[key] = value

	return nil
}

func (u *User) RemoveClaim(key string) error {
	if u == nil {
		return errors.New("user is nil")
	}
	if key == "" {
		return errors.New("invalid key")
	}

	if _, ok := u.Claims[key]; !ok {
		return errors.New("claim not found")
	}

	delete(u.Claims, key)

	return nil
}

func (u *User) UpdateClaim(key string, value interface{}) error {
	if u == nil {
		return errors.New("user is nil")
	}
	if key == "" {
		return errors.New("invalid key")
	}

	if value == nil || value == "" {
		return errors.New("invalid value")
	}

	if _, ok := u.Claims[key]; !ok {
		return errors.New("claim not found")
	}

	u.Claims[key] = value

	return nil
}
