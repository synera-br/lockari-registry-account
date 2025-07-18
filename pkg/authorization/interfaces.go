package authorization

import (
	"context"
	"time"
)

// AuthorizationService é a interface básica para operações OpenFGA
type AuthorizationService interface {
	// Check verifica uma permissão única
	Check(ctx context.Context, req *CheckRequest) (*CheckResponse, error)

	// CheckBatch verifica múltiplas permissões em uma chamada
	// CheckBatch(ctx context.Context, reqs []*CheckRequest) ([]*CheckResponse, error)
	CheckBatch(ctx context.Context, reqs []*CheckRequest) (*BatchCheckResponse, error)

	// ListObjects lista todos os objetos que o usuário pode acessar
	ListObjects(ctx context.Context, req *ListObjectsRequest) (*ListObjectsResponse, error)

	// Write cria relacionamentos (tuplas)
	Write(ctx context.Context, req *WriteRequest) error

	// Delete remove relacionamentos
	Delete(ctx context.Context, req *DeleteRequest) error

	// Health verifica se o OpenFGA está disponível
	Health(ctx context.Context) (*HealthCheckResponse, error)
}

// LockariAuthorizationService é a interface específica do domínio Lockari
type LockariAuthorizationService interface {
	AuthorizationService

	// === VAULT OPERATIONS ===

	// CanAccessVault verifica se o usuário pode acessar um vault
	CanAccessVault(ctx context.Context, userID, vaultID string, permission VaultPermission) (bool, error)

	// SetupVault configura um novo vault com permissões iniciais
	SetupVault(ctx context.Context, vaultID, tenantID, ownerID string) error

	// ShareVault compartilha um vault com outro usuário
	ShareVault(ctx context.Context, vaultID, ownerID, targetUserID string, permission VaultPermission) error

	// RevokeVaultAccess revoga acesso a um vault
	RevokeVaultAccess(ctx context.Context, vaultID, ownerID, targetUserID string, permission VaultPermission) error

	// ListAccessibleVaults lista todos os vaults acessíveis para o usuário
	ListAccessibleVaults(ctx context.Context, userID string) ([]string, error)

	// === SECRET OPERATIONS ===

	// CanAccessSecret verifica se o usuário pode acessar um secret
	CanAccessSecret(ctx context.Context, userID, secretID string, permission SecretPermission) (bool, error)

	// SetupSecret configura um novo secret com permissões iniciais
	SetupSecret(ctx context.Context, secretID, vaultID string) error

	// ListAccessibleSecrets lista todos os secrets acessíveis para o usuário
	ListAccessibleSecrets(ctx context.Context, userID string) ([]string, error)

	// === TENANT OPERATIONS ===

	// SetupTenant configura um novo tenant
	// SetupTenant(ctx context.Context, tenantID, ownerID string, features []PlanFeature, relations []string) error

	// CreateNewTenant cria um novo tenant com o usuário como owner
	CreateNewTenant(ctx context.Context, userID, tenantID string) error

	// AddUserToTenant adiciona um usuário ao tenant
	AddUserToTenant(ctx context.Context, operations []TenantOperation) error

	// RemoveUserFromTenant remove um usuário do tenant
	RemoveUserFromTenant(ctx context.Context, userID, tenantID string) error

	// ChangeUserTenantRole altera o papel de um usuário no tenant
	ChangeUserTenantRole(ctx context.Context, userID, tenantID string, newRole TenantRole) error

	// IsTenantMember verifica se o usuário é membro do tenant
	IsTenantMember(ctx context.Context, userID, tenantID string) (bool, error)

	// HasPermission verifica se o usuário tem permissão no tenant
	CanAssignRoleFromTenant(ctx context.Context, userID, tenantID *string, permission TenantRole) (bool, error)

	ListPermissionFromTenant(ctx context.Context, tenantID string) (interface{}, error)

	ListAllTenantPermissions(ctx context.Context, tenantID string) (*TenantPermissionsReport, error)
	// === GROUP OPERATIONS ===

	// CreateGroup cria um novo grupo
	CreateGroup(ctx context.Context, groupID, tenantID, ownerID string) error

	// AssociateGroupToTenant associa um grupo a um tenant
	AssociateGroupToTenant(ctx context.Context, groupID, tenantID, ownerID, relation string) error

	// AssociateUserToGroupAsMember associa um usuário a um grupo como membro
	AssociateUserToGroupAsMember(ctx context.Context, userID, groupID, tenantID string) error

	// AddUserToGroup adiciona um usuário ao grupo
	AddUserToGroup(ctx context.Context, userID, groupID string, role GroupRole) error

	// RemoveUserFromGroup remove um usuário do grupo
	RemoveUserFromGroup(ctx context.Context, userID, groupID string) error

	// ChangeUserGroupRole altera o papel de um usuário no grupo
	ChangeUserGroupRole(ctx context.Context, userID, groupID string, newRole GroupRole) error

	// DeleteGroup deleta um grupo
	DeleteGroup(ctx context.Context, groupID string) error

	// === TOKEN OPERATIONS ===

	// CreateAPIToken cria um token de API com permissões específicas
	CreateAPIToken(ctx context.Context, userID, vaultID string, permissions []TokenPermission) (string, error)

	// CheckTokenPermission verifica se um token tem uma permissão específica
	CheckTokenPermission(ctx context.Context, tokenID string, permission TokenPermission) (bool, error)

	// RevokeToken revoga um token de API
	RevokeToken(ctx context.Context, tokenID string) error

	// RefreshToken atualiza um token de API
	RefreshToken(ctx context.Context, tokenID string) (string, error)

	// ListTokens lista todos os tokens de um usuário
	ListTokens(ctx context.Context, userID string) ([]string, error)

	// === EXTERNAL SHARING (Enterprise) ===

	// InitiateExternalSharing inicia processo de compartilhamento externo
	InitiateExternalSharing(ctx context.Context, userID, vaultID, targetTenantID string) (string, error)

	// ApproveExternalSharing aprova compartilhamento externo
	ApproveExternalSharing(ctx context.Context, userID, requestID string) error

	// RejectExternalSharing rejeita compartilhamento externo
	RejectExternalSharing(ctx context.Context, userID, requestID string) error

	// ListExternalSharingRequests lista solicitações de compartilhamento externo
	ListExternalSharingRequests(ctx context.Context, tenantID string) ([]string, error)

	// === AUDIT OPERATIONS ===

	// GetAuditLogs recupera logs de auditoria
	GetAuditLogs(ctx context.Context, userID, resource string, limit int) ([]AuditEvent, error)

	// === CACHE OPERATIONS ===

	// InvalidateCache invalida o cache para um usuário específico
	InvalidateCache(ctx context.Context, userID string) error

	// GetCacheStats recupera estatísticas do cache
	GetCacheStats(ctx context.Context) (CacheStats, error)
}

// Cache é a interface para cache de permissões
type Cache interface {
	// Get recupera um valor do cache
	Get(key string) (bool, bool)

	// Set armazena um valor no cache
	Set(key string, value bool, ttl time.Duration)

	// Delete remove um valor do cache
	Delete(key string)

	// Clear limpa todo o cache
	Clear()

	// Stats retorna estatísticas do cache
	Stats() CacheStats

	// Close fecha o cache gracefully
	Close() error
}

// AuditService é a interface para sistema de auditoria
type AuditService interface {
	// LogPermissionCheck registra uma verificação de permissão
	LogPermissionCheck(ctx context.Context, event PermissionCheckEvent)

	// LogPermissionGrant registra concessão de permissão
	LogPermissionGrant(ctx context.Context, event PermissionGrantEvent)

	// LogPermissionRevoke registra revogação de permissão
	LogPermissionRevoke(ctx context.Context, event PermissionRevokeEvent)

	// LogSuspiciousActivity registra atividade suspeita
	LogSuspiciousActivity(ctx context.Context, event SuspiciousActivityEvent)

	// LogGenericEvent registra um evento genérico
	LogGenericEvent(ctx context.Context, event AuditEvent)

	// GetLogs recupera logs de auditoria
	GetLogs(ctx context.Context, userID, resource string, limit int) ([]AuditEvent, error)

	// Close fecha o serviço de auditoria gracefully
	Close() error
}

// CircuitBreaker é a interface para circuit breaker
type CircuitBreaker interface {
	// Execute executa uma função com circuit breaker
	Execute(fn func() (interface{}, error)) (interface{}, error)

	// IsOpen verifica se o circuit breaker está aberto
	IsOpen() bool

	// Reset reseta o circuit breaker
	Reset()

	// GetStats retorna estatísticas do circuit breaker
	GetStats() CircuitBreakerStats
}

// CircuitBreakerStats contém estatísticas do circuit breaker
type CircuitBreakerStats struct {
	State               string    `json:"state"` // open, half-open, closed
	FailureCount        int64     `json:"failure_count"`
	SuccessCount        int64     `json:"success_count"`
	LastFailureTime     time.Time `json:"last_failure_time"`
	LastSuccessTime     time.Time `json:"last_success_time"`
	NextRetryTime       time.Time `json:"next_retry_time"`
	ConsecutiveFailures int64     `json:"consecutive_failures"`
}

// RateLimiter é a interface para rate limiting
type RateLimiter interface {
	// Allow verifica se a operação é permitida
	Allow() bool

	// Wait aguarda até que a operação seja permitida
	Wait(ctx context.Context) error

	// Reserve reserva uma permissão
	Reserve() Reservation

	// Limit retorna o limite atual
	Limit() float64

	// Burst retorna o burst atual
	Burst() int
}

// Reservation representa uma reserva de rate limiter
type Reservation interface {
	// OK verifica se a reserva é válida
	OK() bool

	// Delay retorna o tempo de espera necessário
	Delay() time.Duration

	// Cancel cancela a reserva
	Cancel()
}

// HealthChecker é a interface para verificação de saúde
type HealthChecker interface {
	// Check verifica a saúde do serviço
	Check(ctx context.Context) error

	// Status retorna o status atual da saúde
	Status() HealthStatus
}

// HealthStatus representa o status de saúde
type HealthStatus struct {
	Healthy    bool          `json:"healthy"`
	LastCheck  time.Time     `json:"last_check"`
	LastError  string        `json:"last_error,omitempty"`
	Uptime     time.Duration `json:"uptime"`
	CheckCount int64         `json:"check_count"`
	ErrorCount int64         `json:"error_count"`
}

// MetricsCollector é a interface para coleta de métricas
type MetricsCollector interface {
	// RecordCheck registra uma verificação de permissão
	RecordCheck(allowed bool, duration time.Duration)

	// RecordCacheHit registra um hit no cache
	RecordCacheHit()

	// RecordCacheMiss registra um miss no cache
	RecordCacheMiss()

	// RecordError registra um erro
	RecordError(errorType string)

	// RecordLatency registra latência
	RecordLatency(operation string, duration time.Duration)
}

// ConfigLoader é a interface para carregamento de configuração
type ConfigLoader interface {
	// Load carrega a configuração
	Load() (*Config, error)

	// Watch observa mudanças na configuração
	Watch(callback func(*Config)) error

	// Close fecha o loader gracefully
	Close() error
}

// Logger é a interface para logging
type Logger interface {
	// Debug registra uma mensagem de debug
	Debug(msg string, fields ...interface{})

	// Info registra uma mensagem informativa
	Info(msg string, fields ...interface{})

	// Warn registra uma mensagem de aviso
	Warn(msg string, fields ...interface{})

	// Error registra uma mensagem de erro
	Error(msg string, fields ...interface{})

	// With adiciona campos ao logger
	With(fields ...interface{}) Logger
}

// Validator é a interface para validação
type Validator interface {
	// ValidateUser valida um ID de usuário
	ValidateUser(userID string) error

	// ValidateObject valida um objeto
	ValidateObject(objectType, objectID string) error

	// ValidateRelation valida uma relação
	ValidateRelation(relation string) error

	// ValidateTuple valida uma tupla
	ValidateTuple(tuple *Tuple) error

	// ValidatePermission valida uma permissão
	ValidatePermission(permission string) error
}

// Transformer é a interface para transformação de dados
type Transformer interface {
	// TransformUser transforma um ID de usuário
	TransformUser(userID string) string

	// TransformObject transforma um objeto
	TransformObject(objectType, objectID string) string

	// TransformRelation transforma uma relação
	TransformRelation(relation string) string

	// TransformTuple transforma uma tupla
	TransformTuple(tuple *Tuple) *Tuple
}

// Repository é a interface para persistência de dados
type Repository interface {
	// Store armazena dados
	Store(ctx context.Context, key string, value interface{}) error

	// Retrieve recupera dados
	Retrieve(ctx context.Context, key string) (interface{}, error)

	// Delete deleta dados
	Delete(ctx context.Context, key string) error

	// List lista dados
	List(ctx context.Context, pattern string) ([]interface{}, error)

	// Close fecha o repositório gracefully
	Close() error
}

// EventBus é a interface para sistema de eventos
type EventBus interface {
	// Publish publica um evento
	Publish(ctx context.Context, event interface{}) error

	// Subscribe se inscreve em eventos
	Subscribe(eventType string, handler func(event interface{})) error

	// Unsubscribe cancela inscrição
	Unsubscribe(eventType string, handler func(event interface{})) error

	// Close fecha o event bus gracefully
	Close() error
}
