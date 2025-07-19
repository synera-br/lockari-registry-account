package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"lockari-api-application/pkg/authenticator"
	"lockari-api-application/pkg/tokengen"
)

type ContextKey string

const (
	// Key for storing the user's UID in the context.
	UserIDContextKey = ContextKey("userID")
	// Key for storing the user's tenant ID in the context.
	TenantIDContextKey = ContextKey("tenantID")
	// Key for storing all claims in the context.
	ClaimsContextKey = ContextKey("claims")
)

func ValidateToken(ctx context.Context, auth authenticator.Authenticator) gin.HandlerFunc {

	return func(c *gin.Context) {
		authHeader := c.GetHeader("X-AUTHORIZATION")
		if authHeader == "" {
			log.Println("Authorization header is missing")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token not provided"})
			return
		}

		// The token should be in the format "Bearer <token>".
		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == authHeader {
			log.Println("Bearer token not found in Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token format"})
			return
		}

		claims, err := auth.ValidateToken(c.Request.Context(), idToken)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 3. Extract the UID from the claims (the 'sub' field is the UID).
		uid, ok := claims["sub"].(string)
		if !ok || uid == "" {
			log.Println("UID not found in token claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User identifier not found in token"})
			return
		}

		// 4. Extract our CUSTOM claim 'tenantId'.
		tenantId, ok := claims["tenantId"].(string)
		if !ok || tenantId == "" {
			log.Printf("Custom claim 'tenantId' not found for user %s", uid)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User is not associated with a tenant"})
			return
		}

		// 5. Store the extracted information in the Gin context.
		c.Set(string(UserIDContextKey), uid)
		c.Set(string(TenantIDContextKey), tenantId)
		c.Set(string(ClaimsContextKey), claims) // Optional: store all claims if needed elsewhere

		// 6. Continue to the next handler in the chain.
		c.Next()
	}
}

func ValidateTokenJWT(token tokengen.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the JWT from the Token header
		authHeader := c.GetHeader("X-TOKEN")
		if authHeader == "" {
			log.Println("X-TOKEN header is missing")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token not provided"})
			return
		}

		// Validate the token
		claims, err := token.Validate(authHeader)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Store the claims in the context
		c.Set(string(ClaimsContextKey), claims)

		// Continue to the next handler
		c.Next()
	}
}
