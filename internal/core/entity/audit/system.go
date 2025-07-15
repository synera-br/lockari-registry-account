package entity

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/synera-br/lockari-backend-app/pkg/database"
)

type AuditSystemEventRepository interface {
	Create(ctx context.Context, audit map[string]interface{}) (*AuditSystemEvent, error)
	Get(ctx context.Context, filters database.Conditional) (*AuditSystemEvent, error)
	List(ctx context.Context, filters database.Conditional) ([]AuditSystemEvent, error)
}

type AuditSystemEventService interface {
	Create(ctx context.Context, event *AuditSystemEvent) (*AuditSystemEvent, error)
	Get(ctx context.Context, id string) (*AuditSystemEvent, error)
	List(ctx context.Context) ([]AuditSystemEvent, error)
}

type AuditSystemEvent struct {
	ID            string        `json:"id,omitempty"` // Optional: Unique identifier for the audit event
	EventType     EventType     `json:"eventType" binding:"required"`
	User          User          `json:"user" binding:"required"`
	ClientInfo    Client        `json:"clientInfo" binding:"required"`
	FailureReason FailureReason `json:"failureReason,omitempty"`
	Timestamp     string        `json:"timestamp" binding:"required"` // ISO 8601 format
	CreatedAt     string        `json:"createdAt,omitempty"`          // ISO 8601 format
}

func (a *AuditSystemEvent) IsValid() error {
	if a == nil {
		return errors.New("invalid audit event: event cannot be nil")
	}

	if *a == (AuditSystemEvent{}) {
		return errors.New("invalid audit event: event cannot be empty")
	}

	if err := a.User.IsValid(); err != nil {
		return err
	}

	if err := a.ClientInfo.IsValid(); err != nil {
		return err
	}

	if a.EventType == "" {
		return errors.New("invalid audit event: eventType is required")
	}

	if a.Timestamp == "" {
		return errors.New("invalid audit event: timestamp is required")
	}

	if a.CreatedAt == "" {
		a.CreatedAt = time.Now().Format(time.RFC3339)
	}

	if err := a.FailureReason.IsValid(); err != nil {
		return err
	}

	a.User.Email = strings.ToLower(a.User.Email)
	a.User.Name = strings.ToLower(a.User.Name)

	return nil
}
