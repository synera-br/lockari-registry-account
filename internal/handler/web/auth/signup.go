package webhandler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/auth"
	mid "github.com/synera-br/lockari-backend-app/internal/handler/middleware"
	"github.com/synera-br/lockari-backend-app/pkg/authenticator"
	cryptserver "github.com/synera-br/lockari-backend-app/pkg/crypt/crypt_server"
	"github.com/synera-br/lockari-backend-app/pkg/tokengen"
)

type signupHandler struct {
	svc        entity.SignupEventService
	encryptor  cryptserver.CryptDataInterface
	authClient authenticator.Authenticator
	tokenJWT   tokengen.TokenGenerator
}

type SignupHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Extras(c *gin.Context)
}

func InitializeSignupHandler(
	svc entity.SignupEventService,
	encryptData cryptserver.CryptDataInterface,
	authClient authenticator.Authenticator,
	tokenJWT tokengen.TokenGenerator,
	routerGroup *gin.RouterGroup,
	middleware ...gin.HandlerFunc,
) SignupHandlerInterface {
	handler := &signupHandler{
		svc:        svc,
		encryptor:  encryptData,
		authClient: authClient,
		tokenJWT:   tokenJWT,
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

func (h *signupHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...gin.HandlerFunc) {

	signupRoutes := routerGroup.Group("/auth/signup")
	middleware = append(middleware, mid.ValidateTokenJWT(h.tokenJWT))
	for _, mw := range middleware {
		signupRoutes.Use(mw)
	}

	signupRoutes.POST("", h.Create)
	signupRoutes.GET("", h.List)
	signupRoutes.GET("/:id", h.Get)

}

func (h *signupHandler) Create(c *gin.Context) {
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

	var signup entity.Signup
	if err := json.Unmarshal(decryptedData, &signup); err != nil {
		log.Println("Error unmarshalling signup event:", err)
		c.JSON(400, gin.H{"error": "Invalid signup event data"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "token", token)
	_, err = h.svc.Create(ctx, &signup)
	if err != nil {
		log.Println("Error creating signup event:", err)
		c.JSON(500, gin.H{"error": "Failed to create signup event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signup event created successfully"})
}

func (h *signupHandler) Get(c *gin.Context) {
	log.Println("Retrieving signup event...")
	c.JSON(200, gin.H{"message": "Signup event retrieved successfully"})
}

func (h *signupHandler) List(c *gin.Context) {
	log.Println("Listing signup events...")
	c.JSON(200, gin.H{"message": "Signup events listed successfully"})
}

func (h *signupHandler) Extras(c *gin.Context) {

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

func (h *signupHandler) WithJWT(c *gin.Context) {
	log.Println("Handling signup with JWT...")

	// Exemplo de geração de token normal

	c.JSON(200, gin.H{
		"message": "JWT tokens generated successfully",
	})
}
