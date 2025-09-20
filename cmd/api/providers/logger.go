package providers

import (
	"encoding/json"
	"fmt"
	"registry-account/cmd/api/types"
	"registry-account/pkg/logger"
)

func ProvideLogger(params types.LoggerParams) (logger.LoggerInterface, error) {
	b, err := params.Config.MarshalField("logger")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal logger config: %w", err)
	}

	var config logger.Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal logger config: %w", err)
	}

	log := logger.NewLogger()
	err = log.Initialize(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return log, nil
}
