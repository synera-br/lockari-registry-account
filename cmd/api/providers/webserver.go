package providers

import (
	"encoding/json"
	"fmt"
	"registry-account/cmd/api/types"
	"registry-account/pkg/webserver"
)

func ProvideWebServer(params types.WebServerParams) (webserver.ServerInterface, error) {
	// Lê a configuração do webserver do config
	b, err := params.Config.MarshalField("webserver")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal webserver config: %w", err)
	}

	var webserverConfig webserver.WebServerParams
	err = json.Unmarshal(b, &webserverConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal webserver config: %w", err)
	}

	// Aplica validação para definir defaults se necessário
	if err := webserverConfig.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate webserver config: %w", err)
	}

	srv, err := webserver.NewServer(&webserver.Config{
		Tracer:            params.Obs,
		TracerCleanupFunc: params.Cleanup,
		WebServerParams:   webserverConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create web server: %w", err)
	}

	err = srv.Initialize(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize web server: %w", err)
	}

	return srv, nil
}
