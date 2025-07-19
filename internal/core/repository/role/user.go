package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	entity "lockari-api-application/internal/core/entity/role"
	"lockari-api-application/internal/core/repository"
	corev1 "lockari-api-application/pkg/core/v1"
	"lockari-api-application/pkg/database"
	"lockari-api-application/pkg/utils"
)

type userRoleRepository struct {
	db         database.FirebaseDBInterface
	collection string
}

func InitializeUserRoleRepository(db database.FirebaseDBInterface) (entity.UserRoleRepository, error) {

	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	if !db.IsConnected() {
		return nil, errors.New("database connection is not initialized")
	}

	return &userRoleRepository{
		db:         db,
		collection: "roles",
	}, nil
}

func (r *userRoleRepository) Create(ctx context.Context, tenantid *string, userRole *entity.UserRole) (*entity.UserRole, error) {

	if ctx.Err() != nil {
		return nil, errors.New("context cancelled")
	}

	if tenantid == nil {
		return nil, errors.New("tenant is required")
	}

	if userRole == nil {
		return nil, errors.New("user role is required")
	}

	toMap, err := utils.StructToMap(*userRole)
	if err != nil {
		return nil, errors.New("failed to convert user role to map: " + err.Error())
	}

	collection, err := repository.GetCollection(tenantid, &r.collection)
	if err != nil {
		return nil, err
	}

	response, err := r.db.Create(ctx, toMap, *collection)
	if err != nil {
		return nil, errors.New("failed to save tenant event to database: " + err.Error())
	}

	if response == nil {
		return nil, corev1.ErrGenericError("Failed to create tenant event")
	}

	if len(response) == 0 {
		return nil, corev1.ErrGenericError("No tenant event created")
	}

	return r.convertToEntity(response)

}

func (r *userRoleRepository) Update(ctx context.Context, tenantID, id *string, userRole *entity.UserRole) (*entity.UserRole, error) {

	return nil, errors.New("not implemented")
}

func (r *userRoleRepository) Delete(ctx context.Context, tenantID, id *string) error {

	return errors.New("not implemented")
}

func (r *userRoleRepository) Get(ctx context.Context) (*entity.UserRoleResponse, error) {

	return nil, errors.New("not implemented")
}

func (r *userRoleRepository) GetByFilter(ctx context.Context, tenantID *string, filter map[string]interface{}) (*entity.UserRoleResponse, error) {

	if ctx.Err() != nil {
		return nil, errors.New("context cancelled")
	}

	collection, err := repository.GetCollection(tenantID, &r.collection)
	if err != nil {
		return nil, err
	}

	if collection == nil {
		return nil, errors.New("collection is nil")
	}

	if filter == nil {
		return nil, errors.New("filter cannot be nil")
	}

	query := r.db.GetByQuery(ctx, *collection)
	for key, value := range filter {
		query = query.Where(key, "==", value)
	}

	response, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles by filter: %w", err)
	}

	if len(response) == 0 {
		return nil, corev1.ErrGenericError("no user roles found")
	}

	roles := &entity.UserRoleResponse{}
	if len(response) == 1 && response[0].Data() == nil {
		var userRole entity.UserRole
		if err := response[0].DataTo(&userRole); err != nil {
			return nil, fmt.Errorf("failed to convert document to user role: %w", err)
		}
		roles.UserRole = &userRole
		return roles, nil
	}

	if len(response) > 1 && response[0].Data() == nil {
		return nil, corev1.ErrGenericError("user role data is nil")
	}
	roles.UserRole = nil

	roles.UserRoles = make([]entity.UserRole, 0, len(response))

	for _, doc := range response {
		var userRole entity.UserRole
		if err := doc.DataTo(&userRole); err != nil {
			return nil, fmt.Errorf("failed to convert document to user role: %w", err)
		}
		roles.UserRoles = append(roles.UserRoles, userRole)
	}

	return roles, nil
}

func (r *userRoleRepository) convertToEntity(response []byte) (*entity.UserRole, error) {
	if response == nil {
		return nil, errors.New("response is nil")
	}

	if len(response) == 0 {
		return nil, errors.New("error to convert tenant data to map")
	}

	var userRole entity.UserRole
	err := json.Unmarshal(response, &userRole)
	if err != nil {
		fmt.Println("\nUserRole struct:", userRole)
		fmt.Println("\nUserRole data:", string(response))
		fmt.Println("\nError unmarshalling UserRole data:", err)
		return nil, errors.New("failed to unmarshal data")
	}

	return &userRole, nil

}
