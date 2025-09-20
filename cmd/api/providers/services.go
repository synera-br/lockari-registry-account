package providers

import (
	"fmt"
	"registry-account/cmd/api/types"
	"registry-account/internal/core/entity/entity_registry_request"
	entitytenant "registry-account/internal/core/entity/entity_tenant"
	"registry-account/internal/core/entity/entity_user"
	repositorytenant "registry-account/internal/core/repository/repository_tenant"
	"registry-account/internal/core/repository/repository_user"
	"registry-account/internal/core/service/service_registry_request"
	servicetenant "registry-account/internal/core/service/service_tenant"
	serviceuser "registry-account/internal/core/service/service_user"
)

func ProvideUserService(params types.UserServiceParams) (entity_user.UserRequestData, error) {
	repo, err := repository_user.NewRepositoryUser(params.DB, params.Obs, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create user repository: %w", err)
	}

	svc, err := serviceuser.NewServiceUser(repo, params.MQ, nil, params.Obs, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create user service: %w", err)
	}

	return svc, nil
}

func ProvideTenantService(params types.TenantServiceParams) (entitytenant.TenantRequestData, error) {
	repo, err := repositorytenant.NewRepositoryTenant(params.DB, params.Obs, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant repository: %w", err)
	}

	svc, err := servicetenant.NewServiceTenant(repo, params.MQ, nil, params.Obs, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant service: %w", err)
	}

	return svc, nil
}

func ProvideRegistryService(params types.RegistryServiceParams) (entity_registry_request.ServiceRegistryRequest, error) {
	svc, err := service_registry_request.NewServiceRequest(
		params.Auth, params.UserSvc, params.TenantSvc, params.MQ, nil, params.Obs, params.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry service: %w", err)
	}

	return svc, nil
}
