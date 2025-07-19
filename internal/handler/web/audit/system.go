package webhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	entity "lockari-api-application/internal/core/entity/audit"
	"lockari-api-application/internal/handler/middleware"
	"lockari-api-application/pkg/authenticator"
	cryptserver "lockari-api-application/pkg/crypt/crypt_server"
	"lockari-api-application/pkg/tokengen"
	"lockari-api-application/pkg/utils"
)

type auditSystemEventHandler struct {
	svc        entity.AuditSystemEventService
	encryptor  cryptserver.CryptDataInterface
	authClient authenticator.Authenticator
	token      tokengen.TokenGenerator
}

type auditSystemEventHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
}

func InitializeAuditSystemEventHandler(svc entity.AuditSystemEventService, encryptor cryptserver.CryptDataInterface, authClient authenticator.Authenticator, token tokengen.TokenGenerator, routerGroup *gin.RouterGroup, middlewares ...gin.HandlerFunc) (auditSystemEventHandlerInterface, error) {

	if svc == nil {
		return nil, fmt.Errorf(utils.ServiceNotFoundError, "audit system event service")
	}

	if encryptor == nil {
		return nil, fmt.Errorf(utils.ServiceNotFoundError, "audit system event encryptor")
	}

	if authClient == nil {
		return nil, fmt.Errorf(utils.ServiceNotFoundError, "audit system event auth client")
	}

	if token == nil {
		return nil, fmt.Errorf(utils.ServiceNotFoundError, "audit system event token generator")
	}

	handler := &auditSystemEventHandler{
		svc:        svc,
		encryptor:  encryptor,
		authClient: authClient,
		token:      token,
	}

	handler.setupRoutes(routerGroup, middlewares...)

	return handler, nil
}

func (h *auditSystemEventHandler) setupRoutes(routerGroup *gin.RouterGroup, middlewares ...gin.HandlerFunc) {

	auditRoutes := routerGroup.Group("/audit")
	auditRoutes.Use(middleware.ValidateTokenJWT(h.token))

	auditRoutes.POST("/auth", h.Create)
	auditRoutes.GET("/auth", h.List)
	auditRoutes.GET("/auth/:id", h.Get)

}

func (h *auditSystemEventHandler) Create(c *gin.Context) {

	token := c.GetHeader("X-TOKEN")

	_, err := h.token.Validate(token)
	if err != nil {
		log.Println("Error validating token:", err)
		c.JSON(401, gin.H{"error": "Invalid or expired token"})
		return
	}

	var body cryptserver.CryptData
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	decryptedData, err := h.encryptor.PayloadData(body.Payload)
	if err != nil {
		log.Println("Error decrypting payload:", err)
		c.JSON(400, gin.H{"error": "Error processing request data"})
		return
	}

	var auditEvent entity.AuditSystemEvent
	if err := json.Unmarshal(decryptedData, &auditEvent); err != nil {
		log.Println("Error unmarshalling audit event:", err)
		c.JSON(400, gin.H{"error": "Invalid audit event data"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "token", token)
	_, err = h.svc.Create(ctx, &auditEvent)
	if err != nil {
		log.Println("Error creating audit event:", err)
		c.JSON(500, gin.H{"error": "Failed to create audit event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Audit event created successfully"})
}

func (w *auditSystemEventHandler) Get(c *gin.Context) {

}

func (w *auditSystemEventHandler) List(c *gin.Context) {

}
