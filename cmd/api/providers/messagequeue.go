package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"registry-account/cmd/api/types"
	"registry-account/pkg/messagequeue"
)

func ProvideMessageQueue(params types.MessageQueueParams) (messagequeue.MessageQueue, error) {
	b, err := json.Marshal(params.Config.Settings["messagequeue"])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message queue config: %w", err)
	}

	var mqCfg messagequeue.FullConfig
	err = json.Unmarshal(b, &mqCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message queue config: %w", err)
	}

	if err := mqCfg.RabbitMQ.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message queue config: %w", err)
	}

	mqTracer, err := params.Obs.Tracer()
	if err != nil {
		return nil, fmt.Errorf("failed to get tracer: %w", err)
	}

	client := messagequeue.NewClient(mqTracer, params.Propagator)
	if client == nil {
		return nil, fmt.Errorf("failed to create message queue client")
	}

	err = client.Initialize(&mqCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize message queue client: %w", err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to message queue: %w", err)
	}

	return client, nil
}
