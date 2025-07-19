package service

import (
	"context"
	"errors"

	entity "lockari-api-application/internal/core/entity/audit"
	"lockari-api-application/pkg/authenticator"
	"lockari-api-application/pkg/tokengen"
	"lockari-api-application/pkg/utils"
)

type auditSystemEvent struct {
	repo     entity.AuditSystemEventRepository
	auth     authenticator.Authenticator
	tokenJWT tokengen.TokenGenerator
}

func InitializeAuditSystemEventService(repo entity.AuditSystemEventRepository, auth authenticator.Authenticator, tokenJWT tokengen.TokenGenerator) (entity.AuditSystemEventService, error) {
	if repo == nil {
		return nil, errors.New(utils.RepositoryNotFound + "AuditSystemEventRepository")
	}

	if auth == nil {
		return nil, errors.New(utils.RepositoryNotFound + "Authenticator")
	}

	if tokenJWT == nil {
		return nil, errors.New(utils.RepositoryNotFound + "TokenGenerator")
	}

	return &auditSystemEvent{
		repo:     repo,
		auth:     auth,
		tokenJWT: tokenJWT,
	}, nil
}

func (s *auditSystemEvent) Create(ctx context.Context, event *entity.AuditSystemEvent) (*entity.AuditSystemEvent, error) {

	// CONTEXT
	if ctx.Err() != nil {
		return nil, errors.New(utils.ContextCancelled)
	}

	token := utils.GetTokenFromContext(ctx) // Ensure user ID is retrieved from context
	_, err := s.tokenJWT.Validate(token)
	if err != nil {
		return nil, err
	}

	// AUDIT
	if err := event.IsValid(); err != nil {
		return nil, err
	}

	data, err := utils.StructToMap(event)
	if err != nil {
		return nil, err
	}

	result, err := s.repo.Create(ctx, data)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *auditSystemEvent) Get(ctx context.Context, id string) (*entity.AuditSystemEvent, error) {
	return nil, nil
}

func (a *auditSystemEvent) List(ctx context.Context) ([]entity.AuditSystemEvent, error) {
	return nil, nil
}
