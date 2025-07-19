package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	entity "lockari-api-application/internal/core/entity/audit"
	"lockari-api-application/pkg/database"
	"lockari-api-application/pkg/utils"
)

type auditSystemEvent struct {
	db         database.FirebaseDBInterface
	collection string
}

func InicializeAuditSystemEventRepository(db database.FirebaseDBInterface) (entity.AuditSystemEventRepository, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}

	return &auditSystemEvent{
		db:         db,
		collection: "system_audit",
	}, nil
}

func (r *auditSystemEvent) Create(ctx context.Context, audit map[string]interface{}) (*entity.AuditSystemEvent, error) {

	if ctx.Err() != nil {
		return nil, fmt.Errorf(utils.ContextCancelled, ctx.Err().Error())
	}

	if len(audit) == 0 {
		return nil, errors.New("invalid audit: no data provided")
	}

	response, err := r.db.Create(ctx, audit, r.collection)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit: %w", err)
	}

	auditResponse, err := r.convertToEntity(response)
	if err != nil {
		return nil, fmt.Errorf("failed to convert audit response: %w", err)
	}

	return auditResponse, nil
}

func (r *auditSystemEvent) Get(ctx context.Context, filters database.Conditional) (*entity.AuditSystemEvent, error) {
	return nil, nil
}

func (r *auditSystemEvent) List(ctx context.Context, filters database.Conditional) ([]entity.AuditSystemEvent, error) {
	return nil, nil
}

func (r *auditSystemEvent) convertToEntity(data []byte) (*entity.AuditSystemEvent, error) {

	if len(data) == 0 {
		return nil, errors.New("error to convert audit data to map")
	}

	var audit entity.AuditSystemEvent
	err := json.Unmarshal(data, &audit)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit data: %w", err)
	}

	return &audit, nil
}
