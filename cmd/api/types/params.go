package types

import (
	"registry-account/configs"
	"registry-account/internal/core/entity/entity_registry_request"
	entitytenant "registry-account/internal/core/entity/entity_tenant"
	"registry-account/internal/core/entity/entity_user"
	"registry-account/pkg/authclient"
	"registry-account/pkg/logger"
	"registry-account/pkg/messagequeue"
	"registry-account/pkg/telemetry"
	"registry-account/pkg/webserver"

	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/fx"
)

// Parâmetros de configuração para diferentes componentes
type ConfigParams struct {
	fx.In
	Config *configs.AppConfig
}

type TelemetryParams struct {
	fx.In
	Config *configs.AppConfig
}

type LoggerParams struct {
	fx.In
	Config *configs.AppConfig
}

type MessageQueueParams struct {
	fx.In
	Config     *configs.AppConfig
	Obs        telemetry.OtelObservability
	Propagator propagation.TextMapPropagator
}

type AuthParams struct {
	fx.In
	Config *configs.AppConfig
}

type AuthzParams struct {
	fx.In
	Config *configs.AppConfig
	Obs    telemetry.OtelObservability
}

type DatabaseParams struct {
	fx.In
	Config *configs.AppConfig
}

type UserServiceParams struct {
	fx.In
	Auth   authclient.AuthenticationProvider
	DB     *mongo.Database
	MQ     messagequeue.MessageQueue
	Obs    telemetry.OtelObservability
	Logger logger.LoggerInterface
}

type TenantServiceParams struct {
	fx.In
	DB     *mongo.Database
	MQ     messagequeue.MessageQueue
	Obs    telemetry.OtelObservability
	Logger logger.LoggerInterface
}

type RegistryServiceParams struct {
	fx.In
	Auth      authclient.AuthenticationProvider
	UserSvc   entity_user.UserRequestData
	TenantSvc entitytenant.TenantRequestData
	MQ        messagequeue.MessageQueue
	Obs       telemetry.OtelObservability
	Logger    logger.LoggerInterface
}

type WebServerParams struct {
	fx.In
	Config  *configs.AppConfig
	Auth    authclient.AuthenticationProvider
	Obs     telemetry.OtelObservability
	Tenant  entitytenant.TenantRequestData
	User    entity_user.UserRequestData
	Cleanup func() `optional:"true"`
}

type AppParams struct {
	fx.In
	Router          webserver.ServerInterface
	Auth            authclient.AuthenticationProvider
	RegistryService entity_registry_request.ServiceRegistryRequest
	Obs             telemetry.OtelObservability
	Logger          logger.LoggerInterface
	MQ              messagequeue.MessageQueue
	Tenant          entitytenant.TenantRequestData
	User            entity_user.UserRequestData
}
