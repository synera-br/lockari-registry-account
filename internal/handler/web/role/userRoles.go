package webhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	dto "lockari-api-application/internal/core/dto/role"
	entity "lockari-api-application/internal/core/entity/role"
	mid "lockari-api-application/internal/handler/middleware"
	"lockari-api-application/pkg/authenticator"
	cryptserver "lockari-api-application/pkg/crypt/crypt_server"
	"lockari-api-application/pkg/tokengen"
)

type roleHandler struct {
	svc           entity.UserRoleService
	encryptor     cryptserver.CryptDataInterface
	authenticator authenticator.Authenticator
	tokenJWT      tokengen.TokenGenerator
}

type RoleHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	// List(c *gin.Context)
	// Extras(c *gin.Context)
}

func InitializeRoleHandler(
	svc entity.UserRoleService,
	encryptData cryptserver.CryptDataInterface,
	authClient authenticator.Authenticator,
	tokenJWT tokengen.TokenGenerator,
	routerGroup *gin.RouterGroup,
	middleware ...gin.HandlerFunc,
) RoleHandlerInterface {
	handler := &roleHandler{
		svc:           svc,
		encryptor:     encryptData,
		authenticator: authClient,
		tokenJWT:      tokenJWT,
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

func (h *roleHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...gin.HandlerFunc) {

	roles := routerGroup.Group("/user/roles")
	middleware = append(middleware, mid.ValidateTokenJWT(h.tokenJWT))
	for _, mw := range middleware {
		roles.Use(mw)
	}

	roles.POST("", h.Create)
	roles.GET("", h.Get)

}

func (h *roleHandler) Create(c *gin.Context) {
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

	_, err = h.encryptor.PayloadData(body.Payload)
	if err != nil {
		log.Println("Error decrypting payload:", err)
		c.JSON(400, gin.H{"error": "Error processing request data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "not implemented yet"})

}

func (h *roleHandler) Get(c *gin.Context) {

	token := c.GetHeader("X-TOKEN")
	_, err := h.tokenJWT.Validate(token)
	if err != nil {
		log.Println("Error validating tokenJWT:", err)
		c.JSON(401, gin.H{"error": "Invalid or expired app tokenJWT"})
		return
	}

	authorizationToken := c.GetHeader("X-AUTHORIZATION")
	_, err = h.authenticator.ValidateToken(c.Request.Context(), authorizationToken)
	if err != nil {
		log.Println("Error validating authorization token:", err)
		c.JSON(401, gin.H{"error": fmt.Sprintf("Invalid or expired user token: %s", err.Error())})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", authorizationToken)
	h.authenticator.GetUserID(ctx, authorizationToken)

	claim, err := h.authenticator.GetClaimsFromToken(ctx, authorizationToken)
	if err != nil {
		log.Println("Error getting user claim from context:", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get user claim: %s", err.Error())})
		return
	}

	userRoleFilter := &dto.RoleFilter{
		Email:    &claim.Email,
		TenantID: &claim.TenantID,
	}

	roles, err := h.svc.GetByFilter(ctx, userRoleFilter)
	if err != nil {
		log.Println("Error retrieving user roles:", err)
		c.JSON(500, gin.H{"error": "Failed to retrieve user roles"})
		return
	}

	if roles == nil {
		log.Println("No roles found")
		c.JSON(404, gin.H{"error": "No roles found"})
		return
	}

	var payloadData []byte

	if roles.UserRole != nil {
		payloadData, err = json.Marshal(roles)
		if err != nil {
			log.Println("Error marshalling roles response:", err)
			c.JSON(500, gin.H{"error": "Failed to marshal roles response"})
			return
		}
	}

	if roles.UserRoles != nil {
		payloadData, err = json.Marshal(roles.UserRoles)
		if err != nil {
			log.Println("Error marshalling user roles response:", err)
			c.JSON(500, gin.H{"error": "Failed to marshal user roles response"})
			return
		}
	}

	payload, err := h.encryptor.EncryptPayload(payloadData)
	if err != nil {
		log.Println("Error encrypting roles response:", err)
		c.JSON(500, gin.H{"error": "Failed to encrypt roles response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": payload})
}
