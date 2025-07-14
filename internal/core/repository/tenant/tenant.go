package tenant

import (
	"context"
	"encoding/json"
	"errors"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	"github.com/synera-br/lockari-backend-app/pkg/database"
	"github.com/synera-br/lockari-backend-app/pkg/utils"
)

type tenantEventRepository struct {
	db         database.FirebaseDBInterface
	collection string
}

func InitializeTenantEventRepository(db database.FirebaseDBInterface) (entity.TenantRepository, error) {

	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	if !db.IsConnected() {
		return nil, errors.New("database connection is not initialized")
	}

	return &tenantEventRepository{
		db:         db,
		collection: "tenant",
	}, nil
}

func (r *tenantEventRepository) Create(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {

	if ctx.Err() != nil {
		return nil, errors.New("context cancelled")
	}

	if tenant == nil {
		return nil, errors.New("tenant is required")
	}

	toMap, err := utils.StructToMap(*tenant)
	if err != nil {
		return nil, errors.New("failed to convert tenant to map: " + err.Error())
	}

	response, err := r.db.Create(ctx, toMap, r.collection)
	if err != nil {
		return nil, errors.New("failed to save tenant event to database: " + err.Error())
	}

	return r.convertToEntity(response)
}

func (r *tenantEventRepository) Get(ctx context.Context, filters entity.TenantFilter) (*entity.Tenant, error) {
	return nil, errors.New("not implemented")
}

func (r *tenantEventRepository) List(ctx context.Context, filters []entity.TenantFilter) ([]entity.Tenant, error) {
	return nil, errors.New("not implemented")
}

func (r *tenantEventRepository) Update(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	return nil, errors.New("not implemented")
}

func (r *tenantEventRepository) Delete(ctx context.Context, tenantID string) error {
	return errors.New("not implemented")
}

func (r *tenantEventRepository) convertToEntity(response []byte) (*entity.Tenant, error) {
	if response == nil {
		return nil, errors.New("response is nil")
	}

	if len(response) == 0 {
		return nil, errors.New("error to convert tenant data to map")
	}

	var tenant entity.Tenant
	err := json.Unmarshal(response, &tenant)
	if err != nil {
		return nil, errors.New("failed to unmarshal tenant data")
	}

	return &tenant, nil

}
