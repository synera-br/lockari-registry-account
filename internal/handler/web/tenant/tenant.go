package webhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	mid "github.com/synera-br/lockari-backend-app/internal/handler/middleware"
	"github.com/synera-br/lockari-backend-app/pkg/authenticator"
	cryptserver "github.com/synera-br/lockari-backend-app/pkg/crypt/crypt_server"
	"github.com/synera-br/lockari-backend-app/pkg/tokengen"
)

type tenantHandler struct {
	svc        entity.TenantService
	encryptor  cryptserver.CryptDataInterface
	authClient authenticator.Authenticator
	tokenJWT   tokengen.TokenGenerator
}

type TenantHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	// List(c *gin.Context)
	// Extras(c *gin.Context)
}

func InitializeTenantHandler(
	svc entity.TenantService,
	encryptData cryptserver.CryptDataInterface,
	authClient authenticator.Authenticator,
	tokenJWT tokengen.TokenGenerator,
	routerGroup *gin.RouterGroup,
	middleware ...gin.HandlerFunc,
) TenantHandlerInterface {
	handler := &tenantHandler{
		svc:        svc,
		encryptor:  encryptData,
		authClient: authClient,
		tokenJWT:   tokenJWT,
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

func (h *tenantHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...gin.HandlerFunc) {

	signup := routerGroup.Group("/auth/signup")
	middleware = append(middleware, mid.ValidateTokenJWT(h.tokenJWT))
	for _, mw := range middleware {
		signup.Use(mw)
	}

	signup.POST("", h.Create)

	tenant := routerGroup.Group("/tenant")
	middleware = append(middleware, mid.ValidateTokenJWT(h.tokenJWT))
	for _, mw := range middleware {
		tenant.Use(mw)
	}

	tenant.GET("/", h.Get)

}

func (h *tenantHandler) Create(c *gin.Context) {
	token := c.GetHeader("X-TOKEN")

	_, err := h.tokenJWT.Validate(token)
	if err != nil {
		log.Println("Error validating tokenJWT:", err)
		c.JSON(401, gin.H{"error": "Invalid or expired tokenJWT"})
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

	var tenant entity.Tenant
	if err := json.Unmarshal(decryptedData, &tenant); err != nil {
		log.Println("Error unmarshalling tenant event:", err)
		c.JSON(400, gin.H{"error": "Invalid tenant event data"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "token", token)
	_, err = h.svc.Create(ctx, &tenant)
	if err != nil {
		log.Println("Error creating tenant event:", err)
		c.JSON(500, gin.H{"error": "Failed to create tenant event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant event created successfully"})
}

func (h *tenantHandler) Get(c *gin.Context) {

	token := c.GetHeader("X-TOKEN")
	_, err := h.tokenJWT.Validate(token)
	if err != nil {
		log.Println("Error validating tokenJWT:", err)
		c.JSON(401, gin.H{"error": "Invalid or expired app tokenJWT"})
		return
	}

	authorizationToken := c.GetHeader("X-AUTHORIZATION")
	_, err = h.authClient.ValidateToken(c.Request.Context(), authorizationToken)
	if err != nil {
		log.Println("Error validating authorization token:", err)
		c.JSON(401, gin.H{"error": fmt.Sprintf("Invalid or expired user token: %s", err.Error())})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", authorizationToken)
	tenant, err := h.svc.Get(ctx, entity.TenantFilter{})
	if err != nil {
		log.Println("Error retrieving tenant event:", err)
		c.JSON(500, gin.H{"error": "Failed to retrieve tenant event"})
		return
	}

	if tenant == nil {
		log.Println("Tenant not found")
		c.JSON(404, gin.H{"error": "Tenant not found"})
		return
	}

	responseTenant, err := NewTenantDetailsResponse(tenant)
	if err != nil {
		log.Println("Error creating tenant details response:", err)
		c.JSON(500, gin.H{"error": "Failed to create tenant details response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenant": responseTenant})
}

func (h *tenantHandler) List(c *gin.Context) {
	log.Println("Listing tenant events...")
	c.JSON(http.StatusOK, gin.H{"message": "Tenant events listed successfully"})
}

func (h *tenantHandler) Extras(c *gin.Context) {

	var body cryptserver.CryptData
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	_, err := h.encryptor.PayloadData(body.Payload)
	if err != nil {
		log.Println("Error decrypting payload:", err)
		c.JSON(400, gin.H{"error": "Error processing request data"})
		return
	}

	c.JSON(200, gin.H{"message": "Extras handled successfully"})
}

func (h *tenantHandler) WithJWT(c *gin.Context) {
	log.Println("Handling tenant with JWT...")

	// Exemplo de geração de token normal

	c.JSON(200, gin.H{
		"message": "JWT tokens generated successfully",
	})
}
