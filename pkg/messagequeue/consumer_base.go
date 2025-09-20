package messagequeue

import (
	"context"
	"strings"

	"registry-account/pkg/logger"
	"registry-account/pkg/telemetry"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ErrorType classifica tipos de erros para tratamento adequado
type ErrorType int

const (
	ErrorTypeTemporary ErrorType = iota
	ErrorTypePermanent
	ErrorTypeConnection
)

// MessageHandler define a interface que cada service deve implementar
type MessageHandler interface {
	HandleMessage(ctx context.Context, delivery amqp091.Delivery) error
}

// ConsumerBase fornece funcionalidade comum para todos os consumers
type ConsumerBase struct {
	mq     MessageQueue
	log    logger.LoggerInterface
	obs    telemetry.OtelObservability
	tracer string
}

// NewConsumerBase cria uma nova instância base para consumers
func NewConsumerBase(
	mq MessageQueue,
	log logger.LoggerInterface,
	obs telemetry.OtelObservability,
	tracerName string,
) *ConsumerBase {
	return &ConsumerBase{
		mq:     mq,
		log:    log,
		obs:    obs,
		tracer: tracerName,
	}
}

// ClassifyError classifica erros para tratamento adequado
func (cb *ConsumerBase) ClassifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeTemporary
	}

	errStr := err.Error()

	// Connection errors
	if strings.Contains(errStr, "channel/connection is not open") ||
		strings.Contains(errStr, "unknown delivery tag") ||
		strings.Contains(errStr, "PRECONDITION_FAILED") {
		return ErrorTypeConnection
	}

	// Permanent errors
	if strings.Contains(errStr, "E11000 duplicate key error") ||
		strings.Contains(errStr, "validation") ||
		strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "bad request") {
		return ErrorTypePermanent
	}

	// Default: temporary error
	return ErrorTypeTemporary
}

// HandleMessageError processa erros de mensagem de forma inteligente
func (cb *ConsumerBase) HandleMessageError(delivery amqp091.Delivery, err error) {
	errorType := cb.ClassifyError(err)

	switch errorType {
	case ErrorTypeConnection:
		cb.log.Error("Connection error - will reconnect",
			"error", err,
			"message_id", delivery.MessageId,
			"routing_key", delivery.RoutingKey,
		)
		// Não faz NACK - deixa timeout para reconectar
		return

	case ErrorTypePermanent:
		cb.log.Error("Permanent error - discarding message",
			"error", err,
			"message_id", delivery.MessageId,
			"routing_key", delivery.RoutingKey,
		)
		cb.safeNack(delivery, false, false) // Discard

	case ErrorTypeTemporary:
		cb.log.Warn("Temporary error - requeuing message",
			"error", err,
			"message_id", delivery.MessageId,
			"routing_key", delivery.RoutingKey,
		)
		cb.safeNack(delivery, false, true) // Requeue
	}
}

// SafeNack faz NACK com error handling seguro
func (cb *ConsumerBase) safeNack(delivery amqp091.Delivery, multiple bool, requeue bool) {
	if err := delivery.Nack(multiple, requeue); err != nil {
		cb.log.Error("CRITICAL: Failed to Nack message",
			"error", err,
			"message_id", delivery.MessageId,
			"requeue", requeue,
		)
	}
}

// SafeAck faz ACK com error handling seguro
func (cb *ConsumerBase) SafeAck(delivery amqp091.Delivery) {
	if err := delivery.Ack(false); err != nil {
		cb.log.Error("Failed to Ack message. Connection likely lost.",
			"error", err,
			"message_id", delivery.MessageId,
		)
	}
}

// StartSpan inicia um span de tracing
func (cb *ConsumerBase) StartSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if cb.obs == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return cb.obs.Span(ctx, operationName)
}

// SpanError registra erro no span
func (cb *ConsumerBase) SpanError(ctx context.Context, err error) {
	if cb.obs == nil {
		return
	}
	span := cb.obs.Trace(ctx)
	span.RecordError(err)
	span.SetAttributes(attribute.String("error", err.Error()))
}
