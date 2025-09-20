package authzclient

import (
	"context"
	"fmt"
	"registry-account/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracerName = "authzclient"

type ConfigAuthzClient struct {
	EncryptKey     string `json:"encrypt_key" mapstructure:"encrypt_key"`
	EnabledTracing bool   `json:"enabled_tracing" mapstructure:"enabled_tracing"`
	TTL            int    `json:"ttl" mapstructure:"ttl"`
}

type tokenAuthzClient struct {
	encryptKey    string
	token         string
	exp           int
	method        *jwt.SigningMethodHMAC
	enableTracing bool
}

type AuthzClient interface {
	GenerateJwtToken(ctx context.Context) (*string, error)
	RevokeToken(ctx context.Context)
	GetToken(ctx context.Context) (*jwt.Token, error)
	IsExpiredToken(ctx context.Context) (bool, error)
	IsValidToken(ctx context.Context, tokenStr string) bool
	GetTokenString(ctx context.Context) string
	GetClaims(ctx context.Context) (jwt.MapClaims, error)
	AppendClaims(ctx context.Context, key string, value interface{}) error
}

func NewAuthzClient(cfg *ConfigAuthzClient) (AuthzClient, error) {

	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	tCfg := &tokenAuthzClient{
		exp:           cfg.TTL,
		method:        jwt.SigningMethodHS512,
		encryptKey:    cfg.EncryptKey,
		enableTracing: cfg.EnabledTracing,
	}

	if err := tCfg.validate(); err != nil {
		return nil, err
	}

	return tCfg, nil
}

func (c *tokenAuthzClient) GenerateJwtToken(ctx context.Context) (*string, error) {

	_, span := c.startSpan(ctx, "GenerateJwtToken")
	defer span.End()
	span.SetAttributes(attribute.String("method", "GenerateJwtToken"))

	if err := c.validate(); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	token := jwt.NewWithClaims(c.method, jwt.MapClaims{
		"userID": utils.NewID(),
		"exp":    time.Now().Add(time.Hour * time.Duration(c.exp)).Unix(),
		"foo":    "bar",
		"nbf":    time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(c.encryptKey))
	if err != nil {
		return nil, err
	}

	c.token = tokenString
	return &c.token, nil
}

func (c *tokenAuthzClient) RevokeToken(ctx context.Context) {
	_, span := c.startSpan(ctx, "RevokeToken")
	defer span.End()
	span.SetAttributes(attribute.String("method", "RevokeToken"))

	if c == nil {
		span.RecordError(fmt.Errorf("authz client is nil"))
		return
	}

	c.token = ""
}

func (c *tokenAuthzClient) IsExpiredToken(ctx context.Context) (bool, error) {
	ctx, span := c.startSpan(ctx, "IsExpiredToken")
	defer span.End()
	span.SetAttributes(attribute.String("method", "IsExpiredToken"))

	_, err := c.parseToken(ctx)
	if err != nil {
		return true, err
	}

	if float64(c.exp) < float64(time.Now().Unix()) {
		return true, fmt.Errorf("token is expired")
	}

	return false, nil
}

func (c *tokenAuthzClient) IsValidToken(ctx context.Context, tokenStr string) bool {
	ctx, span := c.startSpan(ctx, "IsValidToken")
	defer span.End()
	span.SetAttributes(attribute.String("method", "IsValidToken"))

	isToken, err := c.parseWithToken(ctx, tokenStr)
	if err != nil {
		return false
	}

	if float64(c.exp) < float64(time.Now().Unix()) {
		return false
	}

	if !isToken.Valid {
		return false
	}

	return true
}

func (c *tokenAuthzClient) GetToken(ctx context.Context) (*jwt.Token, error) {
	ctx, span := c.startSpan(ctx, "GetToken")
	defer span.End()
	span.SetAttributes(attribute.String("method", "GetToken"))

	jwtToken, err := c.parseToken(ctx)
	if err != nil {
		return nil, err
	}

	return jwtToken, nil
}

func (c *tokenAuthzClient) GetClaims(ctx context.Context) (jwt.MapClaims, error) {
	ctx, span := c.startSpan(ctx, "GetClaims")
	defer span.End()
	span.SetAttributes(attribute.String("method", "GetClaims"))

	jwtToken, err := c.parseToken(ctx)
	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (c *tokenAuthzClient) AppendClaims(ctx context.Context, key string, value interface{}) error {
	ctx, span := c.startSpan(ctx, "AppendClaims")
	defer span.End()
	span.SetAttributes(attribute.String("method", "AppendClaims"))

	if c == nil {
		return fmt.Errorf("authz client is nil")
	}

	if c.token == "" {
		return fmt.Errorf("token is empty")
	}

	claims, err := c.GetClaims(ctx)
	if err != nil {
		return err
	}

	claims[key] = value

	token := jwt.NewWithClaims(c.method, claims)

	tokenString, err := token.SignedString([]byte(c.encryptKey))
	if err != nil {
		return err
	}

	c.token = tokenString
	return nil
}

func (c *tokenAuthzClient) RemoveClaims(ctx context.Context) {
	_, span := c.startSpan(ctx, "RemoveClaims")
	defer span.End()
	span.SetAttributes(attribute.String("method", "RemoveClaims"))
}

func (c *tokenAuthzClient) GetTokenString(ctx context.Context) string {
	_, span := c.startSpan(ctx, "GetTokenString")
	defer span.End()
	span.SetAttributes(attribute.String("method", "GetTokenString"))

	if c == nil {
		span.RecordError(fmt.Errorf("authz client is nil"))
		return ""
	}

	if c.token == "" {
		span.RecordError(fmt.Errorf("token is empty"))
		return ""
	}
	return c.token
}

func (c *tokenAuthzClient) parseToken(ctx context.Context) (*jwt.Token, error) {
	_, span := c.startSpan(ctx, "parseToken")
	defer span.End()
	span.SetAttributes(attribute.String("method", "parseToken"))

	if c == nil {
		return nil, fmt.Errorf("authz client is nil")
	}

	if c.token == "" {
		return nil, fmt.Errorf("token is empty")
	}

	jwtToken, err := jwt.Parse(c.token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(c.encryptKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !jwtToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return jwtToken, nil
}

func (c *tokenAuthzClient) parseWithToken(ctx context.Context, token string) (*jwt.Token, error) {
	_, span := c.startSpan(ctx, "parseWithToken")
	defer span.End()
	span.SetAttributes(attribute.String("method", "parseWithToken"))

	if c == nil {
		return nil, fmt.Errorf("authz client is nil")
	}

	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}

	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(c.encryptKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !jwtToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return jwtToken, nil
}

func (c *tokenAuthzClient) validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.encryptKey == "" {
		return fmt.Errorf("encrypt key is empty")
	}

	if c.exp <= 0 {
		c.exp = 3600 * 24
	}

	return nil
}

func (c *tokenAuthzClient) startSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if !c.enableTracing {
		return ctx, trace.SpanFromContext(ctx)
	}

	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, operationName)
	return ctx, span
}
