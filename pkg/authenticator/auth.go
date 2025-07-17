package authenticator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// Custom errors for better error handling
var (
	ErrEmptyUserID    = errors.New("userID cannot be empty")
	ErrEmptyToken     = errors.New("authToken cannot be empty")
	ErrClientNotInit  = errors.New("auth client not initialized")
	ErrNoClaimsFound  = errors.New("verified token contained no claims")
	ErrUIDNotFound    = errors.New("UID not found in verified token or its claims")
	ErrUserIDMismatch = errors.New("userID mismatch")
)

// FirebaseConfig holds the configuration for Firebase initialization
type FirebaseConfig struct {
	// Path to the service account key file
	ServiceAccountKeyPath string `json:"serviceAccountKeyPath" yaml:"serviceAccountKeyPath"`
	// Project ID (optional if using service account key)
	ProjectID string `json:"projectId" yaml:"projectId"`
	// Database URL (optional, for Realtime Database)
	DatabaseURL string `json:"database" yaml:"database"`

	APIKey            string      `json:"apiKey" yaml:"apiKey"`
	AuthDomain        string      `json:"authDomain" yaml:"authDomain"`
	StorageBucket     string      `json:"storageBucket" yaml:"storageBucket"`
	MessagingSenderID interface{} `json:"messagingSenderId,omitempty" yaml:"messagingSenderId,omitempty"`
	AppID             string      `json:"appId" yaml:"appId"`
}

type UserCustomClaims *auth.UserRecord

func ToMap(u UserCustomClaims) map[string]interface{} {
	if u == nil {
		return nil
	}

	claims := map[string]interface{}{
		"uid":       u.UID,
		"email":     u.Email,
		"name":      u.DisplayName,
		"tenant_id": u.TenantID,
	}

	if u.CustomClaims != nil {
		for k, v := range u.CustomClaims {
			claims[k] = v
		}
	}

	return claims
}

// Authenticator defines the interface for authentication operations.
type Authenticator interface {
	ValidateToken(ctx context.Context, authToken string) (map[string]interface{}, error)
	GetClaimsFromToken(ctx context.Context, authToken string) (UserCustomClaims, error)
	IsExpired(ctx context.Context, authToken string) (bool, error)
	IsValid(ctx context.Context, authToken string) (bool, error)
	DebugToken(ctx context.Context, authToken string) (map[string]interface{}, error)
	GetTenant(ctx context.Context, authToken string) (string, error)
	GetUserID(ctx context.Context, authToken string) (string, error)
	GetUserClaim(ctx context.Context, uid string) (UserCustomClaims, error)
	GetUserEmail(ctx context.Context, authToken string) (string, error)
	GetUserName(ctx context.Context, authToken string) (string, error)
	SetTenantId(ctx context.Context, uid string, tenantId string) error
	SetCustomClaims(ctx context.Context, uid string, roles map[string]interface{}) error
	SetTenantRollback(ctx context.Context, uid string, tenantId string) error
}

// firebaseAuthenticator implements the Authenticator interface using Firebase.
type firebaseAuthenticator struct {
	client *auth.Client
}

// InitializeAuth initializes the Firebase application and returns an Authenticator.
func InitializeAuth(ctx context.Context, config *FirebaseConfig) (Authenticator, error) {
	if config == nil {
		return nil, errors.New("firebase config cannot be nil")
	}

	log.Printf("Initializing Firebase Auth with ServiceAccountKeyPath: %s", config.ServiceAccountKeyPath)
	log.Printf("Project ID: %s", config.ProjectID)

	var app *firebase.App
	var err error

	// Option 1: Using service account key file
	if config.ServiceAccountKeyPath != "" {

		// Check if file exists
		if _, err := os.Stat(config.ServiceAccountKeyPath); err != nil {
			return nil, fmt.Errorf("service account key file not found or not accessible: %w", err)
		}

		opt := option.WithCredentialsFile(config.ServiceAccountKeyPath)
		conf := &firebase.Config{
			ProjectID: config.ProjectID,
		}
		if config.DatabaseURL != "" {
			conf.DatabaseURL = config.DatabaseURL
		}
		if config.StorageBucket != "" {
			conf.StorageBucket = config.StorageBucket
		}

		app, err = firebase.NewApp(ctx, conf, opt)

		if err != nil {
			log.Printf("Failed to create Firebase app with service account file, trying environment variable...")
			// Fallback: try using environment variable
			// Set the environment variable and try again
			if envErr := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", config.ServiceAccountKeyPath); envErr != nil {
				log.Printf("Failed to set GOOGLE_APPLICATION_CREDENTIALS: %v", envErr)
			} else {
				app, err = firebase.NewApp(ctx, conf)
			}
		}
	} else {
		log.Printf("Using default credentials (ADC)")
		// Option 2: Using default credentials (ADC - Application Default Credentials)
		// This works when running on Google Cloud or with GOOGLE_APPLICATION_CREDENTIALS env var
		conf := &firebase.Config{}
		if config.ProjectID != "" {
			conf.ProjectID = config.ProjectID
		}
		if config.DatabaseURL != "" {
			conf.DatabaseURL = config.DatabaseURL
		}
		app, err = firebase.NewApp(ctx, conf)
	}

	if err != nil {
		log.Printf("Error initializing Firebase app: %v", err)
		return nil, fmt.Errorf("error initializing Firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Printf("Error getting Auth client: %v", err)
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}

	return &firebaseAuthenticator{client: client}, nil
}

// GetTenant retrieves the tenant ID from the user's custom claims.
// uid is the Firebase user ID, not the auth token.
func (fa *firebaseAuthenticator) GetTenant(ctx context.Context, uid string) (string, error) {
	if uid == "" {
		return "", ErrEmptyUserID
	}
	if fa.client == nil {
		return "", ErrClientNotInit
	}

	fmt.Println("Getting tenant for user:", uid)
	claims, err := fa.client.GetUser(ctx, uid)
	if err != nil {
		return "", fmt.Errorf("error getting claims from token: %w", err)
	}

	if claims == nil {
		return "", fmt.Errorf("user not found")
	}

	if claims.UID == "" {
		return "", fmt.Errorf("user ID not found")
	}

	tenantId, ok := claims.CustomClaims["tenant_id"]
	if !ok {
		if claims.TenantID != "" {
			return claims.TenantID, nil
		}
		return "", errors.New("tenant ID not found in user claims")
	}

	if tenantId == nil {
		return "", errors.New("tenant ID is nil")
	}

	if tenantIdStr, ok := tenantId.(string); ok && tenantIdStr == "" {
		return "", errors.New("tenant ID is empty")
	}

	return tenantId.(string), nil
}

func (fa *firebaseAuthenticator) GetUserID(ctx context.Context, authToken string) (string, error) {
	if authToken == "" {
		return "", ErrEmptyToken
	}
	if fa.client == nil {
		return "", ErrClientNotInit
	}

	user, err := fa.client.GetUser(ctx, authToken)
	if err != nil {
		return "", fmt.Errorf("error getting user: %w", err)
	}

	return user.UID, nil
}

// GetClaimsFromToken extracts and returns all claims from Firebase ID token
func (fa *firebaseAuthenticator) GetClaimsFromToken(ctx context.Context, authToken string) (UserCustomClaims, error) {
	if authToken == "" {
		return nil, ErrEmptyToken
	}
	if fa.client == nil {
		return nil, ErrClientNotInit
	}

	var token string
	if strings.HasPrefix(authToken, "Bearer") {
		result := strings.Split(authToken, " ")
		if len(result) != 2 || result[0] != "Bearer" {
			return nil, fmt.Errorf("invalid auth token format, expected 'Bearer <token>', got: %s", authToken)
		}
		token = result[1]
	} else {
		token = authToken
	}

	// Verify the ID token and extract claims
	verifiedToken, err := fa.client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("error verifying ID token: %w", err)
	}

	// Check if claims exist
	if verifiedToken.Claims == nil {
		return nil, ErrNoClaimsFound
	}

	var records UserCustomClaims
	for k, v := range verifiedToken.Claims {
		if k == "user_id" || k == "sub" {
			if uid, ok := v.(string); ok && uid != "" {
				records, err = fa.client.GetUser(ctx, uid)
				if err != nil {
					return nil, fmt.Errorf("error getting user record: %w", err)
				}
				break
			}
		}
	}

	return records, nil
}

// GetUserClaim retrieves the user claims from the context using the Firebase client.
// uid is the Firebase user ID, not the auth token.
func (fa *firebaseAuthenticator) GetUserClaim(ctx context.Context, uid string) (UserCustomClaims, error) {
	if uid == "" {
		return nil, ErrEmptyToken
	}
	if fa.client == nil {
		return nil, ErrClientNotInit
	}

	user, err := fa.client.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil

}

func (fa *firebaseAuthenticator) GetUserEmail(ctx context.Context, authToken string) (string, error) {
	if authToken == "" {
		return "", ErrEmptyToken
	}
	if fa.client == nil {
		return "", ErrClientNotInit
	}

	user, err := fa.client.GetUser(ctx, authToken)
	if err != nil {
		return "", fmt.Errorf("error getting user: %w", err)
	}

	if user == nil || user.Email == "" {
		return "", errors.New("user or email not found")
	}

	return user.Email, nil
}

func (fa *firebaseAuthenticator) GetUserName(ctx context.Context, authToken string) (string, error) {
	if authToken == "" {
		return "", ErrEmptyToken
	}
	if fa.client == nil {
		return "", ErrClientNotInit
	}

	user, err := fa.client.GetUser(ctx, authToken)
	if err != nil {
		return "", fmt.Errorf("error getting user: %w", err)
	}

	if user == nil || user.DisplayName == "" {
		return "", errors.New("user or name not found")
	}

	return user.DisplayName, nil
}

func (fa *firebaseAuthenticator) SetTenantId(ctx context.Context, uid string, tenantId string) error {
	if fa.client == nil {
		return ErrClientNotInit
	}

	if uid == "" {
		return ErrEmptyToken
	}

	if tenantId == "" {
		return errors.New("tenantId cannot be empty")
	}

	// Set custom user claims with tenantId
	claims := map[string]interface{}{
		"tenant_id": tenantId,
	}

	claims["custom_claims"] = claims

	err := fa.client.SetCustomUserClaims(ctx, uid, claims)
	if err != nil {
		return fmt.Errorf("error setting custom user claims: %w", err)
	}

	return nil
}

func (fa *firebaseAuthenticator) SetTenantRollback(ctx context.Context, uid string, tenantId string) error {
	if fa.client == nil {
		return ErrClientNotInit
	}

	if uid == "" {
		return ErrEmptyToken
	}

	// Set custom user claims with tenantId
	claims := map[string]interface{}{
		"tenantId": tenantId,
	}

	err := fa.client.SetCustomUserClaims(ctx, uid, claims)
	if err != nil {
		log.Printf("Error setting custom user claims: %v", err)
		return fmt.Errorf("error setting custom user claims: %w", err)
	}

	return nil
}

func (fa *firebaseAuthenticator) SetCustomClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	if fa.client == nil {
		return ErrClientNotInit
	}

	if _, err := fa.GetUserID(ctx, uid); err != nil {
		fmt.Println("\nError getting user ID:", err)
		return fmt.Errorf("error getting user ID: %w", err)
	}

	if uid == "" {
		fmt.Println("\nValidate claims for user:", uid)
		return ErrEmptyUserID
	}

	if len(claims) < 1 {
		fmt.Println("\nValidate claims for user:", uid)
		return errors.New("roles cannot be empty")
	}

	fmt.Println("\nGetting claims for user:", claims)
	fmt.Println("\nGetting custom claims for user:", claims["custom_claims"])

	// if claims["custom_claims"] == nil {
	// 	fmt.Println("\nValidate claims for user:", uid)
	// 	return errors.New("custom claims are required")
	// }

	// custom, ok := claims["custom_claims"].(map[string]interface{})
	// if !ok {
	// 	fmt.Println("\nValidate claims for user:", uid)
	// 	fmt.Println("Custom claims must be a map:", claims["custom_claims"])
	// 	return errors.New("custom claims must be a map")
	// }

	// if custom["tenant_id"] == nil {
	// 	fmt.Println("\nValidate claims for user:", uid)
	// 	fmt.Println("Tenant claim is required:", custom["tenant_id"])
	// 	return errors.New("tenant claim is required")
	// }

	// fmt.Println("Validate claims for user:", uid)
	// for k, v := range claims {
	// 	fmt.Printf("Setting claim %s: %v\n", k, v)
	// }

	err := fa.client.SetCustomUserClaims(ctx, uid, claims)
	if err != nil {
		fmt.Printf("Error setting custom user claims: %v", err)
		return fmt.Errorf("error setting custom user claims: %w", err)
	}

	log.Printf("Successfully set custom claims for user %s. Role: %s", uid, claims)

	return nil
}

// ValidateToken validates the Firebase ID token and ensures it belongs to the specified user.
func (fa *firebaseAuthenticator) ValidateToken(ctx context.Context, authToken string) (map[string]interface{}, error) {
	// Input validation
	if authToken == "" {
		return nil, ErrEmptyToken
	}
	if fa.client == nil {
		return nil, ErrClientNotInit
	}

	var token string
	if strings.HasPrefix(authToken, "Bearer") {
		result := strings.Split(authToken, " ")
		if len(result) != 2 || result[0] != "Bearer" {
			return nil, fmt.Errorf("invalid auth token format, expected 'Bearer <token>', got: %s", authToken)
		}

		token = result[1]
	} else {
		token = authToken
	}

	// Verify the ID token
	verifiedToken, err := fa.client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("error verifying ID token: %w", err)
	}

	// Check claims
	claims := verifiedToken.Claims
	if claims == nil {
		return nil, ErrNoClaimsFound
	}

	// Get UID from token
	tokenUID := fa.extractUID(verifiedToken, claims)
	if tokenUID == "" {
		return nil, ErrUIDNotFound
	}

	if verifiedToken.Subject == "" {
		return nil, ErrUIDNotFound
	}

	return claims, nil
}

// extractUID extracts UID from token, falling back to claims if necessary.
func (fa *firebaseAuthenticator) extractUID(token *auth.Token, claims map[string]interface{}) string {
	if token.UID != "" {
		return token.UID
	}

	// Fallback to claims
	if uidFromClaims, ok := claims["user_id"].(string); ok && uidFromClaims != "" {
		return uidFromClaims
	}

	// Try standard 'sub' claim as well
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub
	}

	return ""
}

// isTokenExpiredError checks if the error indicates an expired token
func (fa *firebaseAuthenticator) isTokenExpiredError(err error) bool {
	if err == nil {
		return false
	}

	// Check common error messages that indicate token expiry
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "expired") ||
		strings.Contains(errStr, "token has expired") ||
		strings.Contains(errStr, "id token has expired")
}

// IsExpired checks if the given Firebase authToken is expired.
func (fa *firebaseAuthenticator) IsExpired(ctx context.Context, authToken string) (bool, error) {
	if authToken == "" {
		return true, ErrEmptyToken
	}
	if fa.client == nil {
		return true, ErrClientNotInit
	}

	_, err := fa.client.VerifyIDToken(ctx, authToken)
	if err != nil {
		// Check if the error indicates token expiry
		if fa.isTokenExpiredError(err) {
			return true, nil // Token is confirmed expired
		}

		// For other verification errors, it's invalid but not necessarily expired
		return false, fmt.Errorf("token verification failed, not necessarily due to expiry: %w", err)
	}

	// If VerifyIDToken is successful, the token is not expired
	return false, nil
}

// IsValid checks if the given Firebase authToken is valid (not expired, correctly signed, etc.).
func (fa *firebaseAuthenticator) IsValid(ctx context.Context, authToken string) (bool, error) {
	if authToken == "" {
		return false, ErrEmptyToken
	}
	if fa.client == nil {
		return false, ErrClientNotInit
	}

	_, err := fa.client.VerifyIDToken(ctx, authToken)
	if err != nil {
		// Any error from VerifyIDToken means the token is not valid
		return false, fmt.Errorf("token validation failed: %w", err)
	}

	// If no error, the token is valid
	return true, nil
}

func (fa *firebaseAuthenticator) DebugToken(ctx context.Context, authToken string) (map[string]interface{}, error) {
	if authToken == "" {
		return nil, ErrEmptyToken
	}
	if fa.client == nil {
		return nil, ErrClientNotInit
	}

	// Verify the ID token
	verifiedToken, err := fa.client.VerifyIDToken(ctx, authToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying ID token: %w", err)
	}

	fmt.Println("Verified Token:", *verifiedToken)
	fmt.Println("Claims:", verifiedToken.Claims)
	fmt.Println("UID:", verifiedToken.UID)
	fmt.Println("Tenant:", verifiedToken.Firebase.Tenant)
	fmt.Println("Sign-in provider:", verifiedToken.Firebase.Identities)

	// Return the claims for debugging
	return verifiedToken.Claims, nil
}
