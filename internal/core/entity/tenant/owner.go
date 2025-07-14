package entity

import "errors"

// User
// This event is triggered when a user performs an action that requires authentication, such as logging in or signing up.
type Owner struct {
	Uid      string `json:"uid" binding:"required"`   // Unique identifier for the user in Firebase Authentication
	Email    string `json:"email" binding:"required"` // Email address of the user
	Username string `json:"username,omitempty"`       // Optional: Username of the tenant
}

func (u *Owner) IsValid() error {
	if u == nil {
		return errors.New("invalid user: user event cannot be nil")
	}

	if u.Uid == "" {
		return errors.New("invalid user: uid is required")
	}

	if u.Email == "" {
		return errors.New("invalid user: email is required")
	}

	return nil
}

func (u *Owner) GetUid() string {
	if u == nil {
		return ""
	}
	return u.Uid
}

func (u *Owner) GetEmail() string {
	if u == nil {
		return ""
	}
	return u.Email
}

func (u *Owner) GetUsername() string {
	if u == nil {
		return ""
	}
	return u.Username
}

func (u *Owner) SetUsername(username string) {
	if u == nil {
		return
	}
	u.Username = username
}
func (u *Owner) GetUser() Owner {
	if u == nil {
		return Owner{}
	}
	return *u
}
