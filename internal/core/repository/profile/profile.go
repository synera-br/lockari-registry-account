package profile

import (
	"context"
	"errors"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/profile"
	"github.com/synera-br/lockari-backend-app/pkg/database"
)

type userProfileRepository struct {
	db         database.FirebaseDBInterface
	collection string
}

func InitializeUserProfileRepository(db database.FirebaseDBInterface) (entity.UserProfileRepository, error) {

	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	if !db.IsConnected() {
		return nil, errors.New("database connection is not initialized")
	}

	return &userProfileRepository{
		db:         db,
		collection: "profiles",
	}, nil
}

func (r *userProfileRepository) CreateProfile(ctx context.Context, profile *entity.Profile) (*entity.Profile, error) {
	return nil, errors.New("not implemented")
}

func (r *userProfileRepository) UpdateProfile(ctx context.Context, id *string, profile *entity.Profile) (*entity.Profile, error) {
	return nil, errors.New("not implemented")
}

func (r *userProfileRepository) DeleteProfile(ctx context.Context, id *string) error {
	return errors.New("not implemented")
}

func (r *userProfileRepository) GetProfile(ctx context.Context, filter map[string]interface{}) (*entity.ProfileResponse, error) {

	return nil, errors.New("not implemented")
}
