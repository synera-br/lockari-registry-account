package authclient

import (
	"errors"
	"net/mail"
)

type User struct {
	UID           string                 `json:"uid" binding:"required"`
	Email         string                 `json:"email" binding:"required"`
	DisplayName   string                 `json:"displayName,omitempty"`
	EmailVerified bool                   `json:"emailVerified"`
	Token         string                 `json:"token" binding:"required"`
	Claims        map[string]interface{} `json:"claims,omitempty"`
	Tenants       []string               `json:"tenants,omitempty"`
}

func NewUser(uid, email, displayName string, emailVerified bool, token string) (*User, error) {
	u := &User{
		UID:           uid,
		Email:         email,
		DisplayName:   displayName,
		EmailVerified: emailVerified,
		Token:         token,
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

	if !u.IsEmail() {
		return errors.New("invalid email")
	}

	if !u.IsUID() {
		return errors.New("invalid uid")
	}

	if !u.IsToken() {
		return errors.New("invalid token")
	}

	return nil
}

func (u *User) IsEmail() bool {
	if u == nil {
		return false
	}
	if err := u.isEmail(&u.Email); err != nil {
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

func (u *User) isEmail(email *string) error {
	if email == nil || *email == "" {
		return errors.New("invalid email")
	}

	_, err := mail.ParseAddress(*email)
	if err != nil {
		return errors.New("invalid email")
	}
	return nil
}
