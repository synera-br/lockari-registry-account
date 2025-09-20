package providers

import (
	"go.uber.org/fx"
)

// Module agrupa todos os providers em um módulo reutilizável
var Module = fx.Options(
	fx.Provide(
		ProvideConfig,
		ProvideTelemetry,
		ProvidePropagator,
		ProvideLogger,
		ProvideMessageQueue,
		ProvideAuth,
		ProvideAuthz,
		ProvideDatabase,
		ProvideUserService,
		ProvideTenantService,
		ProvideRegistryService,
		ProvideWebServer,
	),
)
