package providers

import (
	"encoding/json"
	"fmt"
	"registry-account/cmd/api/types"
	"registry-account/pkg/telemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func ProvideTelemetry(params types.TelemetryParams) (telemetry.OtelObservability, func(), error) {
	b, err := params.Config.MarshalField("telemetry")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal telemetry config: %w", err)
	}

	var config telemetry.OtelConfig
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal telemetry config: %w", err)
	}

	obs, cleanup, err := telemetry.InitObservability(&config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	return obs, cleanup, nil
}

func ProvidePropagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}
