package service_registry_request

import (
	"context"
	"encoding/json"
	"fmt"
	"registry-account/internal/core/entity/entity_registry_request"
	entitytenant "registry-account/internal/core/entity/entity_tenant"
	"registry-account/internal/core/entity/entity_user"
	"registry-account/internal/infrastructure/messaging"
	"registry-account/pkg/authclient"
	"registry-account/pkg/logger"
	"registry-account/pkg/messagequeue"
	"registry-account/pkg/telemetry"
	"registry-account/pkg/utils"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type registryRequestParams struct {
	auth   authclient.AuthenticationProvider
	mq     messagequeue.MessageQueue
	cache  interface{}
	obs    telemetry.OtelObservability
	log    logger.LoggerInterface
	user   entity_user.UserRequestData
	tenant entitytenant.TenantRequestData
}

func NewServiceRequest(
	auth authclient.AuthenticationProvider,
	svcUser entity_user.UserRequestData,
	svcTenant entitytenant.TenantRequestData,
	mq messagequeue.MessageQueue,
	cache interface{},
	obs telemetry.OtelObservability,
	log logger.LoggerInterface,
) (entity_registry_request.ServiceRegistryRequest, error) {

	if auth == nil {
		return nil, utils.ServiceRegistryRequestInvalidAuthClient
	}

	if mq == nil {
		return nil, utils.ServiceRegistryRequestInvalidMQ
	}

	if svcUser == nil {
		return nil, utils.ServiceRegistryRequestInvalidUser
	}

	if svcTenant == nil {
		return nil, utils.ServiceRegistryRequestInvalidTenant
	}

	params := registryRequestParams{
		auth:   auth,
		mq:     mq,
		cache:  cache,
		obs:    obs,
		log:    log,
		user:   svcUser,
		tenant: svcTenant,
	}

	return &params, nil
}

func (s *registryRequestParams) RegistryNewAccount(ctx context.Context, user *entity_registry_request.RegistryRequest) (*entity_registry_request.RegistryResponse, error) {
	ctx, span := s.startSpan(ctx, "ServiceRegistryNewAccount")
	defer span.End()

	// 1. initial check
	if user == nil {
		s.spanError(ctx, utils.ErrInvalidUser)
		return nil, utils.ErrInvalidUser
	}

	// 2. ensure that the user exists in the authentication provider
	userInAuthProvider, wasCreatedInAuthProvider, err := s.ensureUserInAuthProvider(ctx, user)
	if err != nil {
		s.spanError(ctx, err)
		return nil, err
	}

	// 3. ensure that the tenant exists
	tenant, err := s.ensureTenant(ctx, &user.Tenant, &userInAuthProvider.Email)
	if err != nil {
		if err != utils.ErrTenantNotFound {
			return nil, err
		}
	}

	// 4. Update the user in the authentication provider with the tenant
	u, err := s.setTenantInAuthProvider(ctx, &userInAuthProvider.UID, &tenant.ID)
	if err != nil {
		s.spanError(ctx, err)
		return nil, err
	}

	if u != nil {
		if u.UID == userInAuthProvider.UID {
			userInAuthProvider.CustomClaims = u.CustomClaims
		}
	}

	// 5. Ensures the user exists in the local database
	userInDatabase, wasCreatedInDB, err := s.ensureUserInDatabase(ctx, user, userInAuthProvider)
	if err != nil {
		if err != utils.UserNotFound {
			s.spanError(ctx, err)
			return nil, err
		}
	}

	account := &registryAccountParams{
		User:   *userInDatabase,
		Tenant: *tenant,
	}

	err = s.prepareToPublish(ctx, account)
	if err != nil {
		s.spanError(ctx, err)
		return nil, err
	}

	response := entity_registry_request.RegistryResponse{
		Success: true,
		Message: "User created successfully",
		User: entity_registry_request.UserData{
			Email:       userInDatabase.Email,
			DisplayName: userInAuthProvider.DisplayName,
			Tenant:      userInDatabase.Tenants,
			UID:         userInDatabase.UID,
		},
	}

	switch {
	case wasCreatedInAuthProvider && wasCreatedInDB:
		response.Message = "User created successfully"
	case wasCreatedInAuthProvider && !wasCreatedInDB:
		response.Message = "User created in auth provider, already exists in database"
	case !wasCreatedInAuthProvider && wasCreatedInDB:
		response.Message = "User exists in auth provider, created in database"
	case !wasCreatedInAuthProvider && !wasCreatedInDB:
		response.Message = "User already exists in both systems"
	default:
		response.Message = "User registration completed"
	}

	return &response, nil
}

func (s *registryRequestParams) prepareToPublish(ctx context.Context, account *registryAccountParams) error {

	if account == nil {
		return fmt.Errorf("service user create: account is nil")
	}

	if account.User.ID == "" {
		return fmt.Errorf("service user create: user is nil")
	}

	if account.Tenant.ID == "" {
		return fmt.Errorf("service user create: tenant is nil")
	}

	data, err := json.Marshal(*account)
	if err != nil {
		s.spanError(ctx, err)
		return fmt.Errorf("service user create: failed to marshal userResponse: %w", err)
	}

	messageId := fmt.Sprintf("user-created-%s-%d", account.Tenant.ID, time.Now().UnixNano())
	msg := amqp091.Publishing{
		Headers: amqp091.Table{
			"source":     "registry-account-service",
			"event_type": messaging.RoutingKeyAccountCreate.String(),
			"created_at": time.Now().UTC().Format(time.RFC3339),
			"version":    "1.0",
			"user_id":    account.User.ID,
			"tenant_id":  account.Tenant.ID,
		},
		ContentType:     "application/json",
		ContentEncoding: "utf-8",
		DeliveryMode:    amqp091.Persistent, // Persiste a mensagem
		Priority:        0,
		MessageId:       messageId,
		Timestamp:       time.Now().UTC(),
		Type:            messaging.RoutingKeyAccountCreate.String(),
		AppId:           "registry-account-service",
		Body:            data,
	}

	if err := s.publishRegistryAccount(ctx, msg); err != nil {
		s.spanError(ctx, fmt.Errorf("CRITICAL: failed to publish user.created event for tenantID %s: %w", account.Tenant.ID, err))
		s.log.Error("CRITICAL: failed to publish user.created event",
			"userID", account.User.ID,
			"tenantID", account.Tenant.ID,
			"messageId", messageId,
			"error", err,
			"exchange", messaging.ExchangeRegistryAccount.String(),
			"routingKey", messaging.RoutingKeyAccountCreate.String(),
		)
		return fmt.Errorf("failed to publish user.created event: %w", err)
	}
	return nil

}

func (s *registryRequestParams) publishRegistryAccount(ctx context.Context, msg amqp091.Publishing) error {
	ctx, span := s.startSpan(ctx, "service user publish user created")
	defer span.End()

	if s.mq == nil {
		s.spanError(ctx, fmt.Errorf("message queue is nil"))
		s.log.Error("message", "CRITICAL: Message queue is nil.", "error", fmt.Errorf("message queue is nil"))
		return fmt.Errorf("message queue is nil")
	}

	exchange := messaging.ExchangeRegistryAccount.String()
	routingKey := messaging.RoutingKeyAccountCreate.String()

	span.SetAttributes(
		attribute.String("method", "servicePublishAccountCreated"),
		attribute.String("exchange", exchange),
		attribute.String("routing_key", routingKey),
		attribute.String("message_id", msg.MessageId),
	)

	// Timeout para a publicação
	publishCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	span.SetAttributes(attribute.String("method", "servicePublishAccountCreated"))
	if err := s.mq.Publish(publishCtx, exchange, routingKey, msg); err != nil {
		s.spanError(ctx, err)
		return fmt.Errorf("failed to publish to exchange '%s' with routing key '%s': %w",
			exchange, routingKey, err)
	}

	span.SetAttributes(attribute.String("status", "published_successfully"))
	return nil
}

func (s *registryRequestParams) ensureTenant(ctx context.Context, userTenant, email *string) (*entitytenant.TenantResponse, error) {
	ctx, span := s.startSpan(ctx, "ServiceEnsureTenant")
	defer span.End()

	if userTenant == nil {
		return nil, utils.ErrInvalidTenant
	}

	tenant, err := s.getTenant(ctx, userTenant, email)
	if err != nil {
		if err != utils.ErrTenantNotFound {
			return nil, err
		}
	}

	if tenant == nil {
		tenant, err = s.createTenant(ctx, &entitytenant.Tenant{
			Name:  *userTenant,
			Owner: *email,
		})
		if err != nil {
			return nil, err
		}

		if tenant == nil {
			return nil, utils.ErrCannotCreateTenant
		}
	}

	return tenant, nil
}

func (s *registryRequestParams) ensureUserInDatabase(ctx context.Context, user *entity_registry_request.RegistryRequest, userInAuthProvider *auth.UserRecord) (*entity_user.UserResponse, bool, error) {
	ctx, span := s.startSpan(ctx, "ServiceEnsureUserInDatabase")
	defer span.End()

	wasCreated := false

	if user == nil {
		return nil, wasCreated, utils.ErrInvalidUser
	}

	if userInAuthProvider == nil {
		return nil, wasCreated, utils.ErrInvalidUser
	}

	if user.Email != userInAuthProvider.Email {
		return nil, wasCreated, utils.ErrInvalidUser
	}

	userFromDB, err := s.getUserInDB(ctx, &user.Email)
	if err != nil {
		if err != utils.UserNotFound {
			return nil, wasCreated, err
		}
	}

	if userFromDB != nil && userFromDB.Email == user.Email {
		return userFromDB, wasCreated, nil
	}

	userFromDB, err = s.createUserInDB(ctx, userInAuthProvider)
	if err != nil {
		return nil, wasCreated, err
	}

	if userFromDB == nil {
		return nil, wasCreated, utils.ErrInvalidUser
	}

	wasCreated = true

	return userFromDB, wasCreated, nil
}

func (s *registryRequestParams) createUserInDB(ctx context.Context, user *auth.UserRecord) (*entity_user.UserResponse, error) {
	ctx, span := s.startSpan(ctx, "ServiceCreateUserInDB")
	defer span.End()

	if user == nil {
		return nil, utils.ErrInvalidUser
	}

	tenantList := make([]string, 0)
	existingTenants, ok := user.CustomClaims["tenants"]
	if ok {
		if et, ok := existingTenants.([]interface{}); ok {
			for _, t := range et {
				if ts, ok := t.(string); ok {
					tenantList = append(tenantList, ts)
				}
			}
		}
	}

	userInDB, err := s.user.Create(ctx, &entity_user.User{
		UID:           user.UID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Claims:        user.CustomClaims,
		EmailVerified: user.EmailVerified,
		Tenants:       tenantList,
	})
	if err != nil {
		return nil, err
	}

	return userInDB, nil
}

func (s *registryRequestParams) getUserInDB(ctx context.Context, email *string) (*entity_user.UserResponse, error) {
	ctx, span := s.startSpan(ctx, "ServiceGetUserInDB")
	defer span.End()

	if email == nil {
		return nil, utils.ErrInvalidUser
	}

	results, err := s.user.Get(ctx, &entity_user.UserFilter{
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, utils.UserNotFound
	}

	return results[0], nil
}

func (s *registryRequestParams) getTenant(ctx context.Context, name, owner *string) (*entitytenant.TenantResponse, error) {
	ctx, span := s.startSpan(ctx, "ServiceGetTenant")
	defer span.End()

	if name == nil {
		return nil, utils.ErrInvalidTenant
	}

	if owner == nil {
		return nil, utils.ErrInvalidUserEmail
	}

	if s.tenant == nil {
		return nil, utils.ServiceRegistryRequestInvalidAuthenticationProvider
	}

	results, err := s.tenant.Get(ctx, &entitytenant.TenantFilter{
		Name:  name,
		Owner: owner,
	})
	if err != nil {
		if err != utils.ErrTenantNotFound {
			return nil, err
		}
	}

	if len(results) == 0 {
		return nil, utils.ErrTenantNotFound
	}

	return results[0], nil
}

func (s *registryRequestParams) createTenant(ctx context.Context, tenant *entitytenant.Tenant) (*entitytenant.TenantResponse, error) {
	ctx, span := s.startSpan(ctx, "ServiceCreateTenant")
	defer span.End()

	if tenant == nil {
		return nil, utils.ErrInvalidTenant
	}

	if s.tenant == nil {
		return nil, utils.ServiceRegistryRequestInvalidTenant
	}

	if err := tenant.Validate(); err != nil {
		return nil, err
	}

	result, err := s.tenant.Create(ctx, tenant)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, utils.ErrCannotCreateTenant
	}

	return result, nil
}

func (s *registryRequestParams) ensureUserInAuthProvider(ctx context.Context, user *entity_registry_request.RegistryRequest) (*auth.UserRecord, bool, error) {
	ctx, span := s.startSpan(ctx, "ServiceEnsureUserInAuthProvider")
	defer span.End()

	wasCreated := false
	if user == nil {
		return nil, wasCreated, utils.ErrInvalidUser
	}

	// Check if user exists in authentication provider
	userInAuthProvider, err := s.getUserInAuthProvider(ctx, user)
	if err != nil {
		s.spanError(ctx, err)
		if err != utils.UserNotFound {
			return nil, wasCreated, err
		}
	}

	startTime := 1.0
	if userInAuthProvider == nil {
		s.log.Warn("User not found in auth provider, starting retry logic", "email", user.Email)

		for i := 0; i < 3; i++ {
			fmt.Printf("Retrying to get user in auth provider, attempt: %d, email: %s\n", i+1, user.Email)
			s.log.Info("Retrying to get user in auth provider", "attempt", i+1, "email", user.Email, "wait_time", startTime)

			time.Sleep(time.Duration(startTime) * time.Second)

			retryUser, retryErr := s.getUserInAuthProvider(ctx, user)
			if retryErr != nil {
				s.log.Error("Retry attempt failed", "attempt", i+1, "email", user.Email, "error", retryErr.Error())
				if retryErr != utils.UserNotFound {
					return nil, wasCreated, retryErr
				}
			}

			if retryUser != nil {
				s.log.Info("User found on retry", "attempt", i+1, "email", user.Email, "uid", retryUser.UID)
				return retryUser, wasCreated, nil
			}

			startTime += 1.5
		}

		s.log.Error("User not found after all retry attempts", "email", user.Email, "attempts", 3)
		return nil, wasCreated, utils.UserNotFound

		// userInAuthProvider, err = s.createUserInAuthProvider(ctx, user)
		// if err != nil {
		// 	s.spanError(ctx, err)
		// 	return nil, wasCreated, err
		// }

		// if userInAuthProvider == nil {
		// 	s.spanError(ctx, utils.UserErrorToCreate)
		// 	return nil, wasCreated, utils.UserErrorToCreate
		// }
		// wasCreated = true
	}

	return userInAuthProvider, wasCreated, nil
}

func (s *registryRequestParams) createUserInAuthProvider(ctx context.Context, user *entity_registry_request.RegistryRequest) (*auth.UserRecord, error) {
	ctx, span := s.startSpan(ctx, "ServiceCreateUserInAuthProvider")
	defer span.End()

	if user == nil {
		return nil, utils.ErrInvalidUser
	}

	result, err := s.auth.CreateUser(ctx, user.Email, user.Password, user.Name)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, utils.ErrInvalidUser
	}

	return result, nil
}

func (s *registryRequestParams) getUserInAuthProvider(ctx context.Context, user *entity_registry_request.RegistryRequest) (*auth.UserRecord, error) {
	ctx, span := s.startSpan(ctx, "ServiceGetUserInAuthProvider")
	defer span.End()

	if user == nil {
		s.log.Error("getUserInAuthProvider: user is nil")
		return nil, utils.ErrInvalidUser
	}

	if s.auth == nil {
		s.log.Error("getUserInAuthProvider: auth client is nil")
		return nil, utils.ServiceRegistryRequestInvalidAuthenticationProvider
	}

	s.log.Info("Attempting to get user from Firebase Auth", "email", user.Email)

	userAuth, err := s.auth.GetUserByEmail(ctx, user.Email)
	if err != nil {
		s.log.Error("Firebase Auth GetUserByEmail failed", "email", user.Email, "error", err.Error())
		if strings.Contains(err.Error(), "no user exists") {
			return nil, utils.UserNotFound
		}
		if strings.Contains(err.Error(), "user not found") {
			return nil, utils.UserNotFound
		}
		// Return the original error for other Firebase errors (config, network, etc.)
		return nil, err
	}

	if userAuth == nil {
		s.log.Warn("Firebase Auth returned nil user", "email", user.Email)
		return nil, utils.UserNotFound
	}

	s.log.Info("User found in Firebase Auth", "email", user.Email, "uid", userAuth.UID, "email_verified", userAuth.EmailVerified)
	return userAuth, nil
}

func (s *registryRequestParams) setTenantInAuthProvider(ctx context.Context, uid, tenant *string) (*auth.UserRecord, error) {
	ctx, span := s.startSpan(ctx, "ServiceSetTenantInAuthProvider")
	defer span.End()

	if uid == nil {
		return nil, utils.ErrInvalidUser
	}

	if tenant == nil {
		return nil, utils.ErrInvalidUser
	}

	user, err := s.auth.SetTenant(ctx, uid, tenant)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *registryRequestParams) startSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if s.obs == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	return s.obs.Span(ctx, fmt.Sprintf("ServiceRegistryRequest.%s", operationName))
}

func (s *registryRequestParams) spanError(ctx context.Context, err error) {
	if s.obs == nil {
		return
	}
	span := s.obs.Trace(ctx)
	span.RecordError(err)
	span.SetAttributes(attribute.String("error", err.Error()))
}
