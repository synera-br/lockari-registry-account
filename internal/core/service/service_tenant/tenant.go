package servicetenant

import (
	"context"
	"fmt"
	entitytenant "registry-account/internal/core/entity/entity_tenant"
	"registry-account/internal/infrastructure/messaging"
	"registry-account/pkg/logger"
	"registry-account/pkg/messagequeue"
	"registry-account/pkg/telemetry"
	"registry-account/pkg/utils"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracerName = "serviceTenant"

type serviceTenantParams struct {
	repo  entitytenant.TenantResponseData
	mq    messagequeue.MessageQueue
	cache interface{}
	obs   telemetry.OtelObservability
	log   logger.LoggerInterface
}

func NewServiceTenant(
	repo entitytenant.TenantResponseData,
	mq messagequeue.MessageQueue,
	cache interface{},
	obs telemetry.OtelObservability,
	log logger.LoggerInterface,
) (entitytenant.TenantRequestData, error) {
	tenant := &serviceTenantParams{
		repo:  repo,
		mq:    mq,
		cache: cache,
		obs:   obs,
		log:   log,
	}

	return tenant, nil
}

func (s *serviceTenantParams) Create(ctx context.Context, tenant *entitytenant.Tenant) (*entitytenant.TenantResponse, error) {
	ctx, span := s.startSpan(ctx, "serviceTenantCreate")
	defer span.End()
	span.SetAttributes(attribute.String("method", "serviceTenantCreate"))

	if tenant == nil {
		s.spanError(ctx, utils.ErrInvalidTenant)
		return nil, utils.ErrInvalidTenant
	}

	if err := tenant.Validate(); err != nil {
		s.spanError(ctx, err)
		return nil, fmt.Errorf("service tenant create: %w", err)
	}

	tenantResponse, err := s.repo.Create(ctx, tenant)
	if err != nil {
		s.spanError(ctx, err)
		return nil, fmt.Errorf("service tenant create: %w", err)
	}

	if err := tenantResponse.Validate(); err != nil {
		s.spanError(ctx, err)
		return nil, fmt.Errorf("service tenant create: %w", err)
	}

	// data, err := json.Marshal(tenantResponse)
	// if err != nil {
	// 	s.spanError(ctx, err)
	// 	return nil, fmt.Errorf("service tenant create: failed to marshal tenantResponse: %w", err)
	// }

	// messageId := fmt.Sprintf("tenant-created-%s-%d", tenantResponse.ID, time.Now().UnixNano())
	// msg := amqp091.Publishing{
	// 	Headers: amqp091.Table{
	// 		"source":     "tenant-service",
	// 		"event_type": messaging.RoutingKeyTenantCreateSecret.String(),
	// 		"created_at": time.Now().UTC().Format(time.RFC3339),
	// 		"version":    "1.0",
	// 		"tenant_id":  tenantResponse.ID,
	// 	},
	// 	ContentType:     "application/json",
	// 	ContentEncoding: "utf-8",
	// 	DeliveryMode:    amqp091.Persistent, // Persiste a mensagem
	// 	Priority:        0,
	// 	MessageId:       messageId,
	// 	Timestamp:       time.Now().UTC(),
	// 	Type:            messaging.RoutingKeyTenantCreateSecret.String(),
	// 	AppId:           "tenant-service",
	// 	Body:            data,
	// }

	// if err := s.publishCreated(ctx, msg); err != nil {
	// 	s.spanError(ctx, fmt.Errorf("CRITICAL: failed to publish tenant.created event", "tenantID", tenantResponse.ID, "error", err))
	// 	s.log.Error("CRITICAL: failed to publish tenant.created event",
	// 		"tenantID", tenantResponse.ID,
	// 		"messageId", messageId,
	// 		"error", err,
	// 		"exchange", messaging.ExchangeTenant.String(),
	// 		"routingKey", messaging.RoutingKeyTenantCreateSecret.String(),
	// 	)
	// 	return nil, fmt.Errorf("failed to publish tenant.created event: %w", err)
	// }

	return tenantResponse, nil
}

func (s *serviceTenantParams) publishCreated(ctx context.Context, msg amqp091.Publishing) error {
	ctx, span := s.startSpan(ctx, "service tenant publish tenant created")
	defer span.End()

	if s.mq == nil {
		s.spanError(ctx, fmt.Errorf("message queue is nil"))
		s.log.Error("message", "CRITICAL: Message queue is nil.", "error", fmt.Errorf("message queue is nil"))
		return fmt.Errorf("message queue is nil")
	}

	exchange := messaging.ExchangeTenant.String()
	routingKey := messaging.RoutingKeyTenantCreateSecret.String()

	span.SetAttributes(
		attribute.String("method", "servicePublishTenantCreated"),
		attribute.String("exchange", exchange),
		attribute.String("routing_key", routingKey),
		attribute.String("message_id", msg.MessageId),
	)

	// Timeout para a publicação
	publishCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	span.SetAttributes(attribute.String("method", "servicePublishTenantCreated"))
	if err := s.mq.Publish(publishCtx, exchange, routingKey, msg); err != nil {
		s.spanError(ctx, err)
		return fmt.Errorf("failed to publish to exchange '%s' with routing key '%s': %w",
			exchange, routingKey, err)
	}

	span.SetAttributes(attribute.String("status", "published_successfully"))
	return nil
}

func (s *serviceTenantParams) Get(ctx context.Context, filter *entitytenant.TenantFilter) ([]*entitytenant.TenantResponse, error) {
	ctx, span := s.startSpan(ctx, "serviceTenantGet")
	defer span.End()
	span.SetAttributes(attribute.String("method", "serviceTenantGet"))

	if err := filter.Validate(); err != nil {
		return nil, err
	}

	results, err := s.repo.Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, utils.ErrTenantNotFound
	}

	return results, nil
}

func (s *serviceTenantParams) startSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if s.obs == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return s.obs.Span(ctx, operationName)
}

func (s *serviceTenantParams) spanError(ctx context.Context, err error) {
	span := s.obs.Trace(ctx)
	span.RecordError(err)
	span.SetAttributes(attribute.String("error", err.Error()))
}
