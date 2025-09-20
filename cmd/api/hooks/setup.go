package hooks

import (
	"context"
	"registry-account/cmd/api/types"
	"registry-account/internal/infrastructure/handler/webhandler/handler_registry_request"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/fx"
)

// Hook de inicialização para configurar handlers
func SetupHandlers(params types.AppParams) error {
	return handler_registry_request.InitializeHandlerRegistryRequest(
		params.Router, params.Auth, params.RegistryService, params.Obs, params.Logger, params.Tenant, params.User,
	)
}

// Hook de inicialização para logging
func LogAppStart(params types.AppParams) {

	// Adicionar span de telemetry
	_, span := params.Obs.Span(context.Background(), "app_startup")
	defer span.End()
	span.SetAttributes(attribute.String("status", "initialized"))
}

// Hook para iniciar o servidor web
func StartWebServer(lc fx.Lifecycle, params types.AppParams) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := params.Router.Start(); err != nil {
					params.Logger.Error("failed to start web server", "error", err)
				}
			}()
			return nil
		},
		OnStop: onStop(params),
	})
}

// Hook de shutdown
func onStop(params types.AppParams) func(context.Context) error {
	return func(ctx context.Context) error {

		// Shutdown do message queue
		if params.MQ != nil {
			if err := params.MQ.GracefulShutdown(ctx); err != nil {
				params.Logger.Error("failed to shutdown message queue", "error", err)
			} else {
				params.Logger.Info("message queue shutdown completed")
			}
		}

		// Shutdown do servidor web
		if params.Router != nil {
			if err := params.Router.Shutdown(ctx); err != nil {
				params.Logger.Error("failed to shutdown web server", "error", err)
			} else {
				params.Logger.Info("web server shutdown completed")
			}
		}

		return nil
	}
}
