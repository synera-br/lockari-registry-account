package repository

import (
	"context"
	"encoding/json"
	"errors"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/auth"
	core "github.com/synera-br/lockari-backend-app/internal/core/repository"
	"github.com/synera-br/lockari-backend-app/pkg/database"
)

type SignupEvent struct {
	db         database.FirebaseDBInterface
	collection string
}

func InitializeSignupEventRepository(db database.FirebaseDBInterface) (entity.SignupEventRepository, error) {

	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	if !db.IsConnected() {
		return nil, errors.New("database connection is not initialized")
	}

	return &SignupEvent{
		db:         db,
		collection: "subscription",
	}, nil
}

func (r *SignupEvent) Create(ctx context.Context, signup map[string]interface{}) (*entity.Signup, error) {
	if len(signup) == 0 {
		return nil, errors.New("invalid signup: no data provided")
	}

	if ctx.Err() != nil {
		return nil, errors.New("context cancelled")
	}

	// Save the signup event to the database na coleção correta
	response, err := r.db.Create(ctx, signup, r.collection)
	if err != nil {
		return nil, errors.New("failed to save signup event to database: " + err.Error())
	}

	newSignup, err := r.convertToEntity(response)
	if err != nil {
		return nil, errors.New("failed to convert response to signup entity: " + err.Error())
	}

	return newSignup, nil
}

func (r *SignupEvent) Get(ctx context.Context, filters database.Conditional) (*entity.Signup, error) {
	var signup entity.Signup

	return &signup, nil
}

func (r *SignupEvent) List(ctx context.Context, filter database.Conditional) ([]entity.Signup, error) {

	if filter.Field == "" {
		return nil, errors.New("invalid filters: no field provided")
	}
	if filter.Value == nil {
		return nil, errors.New("invalid filters: no value provided")
	}

	collection, err := core.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	var response []byte
	if filter.Field != "" {
		filters := []database.Conditional{
			filter,
		}
		response, err = r.db.GetByConditional(ctx, filters, *collection)
		if err != nil {
			return nil, err
		}
	} else {
		response, err = r.db.Get(ctx, *collection)
		if err != nil {
			return nil, err
		}
	}

	items, err := r.convertToEntities(response)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *SignupEvent) convertToEntity(data []byte) (*entity.Signup, error) {

	if len(data) == 0 {
		return nil, errors.New("error to convert signup data to map")
	}

	var signup entity.Signup
	err := json.Unmarshal(data, &signup)
	if err != nil {
		return nil, errors.New("failed to unmarshal signup data")
	}

	return &signup, nil
}

func (r *SignupEvent) convertToEntities(data []byte) ([]entity.Signup, error) {

	if len(data) == 0 {
		return nil, errors.New("error to convert signup data to map")
	}

	var signups []entity.Signup
	err := json.Unmarshal(data, &signups)
	if err != nil {
		return nil, errors.New("failed to unmarshal signup data")
	}

	return signups, nil
}
