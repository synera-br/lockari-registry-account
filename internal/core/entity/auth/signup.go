package entity

import (
	"context"
	"errors"
	"time"

	"github.com/synera-br/lockari-backend-app/pkg/database"
)

// SignupEventRepository interface defines methods for store and retrieve signup events
type SignupEventRepository interface {
	Create(ctx context.Context, filters map[string]interface{}) (*Signup, error)
	Get(ctx context.Context, filters database.Conditional) (*Signup, error)
	List(ctx context.Context, filters database.Conditional) ([]Signup, error)
}

// SignupEventService interface defines methods for handling signup events
type SignupEventService interface {
	Create(ctx context.Context, signup *Signup) (SignupEvent, error)
	Get(ctx context.Context, id string) (SignupEvent, error)
	List(context.Context) ([]SignupEvent, error)
}

// SignupEvent interface defines common methods for all signup events
type SignupEvent interface {
	IsValid() error
	GetEventType() EventType
	GetTimestamp() time.Time
	GetUser() User
	GetClientInfo() Client
	GetTenant() string
	GetSignup() Signup
	SetTenant(tenant *string) error
	GetID() (string, bool) // Returns ID and a boolean indicating if the ID is valid
}

type SignupFilter struct {
	User
	Tenant string `json:"tenant,omitempty"`
}

type Tenant struct {
	TenantID string `json:"tenant_id,omitempty"` // Optional: Unique identifier for the tenant
	Name     string `json:"name,omitempty"`      // Optional: Name of the tenant
	Plan     string `json:"plan,omitempty"`      // Subscription plan of the user
}

// Signup
// This event is triggered when a user successfully signs up for the application.
type Signup struct {
	ID        string    `json:"id,omitempty"`        // Optional: Unique identifier for the signup event
	EventType EventType `json:"eventType,omitempty"` // Event type, e.g., SIGNUP_SUCCESS
	User
	ClientInfo Client    `json:"clientInfo" binding:"required"`
	Timestamp  time.Time `json:"timestamp" binding:"required"`
	Tenant     string    `json:"tenant,omitempty"` // Optional: Tenant information if applicable
}

// Implementando a interface Signup
func (s *Signup) GetEventType() EventType {
	return s.EventType
}

func (s *Signup) GetTimestamp() time.Time {
	return s.Timestamp
}

func (s *Signup) GetUser() User {
	return s.User
}

func (s *Signup) GetClientInfo() Client {
	return s.ClientInfo
}

func (s *Signup) GetTenant() string {
	return s.Tenant
}

func (s *Signup) GetID() (string, bool) {

	isValid := false
	if s.ID != "" {
		isValid = true
	}

	return s.ID, isValid
}

func (s *Signup) GetSignup() Signup {
	return *s
}

func (s *Signup) SetTenant(tenant *string) error {
	if tenant == nil {
		return errors.New("invalid signup: tenant cannot be empty")
	}

	if len(*tenant) != 36 {
		return errors.New("invalid signup: tenant must be a valid UUID")
	}

	if s.Tenant == *tenant {
		return nil
	}

	if s.Tenant != "" {
		return errors.New("invalid signup: tenant is already set")
	}

	s.Tenant = *tenant

	if s.Tenant == "" {
		return errors.New("invalid signup: tenant is required")
	}

	return nil
}

// IsValid
// This method validates the Signup struct to ensure that required fields are present.
func (s *Signup) IsValid() (err error) {

	if s == nil {
		return errors.New("invalid signup: signup event cannot be nil")
	}

	if err = s.User.IsValid(); err != nil {
		return err
	}
	if err = s.ClientInfo.IsValid(); err != nil {
		return err
	}

	if s.EventType == "" {
		s.EventType = SIGNUP_SUCCESS
	}

	if s.Timestamp.IsZero() {
		err = errors.New("invalid signup: timestamp is required")
	}

	if s.User.Name == "" {
		err = errors.New("invalid signup: user name is required")
	}

	if s.User.Plan == "" {
		err = errors.New("invalid signup: user plan is required")
	}

	return err
}

// NewSignup creates a new Signup event with current timestamp
func NewSignup(user User, clientInfo Client, tenant string) SignupEvent {
	return &Signup{
		EventType:  SIGNUP_SUCCESS,
		User:       user,
		ClientInfo: clientInfo,
		Timestamp:  time.Now(),
		Tenant:     tenant,
	}
}
