package serviceuser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"registry-account/internal/core/entity/entity_user"
	"registry-account/internal/infrastructure/messaging"
	"registry-account/pkg/logger"
	"registry-account/pkg/messagequeue"
	"registry-account/pkg/telemetry"
	"registry-account/pkg/utils"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracerName = "serviceUser"

type serviceUserParams struct {
	repo  entity_user.UserResponseData
	mq    messagequeue.MessageQueue
	cache interface{}
	obs   telemetry.OtelObservability
	log   logger.LoggerInterface
}

func NewServiceUser(
	repo entity_user.UserResponseData,
	mq messagequeue.MessageQueue,
	cache interface{},
	obs telemetry.OtelObservability,
	log logger.LoggerInterface,
) (entity_user.UserRequestData, error) {

	if repo == nil {
		return nil, utils.ServiceRegistryRequestInvalidRepository
	}

	if mq == nil {
		return nil, utils.ServiceRegistryRequestInvalidMQ
	}

	user := &serviceUserParams{
		repo:  repo,
		mq:    mq,
		cache: cache,
		obs:   obs,
		log:   log,
	}

	// go user.mqConsumer(context.Background())

	return user, nil
}

func (s *serviceUserParams) Create(ctx context.Context, user *entity_user.User) (*entity_user.UserResponse, error) {
	ctx, span := s.startSpan(ctx, "service user create")
	defer span.End()

	if user == nil {
		s.spanError(ctx, utils.ErrInvalidUser)
		return nil, utils.ErrInvalidUser
	}

	userResponse, err := s.create(ctx, user)
	if err != nil {
		s.spanError(ctx, err)
		return nil, err
	}

	if userResponse == nil {
		s.spanError(ctx, utils.ErrInvalidUser)
		return nil, utils.ErrInvalidUser
	}

	// data, err := json.Marshal(userResponse)
	// if err != nil {
	// 	s.spanError(ctx, err)
	// 	return nil, fmt.Errorf("service user create: failed to marshal userResponse: %w", err)
	// }

	// messageId := fmt.Sprintf("user-created-%s-%d", userResponse.ID, time.Now().UnixNano())
	// msg := amqp091.Publishing{
	// 	Headers: amqp091.Table{
	// 		"source":     "user-service",
	// 		"event_type": messaging.RoutingKeyUserCreateSecret.String(),
	// 		"created_at": time.Now().UTC().Format(time.RFC3339),
	// 		"version":    "1.0",
	// 		"user_id":    userResponse.ID,
	// 	},
	// 	ContentType:     "application/json",
	// 	ContentEncoding: "utf-8",
	// 	DeliveryMode:    amqp091.Persistent, // Persiste a mensagem
	// 	Priority:        0,
	// 	MessageId:       messageId,
	// 	Timestamp:       time.Now().UTC(),
	// 	Type:            messaging.RoutingKeyUserCreateSecret.String(),
	// 	AppId:           "user-service",
	// 	Body:            data,
	// }

	// if err := s.publishUserCreated(ctx, msg); err != nil {
	// 	s.spanError(ctx, fmt.Errorf("CRITICAL: failed to publish user.created event", "userID", userResponse.ID, "error", err))
	// 	s.log.Error("CRITICAL: failed to publish user.created event",
	// 		"userID", userResponse.ID,
	// 		"messageId", messageId,
	// 		"error", err,
	// 		"exchange", messaging.ExchangeUser.String(),
	// 		"routingKey", messaging.RoutingKeyUserCreateSecret.String(),
	// 	)
	// 	return nil, fmt.Errorf("failed to publish user.created event: %w", err)
	// }

	return userResponse, nil
}

func (s *serviceUserParams) create(ctx context.Context, user *entity_user.User) (*entity_user.UserResponse, error) {
	ctx, span := s.startSpan(ctx, "service user create")
	defer span.End()
	span.SetAttributes(attribute.String("method", "serviceCreateUser"))

	if user == nil {
		s.spanError(ctx, utils.ErrInvalidUser)
		return nil, utils.ErrInvalidUser
	}

	if err := user.Validate(); err != nil {
		s.spanError(ctx, err)
		return nil, fmt.Errorf("service user create: %w", err)
	}

	userResponse, err := s.repo.Create(ctx, user)
	if err != nil {
		s.spanError(ctx, err)
		return nil, fmt.Errorf("service user create: %w", err)
	}

	if userResponse == nil {
		s.spanError(ctx, utils.ErrInvalidUser)
		return nil, utils.ErrInvalidUser
	}

	if err := userResponse.Validate(); err != nil {
		s.spanError(ctx, err)
		return nil, fmt.Errorf("service user create: %w", err)
	}

	return userResponse, nil
}

func (s *serviceUserParams) publishUserCreated(ctx context.Context, msg amqp091.Publishing) error {
	ctx, span := s.startSpan(ctx, "service user publish user created")
	defer span.End()

	if s.mq == nil {
		s.spanError(ctx, fmt.Errorf("message queue is nil"))
		s.log.Error("message", "CRITICAL: Message queue is nil.", "error", fmt.Errorf("message queue is nil"))
		return fmt.Errorf("message queue is nil")
	}

	exchange := messaging.ExchangeUser.String()
	routingKey := messaging.RoutingKeyUserCreateSecret.String()

	span.SetAttributes(
		attribute.String("method", "servicePublishUserCreated"),
		attribute.String("exchange", exchange),
		attribute.String("routing_key", routingKey),
		attribute.String("message_id", msg.MessageId),
	)

	// Timeout para a publicação
	publishCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	span.SetAttributes(attribute.String("method", "servicePublishUserCreated"))
	if err := s.mq.Publish(publishCtx, exchange, routingKey, msg); err != nil {
		s.spanError(ctx, err)
		return fmt.Errorf("failed to publish to exchange '%s' with routing key '%s': %w",
			exchange, routingKey, err)
	}

	span.SetAttributes(attribute.String("status", "published_successfully"))
	return nil
}

func (s *serviceUserParams) Get(ctx context.Context, filter *entity_user.UserFilter) ([]*entity_user.UserResponse, error) {
	ctx, span := s.startSpan(ctx, "service user get")
	defer span.End()
	span.SetAttributes(attribute.String("method", "serviceGetUser"))

	if s == nil {
		s.spanError(ctx, fmt.Errorf("serviceUserParamsIsNil"))
		return nil, fmt.Errorf("serviceUserParams is nil")
	}

	if filter == nil {
		s.spanError(ctx, fmt.Errorf("filterIsNil"))
		return nil, fmt.Errorf("filter is nil")
	}

	if err := filter.Validate(); err != nil {
		s.spanError(ctx, err)
		return nil, fmt.Errorf("filter is invalid: %w", err)
	}

	result, err := s.repo.Get(ctx, filter)
	if err != nil {
		s.spanError(ctx, err)
		return nil, err
	}

	if result == nil {
		s.spanError(ctx, fmt.Errorf("resultIsNil"))
		return nil, utils.UserNotFound
	}

	return result, nil
}

func (s *serviceUserParams) startSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if s.obs == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return s.obs.Span(ctx, operationName)
}

func (s *serviceUserParams) spanError(ctx context.Context, err error) {
	span := s.obs.Trace(ctx)
	span.RecordError(err)
	span.SetAttributes(attribute.String("error", err.Error()))
}

func (s *serviceUserParams) mqConsumer(ctx context.Context) error {
	ctx, span := s.startSpan(ctx, "serviceUserMQConsumer")
	defer span.End()
	span.SetAttributes(attribute.String("method", "serviceUserMQConsumer"))

	if s == nil || s.mq == nil {
		return fmt.Errorf("serviceUserParams or MessageQueue is nil")
	}
	tag := fmt.Sprintf("tenant-user-%s-%d", os.Getenv("HOSTNAME"), os.Getpid())
	if tag == "tenant-user--0" {
		tag = fmt.Sprintf("tenant-user-%d", time.Now().UTC().UnixNano())
	}

	var reconnectDelay = 5 * time.Second
	const maxReconnectDelay = 5 * time.Minute

	for {
		select {
		case <-ctx.Done():
			span.AddEvent("Context cancelled. Shutting down the CreateFromMQ consumer.")
			s.mq.CancelConsumer(tag)
		default:
			s.mq.CancelConsumer(tag)

			err := s.mq.Consume(ctx, messaging.QueueUserCreateSecret.String(), false, tag, func(delivery amqp091.Delivery) error {
				msgCtx, span := s.obs.Span(ctx, "HandleUserCreationEvent")
				defer span.End()

				span.SetAttributes(
					attribute.String("message_id", delivery.MessageId),
					attribute.String("routing_key", delivery.RoutingKey),
				)

				// 1. Tenta fazer o unmarshal da mensagem
				var userPayload entity_user.User // ou um DTO específico para o evento
				if err := json.Unmarshal(delivery.Body, &userPayload); err != nil {
					s.spanError(msgCtx, err)
					s.log.Error("Falha no unmarshal da mensagem, descartando.",
						"error", err,
						"routing_key", delivery.RoutingKey,
						"exchange", delivery.Exchange,
					)
					return delivery.Nack(false, false)
				}

				// 2. Valida os dados recebidos
				if err := userPayload.Validate(); err != nil {
					s.spanError(msgCtx, err)

					// Erro permanente. Rejeita e não devolve para a fila.
					return delivery.Nack(false, false)
				}

				// 3. Executa a lógica de negócio (chama o Create)
				_, err := s.create(msgCtx, &userPayload)
				if err != nil {
					s.spanError(msgCtx, err)

					return delivery.Nack(false, true)
				}

				// 4. Sucesso!
				span.SetAttributes(attribute.String("message", "Mensagem processada com sucesso!"), attribute.String("userID", userPayload.UID))

				// Confirma o processamento (Ack). A mensagem será removida da fila.
				return delivery.Ack(false)
			})

			if err != nil {
				s.spanError(ctx, fmt.Errorf("consumer returned with an error: %w", err))
				s.log.Error("Consumer returned with an error. Attempting to reconnect...",
					"error", err,
				)
				time.Sleep(reconnectDelay)

				// Aumenta o delay para a próxima tentativa
				reconnectDelay *= 2
				if reconnectDelay > maxReconnectDelay {
					reconnectDelay = maxReconnectDelay
				}
			} else {
				reconnectDelay = 5 * time.Second
			}
		}
	}
}
