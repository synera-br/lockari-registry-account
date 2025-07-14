package webhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/auth"
	mid "github.com/synera-br/lockari-backend-app/internal/handler/middleware"
	"github.com/synera-br/lockari-backend-app/internal/handler/web"
	"github.com/synera-br/lockari-backend-app/pkg/authenticator"
	cryptserver "github.com/synera-br/lockari-backend-app/pkg/crypt/crypt_server"
	"github.com/synera-br/lockari-backend-app/pkg/tokengen"
)

type loginHandler struct {
	svc        entity.LoginEventService
	encryptor  cryptserver.CryptDataInterface
	authClient authenticator.Authenticator
	token      tokengen.TokenGenerator
}

type LoginHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
}

func InitializeLoginHandler(
	svc entity.LoginEventService,
	encryptData cryptserver.CryptDataInterface,
	authClient authenticator.Authenticator,
	token tokengen.TokenGenerator,
	routerGroup *gin.RouterGroup,
	middleware ...gin.HandlerFunc,
) LoginHandlerInterface {
	handler := &loginHandler{
		svc:        svc,
		encryptor:  encryptData,
		authClient: authClient,
		token:      token,
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

func (h *loginHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...gin.HandlerFunc) {

	loginRoutes := routerGroup.Group("/auth/login")
	middleware = append(middleware, mid.ValidateToken(&gin.Context{}, h.authClient))
	for _, mw := range middleware {
		loginRoutes.Use(mw)
	}

	loginRoutes.POST("", h.Create)
	loginRoutes.GET("", h.List)
	loginRoutes.GET("/:id", h.Get)
}

func (h *loginHandler) Create(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload cryptserver.CryptData
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptor.PayloadData(payload.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data: " + err.Error()})
		return
	}

	var loginEvent entity.LoginEvent
	if err := json.Unmarshal(decryptedData, &loginEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error unmarshalling login event: " + err.Error()})
		return
	}

	if err := loginEvent.IsValid(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login event: " + err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	result, err := h.authClient.DebugToken(ctx, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error validating token: " + err.Error()})
		return
	}
	fmt.Println("Token validation result:", result)

	c.JSON(200, gin.H{"message": "Login event created successfully"})
}

func (h *loginHandler) Get(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Login event retrieved successfully"})
}

func (h *loginHandler) List(c *gin.Context) {
	log.Println("Listing signup events...")
	c.JSON(200, gin.H{"message": "Login events listed successfully"})
}
