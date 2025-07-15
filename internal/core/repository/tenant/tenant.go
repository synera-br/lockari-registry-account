package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	corev1 "github.com/synera-br/lockari-backend-app/pkg/core/v1"
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

func (r *tenantEventRepository) Get(ctx context.Context, filters *entity.TenantFilter) (*entity.Tenant, error) {
	if ctx.Err() != nil {
		return nil, errors.New("context cancelled")
	}

	if filters == nil {
		return nil, errors.New("tenant is required")
	}

	conditionals := database.Conditionals{}
	objectName := ""
	if filters.Plan != "" {
		c := database.Conditional{
			Field:  "plan",
			Value:  filters.Plan,
			Filter: database.FilterEquals,
		}
		conditionals = append(conditionals, c)
		objectName = filters.Plan
	}
	if filters.TenantID != "" {
		c := database.Conditional{
			Field:  "tenant_id",
			Value:  filters.TenantID,
			Filter: database.FilterEquals,
		}
		conditionals = append(conditionals, c)
		objectName = filters.TenantID
	}
	if filters.Email != "" {
		c := database.Conditional{
			Field:  "email",
			Value:  filters.Email,
			Filter: database.FilterEquals,
		}
		conditionals = append(conditionals, c)
		objectName = filters.Email
	}
	if filters.Name != "" {
		c := database.Conditional{
			Field:  "name",
			Value:  filters.Name,
			Filter: database.FilterEquals,
		}
		conditionals = append(conditionals, c)
		objectName = filters.Name
	}

	response, err := r.db.GetByConditional(ctx, conditionals, r.collection)
	if err != nil {
		return nil, errors.New("failed to get tenant event from database: " + err.Error())
	}
	if len(response) == 0 {
		return nil, fmt.Errorf(corev1.TenantNotFoundError, objectName)
	}

	tenant, err := r.convertToEntity(response)
	if err != nil {
		return nil, errors.New("failed to convert response to tenant entity: " + err.Error())
	}

	return tenant, nil
}

func (r *tenantEventRepository) List(ctx context.Context, filters []entity.TenantFilter) ([]entity.Tenant, error) {
	if ctx.Err() != nil {
		return nil, errors.New("context cancelled")
	}

	var response []byte
	var err error

	if len(filters) > 0 {
		conditionals := make([]database.Conditional, len(filters))
		for i, filter := range filters {
			if filter.Plan != "" {
				conditionals[i] = database.Conditional{
					Field:  "plan",
					Value:  filter.Plan,
					Filter: database.FilterEquals,
				}
			}
			if filter.TenantID != "" {
				conditionals[i] = database.Conditional{
					Field:  "tenant_id",
					Value:  filter.TenantID,
					Filter: database.FilterEquals,
				}
			}
			if filter.Email != "" {
				conditionals[i] = database.Conditional{
					Field:  "email",
					Value:  filter.Email,
					Filter: database.FilterEquals,
				}
			}
			if filter.Name != "" {
				conditionals[i] = database.Conditional{
					Field:  "name",
					Value:  filter.Name,
					Filter: database.FilterEquals,
				}
			}
		}
		response, err = r.db.GetByConditional(ctx, conditionals, r.collection)
		if err != nil {
			return nil, errors.New("failed to get tenants by conditional: " + err.Error())
		}
	} else {
		response, err = r.db.Get(ctx, r.collection)
		if err != nil {
			return nil, errors.New("failed to get tenants: " + err.Error())
		}
	}
	if len(response) == 0 {
		return nil, errors.New("no tenants found")
	}

	tenants, err := r.convertToEntities(response)
	if err != nil {
		return nil, errors.New("failed to convert response to tenant entity: " + err.Error())
	}

	return tenants, nil
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

func (r *tenantEventRepository) convertToEntities(data []byte) ([]entity.Tenant, error) {

	if len(data) == 0 {
		return nil, errors.New("error to convert tenant data to map")
	}

	var tenants []entity.Tenant
	err := json.Unmarshal(data, &tenants)
	if err != nil {
		return nil, errors.New("failed to unmarshal tenant data")
	}

	return tenants, nil
}
