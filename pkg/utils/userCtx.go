package utils

import (
	"context"
	"fmt"
	"strings"

	"firebase.google.com/go/v4/auth"
	"lockari-api-application/pkg/authenticator"
)

type AuthorizationToken string
type UserIDContextKey struct {
	auth.UserProvider
	Token string
}

func GetUserID(ctx context.Context) (*string, error) {

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf(ContextError, err.Error())
	}

	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil {
		return nil, fmt.Errorf("userID not found in context")
	}

	userID, ok := userIDFromCtx.(string)
	if !ok {
		return nil, fmt.Errorf("userID in context is not a string")
	}

	if userID == "" {
		return nil, fmt.Errorf("userID in context is empty")
	}

	if len(userID) < 1 {
		return nil, fmt.Errorf("userID in context is too short")
	}

	return &userID, nil
}

func ValidateTokenFromContext(ctx context.Context, auth authenticator.Authenticator) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf(ContextError, err.Error())
	}
	tokenFromCtx := ctx.Value("token")
	if tokenFromCtx == nil {
		return "", fmt.Errorf("token not found in context")
	}

	_, err := auth.ValidateToken(ctx, tokenFromCtx.(string))
	if err != nil {
		return "", fmt.Errorf("invalid token: %s", err.Error())
	}

	return tokenFromCtx.(string), nil
}

func GetTokenFromContext(ctx context.Context) string {
	if err := ctx.Err(); err != nil {
		return ""
	}

	tokenFromCtx := ctx.Value("token")
	if tokenFromCtx == nil {
		return ""
	}

	return tokenFromCtx.(string)
}

func GetAuthorizationFromContext(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf(ContextError, err.Error())
	}

	authFromCtx := ctx.Value("Authorization")
	if authFromCtx == nil {
		return "", fmt.Errorf("authorization not found in context")
	}

	auth, ok := authFromCtx.(string)
	if !ok {
		return "", fmt.Errorf("authorization in context is not a string")
	}

	if auth == "" {
		return "", fmt.Errorf("authorization in context is empty")
	}

	if strings.HasPrefix(auth, "Bearer") {
		token := strings.Split(auth, " ")
		if len(token) != 2 || token[0] != "Bearer" {
			return "", fmt.Errorf("invalid authorization format, expected 'Bearer <token>', got: %s", auth)
		}

		return token[1], nil
	}

	return auth, nil
}

func GetAuthorizationClaimFromContext(ctx context.Context, auth authenticator.Authenticator) (*auth.UserRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf(ContextError, err.Error())
	}

	authFromCtx := ctx.Value("Authorization")
	if authFromCtx == nil {
		return nil, fmt.Errorf("authorization not found in context")
	}

	tokenStr, ok := authFromCtx.(string)
	if !ok {
		return nil, fmt.Errorf("authorization in context is not a string")
	}

	if tokenStr == "" {
		return nil, fmt.Errorf("authorization in context is empty")
	}

	token := tokenStr
	if strings.HasPrefix(tokenStr, "Bearer") {
		hasToken := strings.Split(tokenStr, " ")
		if len(hasToken) != 2 || hasToken[0] != "Bearer" {
			return nil, fmt.Errorf("invalid authorization format, expected 'Bearer <token>', got: %s", tokenStr)
		}
		token = hasToken[1]
	}

	claim, err := auth.GetUserClaim(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user claim from context: %s", err.Error())
	}

	if claim == nil {
		return nil, fmt.Errorf("user claim is nil")
	}

	return claim, nil
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf(ContextError, err.Error())
	}

	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil {
		return "", fmt.Errorf("userID not found in context")
	}

	userID, ok := userIDFromCtx.(string)
	if !ok {
		return "", fmt.Errorf("userID in context is not a string")
	}

	if userID == "" {
		return "", fmt.Errorf("userID in context is empty")
	}

	return userID, nil
}
