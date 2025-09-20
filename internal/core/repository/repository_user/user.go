package repository_user

import (
	"context"
	"errors"
	"fmt"
	"registry-account/internal/core/entity/entity_user"
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

type repositoryUserParams struct {
	tracerName         string
	collectionUserName string
	db                 *mongo.Collection
	obs                telemetry.OtelObservability
	log                logger.LoggerInterface
}

func NewRepositoryUser(db *mongo.Database,
	obs telemetry.OtelObservability,
	log logger.LoggerInterface,
) (entity_user.UserResponseData, error) {
	if db == nil {
		return nil, errors.New("repository user create is nil")
	}

	repo := &repositoryUserParams{
		tracerName:         "repositoryRegistryUser",
		collectionUserName: "users",
		db:                 db.Collection("users"),
		obs:                obs,
		log:                log,
	}

	if err := repo.createCaseInsensitiveUniqueIndex(context.Background(), repo.db); err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *repositoryUserParams) Create(ctx context.Context, user *entity_user.User) (*entity_user.UserResponse, error) {

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("repository user create: %w", err)
	}

	userID := utils.NewID()
	response := entity_user.UserResponse{
		ID:        userID,
		User:      user,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := s.db.InsertOne(ctx, bson.M{
		"_id":         userID,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
		"uid":         user.UID,
		"email":       user.Email,
		"displayName": user.DisplayName,
		"token":       user.Token,
		"claims":      user.Claims,
		"tenants":     user.Tenants,
	})
	if err != nil {
		return nil, fmt.Errorf("repository user create: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("repository user create: %w", err)
	}

	return &response, nil
}

func (s *repositoryUserParams) Get(ctx context.Context, filter *entity_user.UserFilter) ([]*entity_user.UserResponse, error) {
	ctx, span := s.startSpan(ctx, "repository user get")
	defer span.End()
	span.SetAttributes(attribute.String("method", "repositoryGetUser"))

	// Criar o filtro base
	mongoFilter := bson.M{}
	if filter.ID != nil {
		mongoFilter["_id"] = *filter.ID
	}
	if filter.Email != nil {
		mongoFilter["email"] = *filter.Email
	}
	if filter.UUID != nil {
		mongoFilter["uuid"] = *filter.UUID
	}
	if filter.Tenants != nil && len(*filter.Tenants) > 0 {
		// Para arrays, você pode usar $in para match qualquer valor do array
		mongoFilter["tenants"] = bson.M{"$in": *filter.Tenants}
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

	var results []*entity_user.UserResponse

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}
	// Verificar se houve erro durante a iteração
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if results == nil {
		return nil, utils.UserNotFound
	}

	return results, nil
}

func (s *repositoryUserParams) indexExists(ctx context.Context, collection *mongo.Collection, indexName string) bool {
	cursor, err := collection.Indexes().List(ctx)
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

// Função para criar índice único no email
func (s *repositoryUserParams) createCaseInsensitiveUniqueIndex(ctx context.Context, collection *mongo.Collection) error {
	indexStaring := "email"
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: indexStaring, Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetCollation(&options.Collation{
				Locale:   "en",
				Strength: 2, // Case insensitive
			}),
	}

	if s.indexExists(ctx, collection, indexStaring) {
		return nil
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create case-insensitive unique index: %w", err)
	}

	return nil
}

func (s *repositoryUserParams) startSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if s.obs == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return s.obs.Span(ctx, operationName)
}

func (s *repositoryUserParams) spanError(ctx context.Context, err error) {
	span := s.obs.Trace(ctx)
	span.RecordError(err)
	span.SetAttributes(attribute.String("error", err.Error()))
}
