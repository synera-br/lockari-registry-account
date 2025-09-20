package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(c *gin.Context) {

	// Retieve the cookie from the request
	tokenStr, err := c.Cookie("Auth")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	// Extract the JWT token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte("secret"), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to parse Token"})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token claims failed"})
		c.Abort()
		return
	}

	// Check expiry of the token
	if claims["ttl"].(float64) < float64(time.Now().Unix()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
		c.Abort()
		return
	}

	// Extract the user from the Token
	user, ok := claims["userID"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
		c.Abort()
		return
	}

	// Set the current user in the context
	c.Set("user", user)

	// Go to the next in chain
	c.Next()
}
