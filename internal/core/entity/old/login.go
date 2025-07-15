package entity

import (
	"context"
	"errors"
	"time"
)

// LoginEventRepository interface defines methods for store and retrieve login events
type LoginEventRepository interface {
	Create(ctx context.Context, login LoginEvent) (LoginEvent, error)
	Get(ctx context.Context, filters []map[string]interface{}) (LoginEvent, error)
	List(context.Context) ([]LoginEvent, error)
}

// LoginEventService interface defines methods for handling login events
type LoginEventService interface {
	Create(ctx context.Context, login LoginEvent) (LoginEvent, error)
	Get(ctx context.Context, id string) (LoginEvent, error)
	List(context.Context) ([]LoginEvent, error)
}

// LoginEvent interface defines common methods for all login events
type LoginEvent interface {
	IsValid() error
	GetEventType() EventType
	GetTimestamp() time.Time
	GetUser() User
	GetClientInfo() Client
	GetTenant() string
	GetLogin() Login
	GetID() (string, bool)
}

// Login
// This event is triggered after a user successfully logs in to the application.
type Login struct {
	ID         string    `json:"id,omitempty"` // Optional: Unique identifier for the login event
	EventType  EventType `json:"eventType"`
	User       User      `json:"user" binding:"required"`
	ClientInfo Client    `json:"clientInfo" binding:"required"`
	Timestamp  time.Time `json:"timestamp" binding:"required"`
	Tenant     string    `json:"tenant,omitempty"`
}

// Implementando a interface Login
func (l *Login) GetEventType() EventType {
	return l.EventType
}

func (l *Login) GetTimestamp() time.Time {
	return l.Timestamp
}

func (l *Login) GetUser() User {
	return l.User
}

func (l *Login) GetClientInfo() Client {
	return l.ClientInfo
}

func (l *Login) GetTenant() string {
	return l.Tenant
}

func (l *Login) GetLogin() Login {
	return *l
}

func (l *Login) GetID() (string, bool) {
	isValid := false
	if l.ID != "" {
		isValid = true
	}
	return l.ID, isValid
}

// IsValid
// This method validates the Login struct to ensure that required fields are present.
func (l *Login) IsValid() (err error) {
	if err = l.User.IsValid(); err != nil {
		return err
	}
	if err = l.ClientInfo.IsValid(); err != nil {
		return err
	}
	if l.EventType == "" {
		err = errors.New("invalid login: eventType is required")
	}
	if l.Timestamp.IsZero() {
		err = errors.New("invalid login: timestamp is required")
	}
	return err
}

// NewLogin creates a new Login event with current timestamp
func NewLogin(user User, clientInfo Client) LoginEvent {
	return &Login{
		EventType:  LOGIN_SUCCESS,
		User:       user,
		ClientInfo: clientInfo,
		Timestamp:  time.Now(),
	}
}
