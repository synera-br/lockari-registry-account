package providers

import (
	"fmt"
	"registry-account/configs"
)

func ProvideConfig() (*configs.AppConfig, error) {
	cfg, err := configs.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}
