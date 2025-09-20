package entity_user

import (
	"errors"
	"time"
)

type UserResponse struct {
	*User     `bson:",inline"`
	ID        string    `bson:"_id,omitempty"  json:"_id" binding:"required"`
	CreatedAt time.Time `bson:"created_at" json:"-"`
	UpdatedAt time.Time `bson:"updated_at" json:"-"`
}

func NewUserResponse(id string, user *User) (*UserResponse, error) {
	u := &UserResponse{
		User: user,
		ID:   id,
	}

	err := u.Validate()
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (u *UserResponse) Validate() error {

	if u == nil {
		return errors.New("invalid user response")
	}

	if !u.IsID() {
		return errors.New("invalid id")
	}

	if err := u.User.Validate(); err != nil {
		return err
	}

	return nil
}

func (u *UserResponse) IsID() bool {
	if u == nil {
		return false
	}
	return u.ID != ""
}

func (u *UserResponse) GetID() string {
	if u == nil {
		return ""
	}
	return u.ID
}

func (u *UserResponse) GetUser() *User {
	if u == nil {
		return nil
	}
	return u.User
}
