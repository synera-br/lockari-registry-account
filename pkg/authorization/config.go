package authorization

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config contém todas as configurações do OpenFGA
type Config struct {
	// === OPENFGA SERVER ===
	APIURL               string `mapstructure:"api_url"`
	StoreID              string `mapstructure:"store_id"`
	AuthorizationModelID string `mapstructure:"authorization_model_id"`

	// === AUTHENTICATION ===
	APITokenIssuer string `mapstructure:"api_token_issuer"`
	APIAudience    string `mapstructure:"api_audience"`
	ClientID       string `mapstructure:"client_id"`
	ClientSecret   string `mapstructure:"client_secret"`
	Scopes         string `mapstructure:"scopes"`

	// === TIMEOUTS & RETRY ===
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
	RetryDelay    time.Duration `mapstructure:"retry_delay"`
	MaxRetryDelay time.Duration `mapstructure:"max_retry_delay"`

	// === CACHE ===
	CacheEnabled         bool          `mapstructure:"cache_enabled"`
	CacheTTL             time.Duration `mapstructure:"cache_ttl"`
	CacheMaxSize         int           `mapstructure:"cache_max_size"`
	CacheCleanupInterval time.Duration `mapstructure:"cache_cleanup_interval"`

	// === AUDIT ===
	AuditEnabled bool   `mapstructure:"audit_enabled"`
	AuditLevel   string `mapstructure:"audit_level"`

	// === PERFORMANCE ===
	MaxConcurrentRequests int `mapstructure:"max_concurrent_requests"`
	BatchSize             int `mapstructure:"batch_size"`
	ConnectionPoolSize    int `mapstructure:"connection_pool_size"`

	// === HEALTH CHECK ===
	HealthCheckEnabled  bool          `mapstructure:"health_check_enabled"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
	HealthCheckTimeout  time.Duration `mapstructure:"health_check_timeout"`

	// === RATE LIMITING ===
	RateLimitEnabled bool    `mapstructure:"rate_limit_enabled"`
	RateLimitRPS     float64 `mapstructure:"rate_limit_rps"`
	RateLimitBurst   int     `mapstructure:"rate_limit_burst"`

	// === CIRCUIT BREAKER ===
	CircuitBreakerEnabled   bool          `mapstructure:"circuit_breaker_enabled"`
	CircuitBreakerThreshold int           `mapstructure:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `mapstructure:"circuit_breaker_timeout"`
	CircuitBreakerResetTime time.Duration `mapstructure:"circuit_breaker_reset_time"`

	// === METRICS ===
	MetricsEnabled bool   `mapstructure:"metrics_enabled"`
	MetricsPrefix  string `mapstructure:"metrics_prefix"`

	// === LOGGING ===
	LogLevel  string `mapstructure:"log_level"`
	LogFormat string `mapstructure:"log_format"`

	// === DEVELOPMENT ===
	Development bool `mapstructure:"development"`
	Debug       bool `mapstructure:"debug"`
}

// NewConfig cria uma nova configuração com valores padrão
func NewConfig() *Config {
	cfg := &Config{
		// OpenFGA Server
		APIURL:               "http://localhost:8080",
		StoreID:              "",
		AuthorizationModelID: "",

		// Authentication
		APITokenIssuer: "",
		APIAudience:    "",
		ClientID:       "",
		ClientSecret:   "",
		Scopes:         "read write",

		// Timeouts & Retry
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
		MaxRetryDelay: 10 * time.Second,

		// Cache
		CacheEnabled:         true,
		CacheTTL:             5 * time.Minute,
		CacheMaxSize:         10000,
		CacheCleanupInterval: 10 * time.Minute,

		// Audit
		AuditEnabled: true,
		AuditLevel:   "info",

		// Performance
		MaxConcurrentRequests: 100,
		BatchSize:             50,
		ConnectionPoolSize:    10,

		// Health Check
		HealthCheckEnabled:  true,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,

		// Rate Limiting
		RateLimitEnabled: true,
		RateLimitRPS:     100.0,
		RateLimitBurst:   200,

		// Circuit Breaker
		CircuitBreakerEnabled:   true,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
		CircuitBreakerResetTime: 60 * time.Second,

		// Metrics
		MetricsEnabled: true,
		MetricsPrefix:  "lockari_authorization",

		// Logging
		LogLevel:  "info",
		LogFormat: "json",

		// Development
		Development: false,
		Debug:       false,
	}

	fmt.Println("\nUsing default OpenFGA configuration:", *cfg)
	return cfg
}

// Validate valida se a configuração está correta
func (c *Config) Validate() error {
	if c.APIURL == "" {
		return NewConfigError("api_url", "API URL is required", nil)
	}

	if c.StoreID == "" {
		return NewConfigError("store_id", "Store ID is required", nil)
	}

	if c.AuthorizationModelID == "" {
		return NewConfigError("authorization_model_id", "Authorization Model ID is required", nil)
	}

	if c.Timeout <= 0 {
		return NewConfigError("timeout", "timeout must be positive", nil)
	}

	if c.RetryAttempts < 0 {
		return NewConfigError("retry_attempts", "retry attempts cannot be negative", nil)
	}

	if c.RetryDelay < 0 {
		return NewConfigError("retry_delay", "retry delay cannot be negative", nil)
	}

	if c.MaxRetryDelay < c.RetryDelay {
		return NewConfigError("max_retry_delay", "max retry delay must be greater than retry delay", nil)
	}

	if c.CacheEnabled && c.CacheMaxSize <= 0 {
		return NewConfigError("cache_max_size", "cache max size must be positive when cache is enabled", nil)
	}

	if c.CacheEnabled && c.CacheTTL <= 0 {
		return NewConfigError("cache_ttl", "cache TTL must be positive when cache is enabled", nil)
	}

	if c.MaxConcurrentRequests <= 0 {
		return NewConfigError("max_concurrent_requests", "max concurrent requests must be positive", nil)
	}

	if c.BatchSize <= 0 {
		return NewConfigError("batch_size", "batch size must be positive", nil)
	}

	if c.ConnectionPoolSize <= 0 {
		return NewConfigError("connection_pool_size", "connection pool size must be positive", nil)
	}

	if c.HealthCheckEnabled && c.HealthCheckInterval <= 0 {
		return NewConfigError("health_check_interval", "health check interval must be positive when health check is enabled", nil)
	}

	if c.HealthCheckEnabled && c.HealthCheckTimeout <= 0 {
		return NewConfigError("health_check_timeout", "health check timeout must be positive when health check is enabled", nil)
	}

	if c.RateLimitEnabled && c.RateLimitRPS <= 0 {
		return NewConfigError("rate_limit_rps", "rate limit RPS must be positive when rate limiting is enabled", nil)
	}

	if c.RateLimitEnabled && c.RateLimitBurst <= 0 {
		return NewConfigError("rate_limit_burst", "rate limit burst must be positive when rate limiting is enabled", nil)
	}

	if c.CircuitBreakerEnabled && c.CircuitBreakerThreshold <= 0 {
		return NewConfigError("circuit_breaker_threshold", "circuit breaker threshold must be positive when circuit breaker is enabled", nil)
	}

	if c.CircuitBreakerEnabled && c.CircuitBreakerTimeout <= 0 {
		return NewConfigError("circuit_breaker_timeout", "circuit breaker timeout must be positive when circuit breaker is enabled", nil)
	}

	if c.CircuitBreakerEnabled && c.CircuitBreakerResetTime <= 0 {
		return NewConfigError("circuit_breaker_reset_time", "circuit breaker reset time must be positive when circuit breaker is enabled", nil)
	}

	// Validar log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.LogLevel) {
		return NewConfigError("log_level", fmt.Sprintf("log level must be one of: %v", validLogLevels), nil)
	}

	// Validar log format
	validLogFormats := []string{"json", "text"}
	if !contains(validLogFormats, c.LogFormat) {
		return NewConfigError("log_format", fmt.Sprintf("log format must be one of: %v", validLogFormats), nil)
	}

	// Validar audit level
	validAuditLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validAuditLevels, c.AuditLevel) {
		return NewConfigError("audit_level", fmt.Sprintf("audit level must be one of: %v", validAuditLevels), nil)
	}

	return nil
}

// LoadFromViper carrega configuração do Viper
func LoadFromViper(v *viper.Viper) (*Config, error) {
	config := NewConfig()

	// Definir valores padrão no Viper
	v.SetDefault("openfga.api_url", config.APIURL)
	v.SetDefault("openfga.timeout", config.Timeout)
	v.SetDefault("openfga.retry_attempts", config.RetryAttempts)
	v.SetDefault("openfga.retry_delay", config.RetryDelay)
	v.SetDefault("openfga.max_retry_delay", config.MaxRetryDelay)
	v.SetDefault("openfga.cache_enabled", config.CacheEnabled)
	v.SetDefault("openfga.cache_ttl", config.CacheTTL)
	v.SetDefault("openfga.cache_max_size", config.CacheMaxSize)
	v.SetDefault("openfga.cache_cleanup_interval", config.CacheCleanupInterval)
	v.SetDefault("openfga.audit_enabled", config.AuditEnabled)
	v.SetDefault("openfga.audit_level", config.AuditLevel)
	v.SetDefault("openfga.max_concurrent_requests", config.MaxConcurrentRequests)
	v.SetDefault("openfga.batch_size", config.BatchSize)
	v.SetDefault("openfga.connection_pool_size", config.ConnectionPoolSize)
	v.SetDefault("openfga.health_check_enabled", config.HealthCheckEnabled)
	v.SetDefault("openfga.health_check_interval", config.HealthCheckInterval)
	v.SetDefault("openfga.health_check_timeout", config.HealthCheckTimeout)
	v.SetDefault("openfga.rate_limit_enabled", config.RateLimitEnabled)
	v.SetDefault("openfga.rate_limit_rps", config.RateLimitRPS)
	v.SetDefault("openfga.rate_limit_burst", config.RateLimitBurst)
	v.SetDefault("openfga.circuit_breaker_enabled", config.CircuitBreakerEnabled)
	v.SetDefault("openfga.circuit_breaker_threshold", config.CircuitBreakerThreshold)
	v.SetDefault("openfga.circuit_breaker_timeout", config.CircuitBreakerTimeout)
	v.SetDefault("openfga.circuit_breaker_reset_time", config.CircuitBreakerResetTime)
	v.SetDefault("openfga.metrics_enabled", config.MetricsEnabled)
	v.SetDefault("openfga.metrics_prefix", config.MetricsPrefix)
	v.SetDefault("openfga.log_level", config.LogLevel)
	v.SetDefault("openfga.log_format", config.LogFormat)
	v.SetDefault("openfga.development", config.Development)
	v.SetDefault("openfga.debug", config.Debug)

	// Unmarshal a configuração
	if err := v.UnmarshalKey("openfga", config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validar configuração
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// LoadFromFile carrega configuração de um arquivo
func LoadFromFile(filePath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(filePath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return LoadFromViper(v)
}

// LoadFromEnv carrega configuração de variáveis de ambiente
func LoadFromEnv() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("OPENFGA")
	v.AutomaticEnv()

	// Mapear variáveis de ambiente
	v.BindEnv("api_url", "OPENFGA_API_URL")
	v.BindEnv("store_id", "OPENFGA_STORE_ID")
	v.BindEnv("authorization_model_id", "OPENFGA_AUTHORIZATION_MODEL_ID")
	v.BindEnv("api_token_issuer", "OPENFGA_API_TOKEN_ISSUER")
	v.BindEnv("api_audience", "OPENFGA_API_AUDIENCE")
	v.BindEnv("client_id", "OPENFGA_CLIENT_ID")
	v.BindEnv("client_secret", "OPENFGA_CLIENT_SECRET")
	v.BindEnv("scopes", "OPENFGA_SCOPES")
	v.BindEnv("timeout", "OPENFGA_TIMEOUT")
	v.BindEnv("retry_attempts", "OPENFGA_RETRY_ATTEMPTS")
	v.BindEnv("retry_delay", "OPENFGA_RETRY_DELAY")
	v.BindEnv("max_retry_delay", "OPENFGA_MAX_RETRY_DELAY")
	v.BindEnv("cache_enabled", "OPENFGA_CACHE_ENABLED")
	v.BindEnv("cache_ttl", "OPENFGA_CACHE_TTL")
	v.BindEnv("cache_max_size", "OPENFGA_CACHE_MAX_SIZE")
	v.BindEnv("cache_cleanup_interval", "OPENFGA_CACHE_CLEANUP_INTERVAL")
	v.BindEnv("audit_enabled", "OPENFGA_AUDIT_ENABLED")
	v.BindEnv("audit_level", "OPENFGA_AUDIT_LEVEL")
	v.BindEnv("max_concurrent_requests", "OPENFGA_MAX_CONCURRENT_REQUESTS")
	v.BindEnv("batch_size", "OPENFGA_BATCH_SIZE")
	v.BindEnv("connection_pool_size", "OPENFGA_CONNECTION_POOL_SIZE")
	v.BindEnv("health_check_enabled", "OPENFGA_HEALTH_CHECK_ENABLED")
	v.BindEnv("health_check_interval", "OPENFGA_HEALTH_CHECK_INTERVAL")
	v.BindEnv("health_check_timeout", "OPENFGA_HEALTH_CHECK_TIMEOUT")
	v.BindEnv("rate_limit_enabled", "OPENFGA_RATE_LIMIT_ENABLED")
	v.BindEnv("rate_limit_rps", "OPENFGA_RATE_LIMIT_RPS")
	v.BindEnv("rate_limit_burst", "OPENFGA_RATE_LIMIT_BURST")
	v.BindEnv("circuit_breaker_enabled", "OPENFGA_CIRCUIT_BREAKER_ENABLED")
	v.BindEnv("circuit_breaker_threshold", "OPENFGA_CIRCUIT_BREAKER_THRESHOLD")
	v.BindEnv("circuit_breaker_timeout", "OPENFGA_CIRCUIT_BREAKER_TIMEOUT")
	v.BindEnv("circuit_breaker_reset_time", "OPENFGA_CIRCUIT_BREAKER_RESET_TIME")
	v.BindEnv("metrics_enabled", "OPENFGA_METRICS_ENABLED")
	v.BindEnv("metrics_prefix", "OPENFGA_METRICS_PREFIX")
	v.BindEnv("log_level", "OPENFGA_LOG_LEVEL")
	v.BindEnv("log_format", "OPENFGA_LOG_FORMAT")
	v.BindEnv("development", "OPENFGA_DEVELOPMENT")
	v.BindEnv("debug", "OPENFGA_DEBUG")

	return LoadFromViper(v)
}

// String retorna uma representação string da configuração
func (c *Config) String() string {
	return fmt.Sprintf("Config{APIURL: %s, StoreID: %s, AuthorizationModelID: %s, CacheEnabled: %t, AuditEnabled: %t}",
		c.APIURL, c.StoreID, c.AuthorizationModelID, c.CacheEnabled, c.AuditEnabled)
}

// IsDevelopment retorna true se está em modo desenvolvimento
func (c *Config) IsDevelopment() bool {
	return c.Development
}

// IsDebug retorna true se está em modo debug
func (c *Config) IsDebug() bool {
	return c.Debug
}

// GetLogLevel retorna o nível de log configurado
func (c *Config) GetLogLevel() string {
	return c.LogLevel
}

// GetMetricsPrefix retorna o prefixo das métricas
func (c *Config) GetMetricsPrefix() string {
	return c.MetricsPrefix
}

// HasAuthentication retorna true se autenticação está configurada
func (c *Config) HasAuthentication() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// Clone cria uma cópia da configuração
func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}

// Merge combina esta configuração com outra
func (c *Config) Merge(other *Config) *Config {
	merged := c.Clone()

	if other.APIURL != "" {
		merged.APIURL = other.APIURL
	}
	if other.StoreID != "" {
		merged.StoreID = other.StoreID
	}
	if other.AuthorizationModelID != "" {
		merged.AuthorizationModelID = other.AuthorizationModelID
	}
	if other.ClientID != "" {
		merged.ClientID = other.ClientID
	}
	if other.ClientSecret != "" {
		merged.ClientSecret = other.ClientSecret
	}
	if other.Timeout > 0 {
		merged.Timeout = other.Timeout
	}
	if other.RetryAttempts > 0 {
		merged.RetryAttempts = other.RetryAttempts
	}
	if other.CacheMaxSize > 0 {
		merged.CacheMaxSize = other.CacheMaxSize
	}
	if other.CacheTTL > 0 {
		merged.CacheTTL = other.CacheTTL
	}

	return merged
}

// contains verifica se um slice contém um valor
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidateRequired valida se os campos obrigatórios estão preenchidos
func ValidateRequired(config *Config) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	if config.APIURL == "" {
		return errors.New("api_url is required")
	}

	if config.StoreID == "" {
		return errors.New("store_id is required")
	}

	if config.AuthorizationModelID == "" {
		return errors.New("authorization_model_id is required")
	}

	return nil
}

// SetDefaults define valores padrão para campos não preenchidos
func SetDefaults(config *Config) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}

	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	if config.MaxRetryDelay == 0 {
		config.MaxRetryDelay = 10 * time.Second
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}

	if config.CacheMaxSize == 0 {
		config.CacheMaxSize = 10000
	}

	if config.MaxConcurrentRequests == 0 {
		config.MaxConcurrentRequests = 100
	}

	if config.BatchSize == 0 {
		config.BatchSize = 50
	}

	if config.ConnectionPoolSize == 0 {
		config.ConnectionPoolSize = 10
	}

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	if config.LogFormat == "" {
		config.LogFormat = "json"
	}

	if config.AuditLevel == "" {
		config.AuditLevel = "info"
	}

	if config.MetricsPrefix == "" {
		config.MetricsPrefix = "lockari_authorization"
	}

	if config.Scopes == "" {
		config.Scopes = "read write"
	}
}
