package service_registry_request

import (
	entitytenant "registry-account/internal/core/entity/entity_tenant"
	"registry-account/internal/core/entity/entity_user"
)

type registryAccountParams struct {
	User   entity_user.UserResponse    `json:"user"`
	Tenant entitytenant.TenantResponse `json:"tenant"`
}
