package entitytenant

import (
	"context"
)

type TenantRequestData interface {
	Create(ctx context.Context, tenant *Tenant) (*TenantResponse, error)
	Get(ctx context.Context, filter *TenantFilter) ([]*TenantResponse, error)
}

type TenantResponseData interface {
	Create(ctx context.Context, tenant *Tenant) (*TenantResponse, error)
	Get(ctx context.Context, filter *TenantFilter) ([]*TenantResponse, error)
}
