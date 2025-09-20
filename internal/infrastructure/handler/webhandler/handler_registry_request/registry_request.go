package handler_registry_request

import (
	"context"
	"fmt"
	"net/http"
	"registry-account/internal/core/dto/dto_registry_request"
	"registry-account/internal/core/entity/entity_registry_request"
	entitytenant "registry-account/internal/core/entity/entity_tenant"
	"registry-account/internal/core/entity/entity_user"
	"registry-account/pkg/authclient"
	"registry-account/pkg/logger"
	"registry-account/pkg/telemetry"
	"registry-account/pkg/utils"
	"registry-account/pkg/webserver"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type HandlerRegistryRequest interface{}

type handlerRegistryRequestParams struct {
	web                   webserver.ServerInterface
	auth                  authclient.AuthenticationProvider
	svcRegistryRequest    entity_registry_request.ServiceRegistryRequest
	obs                   telemetry.OtelObservability
	log                   logger.LoggerInterface
	entityRegistryRequest entity_registry_request.RegistryRequest
	userRecord            *auth.UserRecord
	tenant                entitytenant.TenantRequestData
	user                  entity_user.UserRequestData
}

func InitializeHandlerRegistryRequest(
	web webserver.ServerInterface,
	auth authclient.AuthenticationProvider,
	rr entity_registry_request.ServiceRegistryRequest,
	obs telemetry.OtelObservability,
	log logger.LoggerInterface,
	tenant entitytenant.TenantRequestData,
	user entity_user.UserRequestData,
) error {

	if rr == nil {
		return utils.ServiceRegistryRequestInvalidRepository
	}

	if web == nil {
		return utils.ServiceRegistryRequestInvalidWebServer
	}

	if auth == nil {
		return utils.ServiceRegistryRequestInvalidAuthenticationProvider
	}

	if tenant == nil {
		return utils.ServiceRegistryRequestInvalidTenantProvider
	}

	if user == nil {
		return utils.ServiceRegistryRequestInvalidUserProvider
	}

	hrr := handlerRegistryRequestParams{
		obs:                obs,
		log:                log,
		svcRegistryRequest: rr,
		web:                web,
		auth:               auth,
		tenant:             tenant,
		user:               user,
	}

	err := hrr.setupRoutes()
	if err != nil {
		return fmt.Errorf("failed to setup routes at handlerRegistryRequest: %w", err)
	}

	return nil
}

func (h *handlerRegistryRequestParams) setupRoutes() error {
	h.web.RegisterRoute("OPTIONS", "/api/v1/registry", h.handleOptionsRequest)
	h.web.RegisterRoute("POST", "/api/v1/registry", h.registryNewAccount)
	h.web.RegisterRoute("GET", "/api/v1/tenant/:name", h.getTenant)
	h.web.RegisterRoute("GET", "/api/v1/user/:email", h.getUser)

	return nil
}

func (h *handlerRegistryRequestParams) getTenant(c *gin.Context) {
	ctx, span := h.startSpan(c.Request.Context(), "handlerGetTenant")
	defer span.End()
	tenantName := c.Param("name")
	if tenantName == "" {
		h.handleError(c, ctx, http.StatusBadRequest, "tenant name is required", nil)
		return
	}

	tenant, err := h.tenant.Get(ctx, &entitytenant.TenantFilter{
		Name: &tenantName,
	})
	if err != nil {
		h.handleError(c, ctx, http.StatusInternalServerError, "error to get tenant", err)
		return
	}

	if tenant == nil {
		h.handleError(c, ctx, http.StatusNotFound, "tenant not found", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "tenant found",
		"tenant":  tenant,
	})
}
func (h *handlerRegistryRequestParams) getUser(c *gin.Context) {
	ctx, span := h.startSpan(c.Request.Context(), "handlerGetUser")
	defer span.End()

	userEmail := c.Param("email")
	if userEmail == "" {
		h.handleError(c, ctx, http.StatusBadRequest, "email of user is required", nil)
		return
	}

	user, err := h.user.Get(ctx, &entity_user.UserFilter{
		Email: &userEmail,
	})
	if err != nil {
		h.handleError(c, ctx, http.StatusInternalServerError, "error to get user", err)
		return
	}

	if user == nil {
		h.handleError(c, ctx, http.StatusNotFound, "user not found", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user found",
		"tenant":  user,
	})
}

func (h *handlerRegistryRequestParams) registryNewAccount(c *gin.Context) {
	ctx, span := h.startSpan(c.Request.Context(), "handlerRegistryNewAccount")
	defer span.End()

	var registryRequest dto_registry_request.RegistryRequestParams
	err := c.ShouldBindBodyWithJSON(&registryRequest)
	if err != nil {
		h.log.Error("Failed to bind JSON body: %v", err)
		h.handleError(c, ctx, http.StatusBadRequest, "invalid body request", err)
		return
	}

	if err = registryRequest.Validate(); err != nil {
		h.log.Error("Validation error: %v", err)
		h.handleError(c, ctx, http.StatusBadRequest, "data is not valid", err)
		return
	}

	userRequest := &entity_registry_request.RegistryRequest{
		Name:   registryRequest.UserData.Name,
		Email:  registryRequest.UserData.Email,
		Tenant: registryRequest.UserData.TenantName,
	}

	// var registryRequest entity_registry_request.RegistryRequest
	userResponse, err := h.svcRegistryRequest.RegistryNewAccount(ctx, userRequest)
	if err != nil {
		h.log.Error("Failed to register new account: %v", err)
		h.handleError(c, ctx, http.StatusInternalServerError, "error to registry a new account", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": userResponse.Success,
		"message": userResponse.Message,
		"user":    userResponse.User,
	})
}

func (h *handlerRegistryRequestParams) handleError(c *gin.Context, ctx context.Context, statusCode int, message string, err error) {
	// Loga o erro no span de telemetria
	h.spanError(ctx, err)

	// Cria a resposta JSON padronizada
	response := gin.H{
		"success": false,
		"error":   message,
		"details": err.Error(), // Adicionar detalhes pode ser útil para o debug
	}

	// Aborta a requisição com o status e o JSON
	c.AbortWithStatusJSON(statusCode, response)
}

// handleOptionsRequest handles CORS preflight requests
func (h *handlerRegistryRequestParams) handleOptionsRequest(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization")
	c.Header("Access-Control-Expose-Headers", "Content-Length")
	c.Status(http.StatusOK)
}

func (h *handlerRegistryRequestParams) startSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if h.obs == nil {
		return ctx, h.obs.Trace(ctx)
	}
	return h.obs.Span(ctx, operationName)
}

func (h *handlerRegistryRequestParams) spanError(ctx context.Context, err error) {
	span := h.obs.Trace(ctx)
	span.RecordError(err)
	span.SetAttributes(attribute.String("error", err.Error()))
}
