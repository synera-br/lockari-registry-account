package authclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type AuthenticationProvider interface {
	AuthMiddleware(tenantIDKey string) gin.HandlerFunc
	AuthenticateUser(ctx context.Context, token string) (*auth.Token, error)
	CreateUser(ctx context.Context, email, password, displayName string) (*auth.UserRecord, error)
	GetCurrentTenantUser(ctx context.Context, token string, tenantIDKey string) (*auth.Token, string, error)
	GetUserByID(ctx context.Context, uid string) (*auth.UserRecord, error)
	Initialize(ctx context.Context, cfg Config) error
	UpdateUserCustomClaims(ctx context.Context, uid string, claims map[string]interface{}) error
	ValidateToken(ctx context.Context, token string) (*auth.Token, error)
	GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error)
	SetTenant(ctx context.Context, uid, tenant *string) (*auth.UserRecord, error)
	GetClaims(ctx context.Context, uid *string) (map[string]interface{}, error)
}

const (
	FirebaseAuthHeader = "Authorization"
	TracerName         = "firebaseauth"
)

type AuthConfig struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

// Config holds the Firebase authentication configuration.
type Config struct {
	ServiceAccountKeyPath string
	ProjectID             string
	EnableTracing         bool
}

// FirebaseAuth provides methods for interacting with Firebase Authentication.
type FirebaseAuth struct {
	client        *auth.Client
	projectID     string
	enableTracing bool
}

// NewFirebaseAuth creates a new FirebaseAuth instance.
func NewFirebaseAuth(filePath *string) (AuthenticationProvider, error) {

	if filePath == nil || *filePath == "" {
		return nil, fmt.Errorf("config file to authentication cannot be empty")
	}

	cfg, err := loadConfig(filePath)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	if cfg == nil || cfg.ProjectID == "" {
		return nil, fmt.Errorf("config is nil")
	}

	return &FirebaseAuth{
		projectID: cfg.ProjectID,
	}, nil
}

func loadConfig(filePath *string) (*AuthConfig, error) {
	b, err := os.ReadFile(*filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config AuthConfig
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config file: %w", err)
	}

	if config.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	return &config, nil
}

// Initialize initializes the Firebase authentication client.
func (f *FirebaseAuth) Initialize(ctx context.Context, cfg Config) error {
	opt := option.WithCredentialsFile(cfg.ServiceAccountKeyPath)
	config := &firebase.Config{
		ProjectID: cfg.ProjectID,
	}

	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		return fmt.Errorf("error initializing firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("error initializing firebase auth client: %w", err)
	}

	f.client = client
	f.projectID = cfg.ProjectID
	f.enableTracing = cfg.EnableTracing

	return nil
}

// AuthenticateUser authenticates a user based on the provided JWT token.
func (f *FirebaseAuth) AuthenticateUser(ctx context.Context, token string) (*auth.Token, error) {
	ctx, span := f.startSpan(ctx, "AuthenticateUser")
	defer span.End()

	decodedToken, err := f.client.VerifyIDToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, fmt.Errorf("error verifying id token: %w", err)
	}

	span.SetAttributes(attribute.String("user_id", decodedToken.UID))
	return decodedToken, nil
}

// ValidateToken validates a JWT token.
func (f *FirebaseAuth) ValidateToken(ctx context.Context, token string) (*auth.Token, error) {
	ctx, span := f.startSpan(ctx, "ValidateToken")
	defer span.End()

	decodedToken, err := f.client.VerifyIDToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, fmt.Errorf("error verifying id token: %w", err)
	}

	span.SetAttributes(attribute.String("user_id", decodedToken.UID))
	return decodedToken, nil
}

func (f *FirebaseAuth) SetTenant(ctx context.Context, uid, tenant *string) (*auth.UserRecord, error) {
	ctx, span := f.startSpan(ctx, "AppendTenant")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", *uid))

	if tenant == nil || *tenant == "" {
		return nil, fmt.Errorf("tenant cannot be empty")
	}

	if uid == nil || *uid == "" {
		return nil, fmt.Errorf("uid cannot be empty")
	}

	user, err := f.GetUserByID(ctx, *uid)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	claims := user.CustomClaims
	if claims == nil {
		claims = make(map[string]interface{})
	}

	var tenantList []string
	if existingTenants, ok := claims["tenants"]; ok {
		if et, ok := existingTenants.([]interface{}); ok {
			for _, t := range et {
				if ts, ok := t.(string); ok {
					tenantList = append(tenantList, ts)
				}
			}
		} else if et, ok := existingTenants.([]string); ok {
			tenantList = et
		}
	}

	tenantList = append(tenantList, *tenant)
	claims["tenants"] = tenantList

	if *uid != user.UID {
		return nil, fmt.Errorf("error to validate user")
	}

	err = f.client.SetCustomUserClaims(ctx, user.UID, claims)
	if err != nil {
		return nil, err
	}

	user, err = f.GetUserByID(ctx, *uid)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}

func (f *FirebaseAuth) GetClaims(ctx context.Context, uid *string) (map[string]interface{}, error) {
	ctx, span := f.startSpan(ctx, "GetClaims")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", *uid))
	if uid == nil || *uid == "" {
		return nil, fmt.Errorf("uid cannot be empty")
	}
	user, err := f.GetUserByID(ctx, *uid)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user.CustomClaims, nil
}

// GetUserByID retrieves a user by their ID.
func (f *FirebaseAuth) GetUserByID(ctx context.Context, uid string) (*auth.UserRecord, error) {
	ctx, span := f.startSpan(ctx, "GetUserByID")
	defer span.End()

	user, err := f.client.GetUser(ctx, uid)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	span.SetAttributes(attribute.String("user_id", user.UID))
	return user, nil
}

// GetUserByEmail retrieves a user by their ID.
func (f *FirebaseAuth) GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error) {
	ctx, span := f.startSpan(ctx, "GetUserByEmail")
	defer span.End()

	user, err := f.client.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	span.SetAttributes(attribute.String("user_id", user.UID))
	return user, nil
}

// UpdateUserCustomClaims updates the custom claims of a user.
func (f *FirebaseAuth) UpdateUserCustomClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	ctx, span := f.startSpan(ctx, "UpdateUserCustomClaims")
	defer span.End()

	err := f.client.SetCustomUserClaims(ctx, uid, claims)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return fmt.Errorf("error setting custom user claims: %w", err)
	}

	span.SetAttributes(attribute.String("user_id", uid))
	return nil
}

// GetCurrentTenantUser retrieves the current tenant user based on the JWT token.
func (f *FirebaseAuth) GetCurrentTenantUser(ctx context.Context, token string, tenantIDKey string) (*auth.Token, string, error) {
	ctx, span := f.startSpan(ctx, "GetCurrentTenantUser")
	defer span.End()

	decodedToken, err := f.ValidateToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		return nil, "", fmt.Errorf("error validating token: %w", err)
	}

	claims := decodedToken.Claims
	tenantID, ok := claims[tenantIDKey].(string)

	if !ok || tenantID == "" {
		span.RecordError(fmt.Errorf("tenant ID not found in claims"))
		return nil, "", fmt.Errorf("tenant ID not found in claims")
	}

	span.SetAttributes(attribute.String("user_id", decodedToken.UID), attribute.String("tenant_id", tenantID))
	return decodedToken, tenantID, nil
}

// AuthMiddleware is a Gin middleware that authenticates users using Firebase.
func (f *FirebaseAuth) AuthMiddleware(tenantIDKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx, span := f.startSpan(ctx, "AuthMiddleware")
		defer span.End()

		authHeader := c.GetHeader(FirebaseAuthHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			span.RecordError(fmt.Errorf("missing authorization header"))
			return
		}

		token := strings.Replace(authHeader, "Bearer ", "", 1)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			span.RecordError(fmt.Errorf("invalid authorization header format"))
			return
		}

		decodedToken, tenantID, err := f.GetCurrentTenantUser(ctx, token, tenantIDKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			span.RecordError(err)
			return
		}

		c.Set("uid", decodedToken.UID)
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}

func (f *FirebaseAuth) startSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if !f.enableTracing {
		return ctx, trace.SpanFromContext(ctx)
	}

	tracer := otel.Tracer(TracerName)
	ctx, span := tracer.Start(ctx, operationName)
	return ctx, span
}

// ExtractTenantIDFromToken extracts the tenant ID from the JWT token claims.
func ExtractTenantIDFromToken(tokenString string, tenantIDKey string) (string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("error parsing token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims type")
	}

	tenantID, ok := claims[tenantIDKey].(string)
	if !ok {
		return "", fmt.Errorf("tenant ID not found in claims")
	}

	return tenantID, nil
}

// CreateUser creates a new user in Firebase Authentication.
func (f *FirebaseAuth) CreateUser(ctx context.Context, email, password, displayName string) (*auth.UserRecord, error) {
	ctx, span := f.startSpan(ctx, "CreateUser")
	defer span.End()

	params := (&auth.UserToCreate{}).
		Email(email).
		DisplayName(displayName)

	customClaims := make(map[string]interface{})

	// If password is not provided, it's a federated login (e.g., Google)
	if password == "google-auth" {
		params.EmailVerified(true)
		customClaims["auth_method"] = "google" // Or simply "federated"
	} else {
		// Otherwise, it's a standard email/password user
		params.Password(password)
		params.EmailVerified(false)
		customClaims["auth_method"] = "email_password"
	}

	user, err := f.client.CreateUser(ctx, params)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, fmt.Errorf("error creating user: %w", err)
	}


	span.SetAttributes(attribute.String("user_id", user.UID))
	span.SetAttributes(attribute.String("email", email))
	return user, nil
}
