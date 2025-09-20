// cmd/app/main.go
package main

import (
	"context"
	"fmt"
	"registry-account/cmd/api/hooks"
	"registry-account/cmd/api/providers"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	app := fx.New(
		// Logger customizado para Fx
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),

		// Módulo principal com todos os providers
		providers.Module,

		// Hooks de inicialização
		fx.Invoke(
			hooks.SetupHandlers,
			hooks.LogAppStart,
			hooks.StartWebServer,
		),
	)

	// Configurar timeout de inicialização
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Iniciar aplicação
	if err := app.Start(ctx); err != nil {
		panic(fmt.Sprintf("failed to start application: %v", err))
	}

	// Aguardar sinal de parada
	<-app.Wait()

	// Parar aplicação gracefully
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer stopCancel()

	if err := app.Stop(stopCtx); err != nil {
		panic(fmt.Sprintf("failed to stop application: %v", err))
	}
}
