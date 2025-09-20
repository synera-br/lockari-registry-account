package messagequeue

import (
	"context"
	"fmt"
	"os"
	"time"

	"registry-account/pkg/logger"
	"registry-account/pkg/telemetry"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
)

// ConsumerConfig configuração para o consumer genérico
type ConsumerConfig struct {
	QueueName         string
	ConsumerTag       string
	MaxReconnectDelay time.Duration
	InitialDelay      time.Duration
	AutoAck           bool
}

// GenericConsumer consumer genérico reutilizável
type GenericConsumer struct {
	*ConsumerBase
	config  ConsumerConfig
	handler MessageHandler
}

// NewGenericConsumer cria um novo consumer genérico
func NewGenericConsumer(
	mq MessageQueue,
	log logger.LoggerInterface,
	obs telemetry.OtelObservability,
	tracerName string,
	config ConsumerConfig,
	handler MessageHandler,
) *GenericConsumer {
	return &GenericConsumer{
		ConsumerBase: NewConsumerBase(mq, log, obs, tracerName),
		config:       config,
		handler:      handler,
	}
}

// Start inicia o consumer com reconexão automática
func (gc *GenericConsumer) Start(ctx context.Context) error {
	ctx, span := gc.StartSpan(ctx, fmt.Sprintf("%s_consumer", gc.tracer))
	defer span.End()
	span.SetAttributes(attribute.String("method", "genericConsumerStart"))

	if gc.mq == nil {
		return fmt.Errorf("message queue is nil")
	}

	// Generate unique consumer tag
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}

	var reconnectDelay = gc.config.InitialDelay
	if reconnectDelay == 0 {
		reconnectDelay = 5 * time.Second
	}

	maxReconnectDelay := gc.config.MaxReconnectDelay
	if maxReconnectDelay == 0 {
		maxReconnectDelay = 5 * time.Minute
	}

	// Consumer principal com controle de reconexão
	for {
		// Check cancellation antes de tentar consumir
		select {
		case <-ctx.Done():
			span.AddEvent("Context cancelled. Shutting down consumer.")
			gc.mq.CancelConsumer(gc.config.ConsumerTag)
			return ctx.Err()
		default:
		}

		// Generate NEW unique tag for each reconnection attempt
		tag := fmt.Sprintf("%s-%s-%d-%d",
			gc.config.ConsumerTag,
			hostname,
			os.Getpid(),
			time.Now().UTC().UnixNano())

		if len(tag) > 255 {
			tag = fmt.Sprintf("%s-%d", gc.config.ConsumerTag, time.Now().UTC().UnixNano())
		}

		// Cancel any previous consumer (ignore errors - may not exist)
		gc.mq.CancelConsumer(tag)

		// Small delay for cleanup
		time.Sleep(100 * time.Millisecond)

		err := gc.consumeMessages(ctx, tag)

		if err != nil {
			gc.SpanError(ctx, fmt.Errorf("consumer returned with an error: %w", err))

			errorType := gc.ClassifyError(err)
			switch errorType {
			case ErrorTypeConnection:
				gc.log.Error("Consumer connection error. Attempting to reconnect...",
					"error", err,
					"tag", tag,
				)
			default:
				gc.log.Warn("Consumer error, retrying...",
					"error", err,
					"tag", tag,
				)
			}

			time.Sleep(reconnectDelay)
			reconnectDelay *= 2
			if reconnectDelay > maxReconnectDelay {
				reconnectDelay = maxReconnectDelay
			}
		} else {
			reconnectDelay = gc.config.InitialDelay
		}
	}
}

// consumeMessages consome mensagens usando o handler fornecido
func (gc *GenericConsumer) consumeMessages(ctx context.Context, tag string) error {
	return gc.mq.Consume(ctx, gc.config.QueueName, gc.config.AutoAck, tag,
		func(delivery amqp091.Delivery) error {
			msgCtx, span := gc.StartSpan(ctx, fmt.Sprintf("%s_handle_message", gc.tracer))
			defer span.End()

			span.SetAttributes(
				attribute.String("message_id", delivery.MessageId),
				attribute.String("routing_key", delivery.RoutingKey),
			)

			// Chama o handler específico do service
			if err := gc.handler.HandleMessage(msgCtx, delivery); err != nil {
				gc.SpanError(msgCtx, err)
				gc.HandleMessageError(delivery, err)
				return nil // Não retorna erro para evitar reconnect desnecessário
			}

			// Success - ACK the message
			span.SetAttributes(
				attribute.String("message", "Message processed successfully!"),
				attribute.String("messageID", delivery.MessageId),
			)

			gc.SafeAck(delivery)
			return nil
		})
}
