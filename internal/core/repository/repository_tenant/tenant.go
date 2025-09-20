package repositorytenant

import (
	"context"
	"errors"
	"fmt"
	entitytenant "registry-account/internal/core/entity/entity_tenant"
	"registry-account/pkg/logger"
	"registry-account/pkg/telemetry"
	"registry-account/pkg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repositoryTenantParams struct {
	tracerName         string
	collectionUserName string
	db                 *mongo.Collection
	obs                telemetry.OtelObservability
	log                logger.LoggerInterface
}

func NewRepositoryTenant(db *mongo.Database,
	obs telemetry.OtelObservability,
	log logger.LoggerInterface,
) (entitytenant.TenantResponseData, error) {
	if db == nil {
		return nil, errors.New("repository user create is nil")
	}
	collection := "tenants"
	repo := &repositoryTenantParams{
		tracerName:         "repositoryTenant",
		collectionUserName: collection,
		db:                 db.Collection(collection),
		obs:                obs,
		log:                log,
	}
	if err := repo.createCaseInsensitiveUniqueIndex(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *repositoryTenantParams) Create(ctx context.Context, tenant *entitytenant.Tenant) (*entitytenant.TenantResponse, error) {
	ctx, span := s.startSpan(ctx, "repository tenant create")
	defer span.End()
	span.SetAttributes(attribute.String("method", "repositoryCreateTenant"))

	if err := tenant.Validate(); err != nil {
		return nil, fmt.Errorf("repository tenant create: %w", err)
	}

	response, err := entitytenant.NewTenantResponse(tenant)
	if err != nil {
		return nil, fmt.Errorf("repository tenant create: %w", err)
	}

	if response == nil {
		return nil, fmt.Errorf("repository tenant create: %w", utils.ErrInvalidTenantResponse)
	}

	result, err := s.db.InsertOne(ctx, bson.M{
		"_id":         response.ID,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
		"name":        response.Name,
		"slug":        response.Slug,
		"owner":       response.Owner,
		"displayName": response.DisplayName,
		"description": response.Description,
	})

	if err != nil {
		return nil, fmt.Errorf("repository user create: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("repository user create: %w", err)
	}

	return response, nil
}

func (s *repositoryTenantParams) Get(ctx context.Context, filter *entitytenant.TenantFilter) ([]*entitytenant.TenantResponse, error) {
	ctx, span := s.startSpan(ctx, "repository tenant get")
	defer span.End()
	span.SetAttributes(attribute.String("method", "repositoryGetTenant"))

	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("repository tenant get: %w", err)
	}

	mongoFilter := bson.M{}
	if filter.ID != nil {
		mongoFilter["_id"] = *filter.ID
	}
	if filter.Name != nil {
		mongoFilter["name"] = *filter.Name
	}
	if filter.Slug != nil {
		mongoFilter["slug"] = *filter.Slug
	}
	if filter.Owner != nil {
		mongoFilter["owner"] = *filter.Owner
	}
	if filter.DisplayName != nil {
		mongoFilter["displayName"] = bson.M{
			"$regex":   *filter.DisplayName,
			"$options": "i", // "i" para case-insensitive
		}
	}
	if filter.Description != nil {
		mongoFilter["description"] = bson.M{
			"$regex":   *filter.Description,
			"$options": "i", // "i" para case-insensitive
		}
	}

	if len(mongoFilter) == 0 {
		return nil, fmt.Errorf("not found filter")
	}

	// Executar a query
	cursor, err := s.db.Find(ctx, mongoFilter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var results []*entitytenant.TenantResponse
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		return nil, utils.ErrTenantNotFound
	}

	return results, nil
}

func (r *repositoryTenantParams) createCaseInsensitiveUniqueIndex(ctx context.Context) error {
	indexStaring := "name"
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: indexStaring, Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetCollation(&options.Collation{
				Locale:   "en",
				Strength: 2,
			}),
	}

	if r.indexExists(ctx, indexStaring) {
		return nil
	}

	_, err := r.db.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create case-insensitive unique index: %w", err)
	}

	return nil
}

func (r *repositoryTenantParams) indexExists(ctx context.Context, indexName string) bool {
	cursor, err := r.db.Indexes().List(ctx)
	if err != nil {
		return false
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			continue
		}
		if name, ok := index["name"].(string); ok && name == indexName {
			return true
		}
	}
	return false
}

func (s *repositoryTenantParams) startSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if s.obs == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return s.obs.Span(ctx, operationName)
}

func (s *repositoryTenantParams) spanError(ctx context.Context, err error) {
	span := s.obs.Trace(ctx)
	span.RecordError(err)
	span.SetAttributes(attribute.String("error", err.Error()))
}
