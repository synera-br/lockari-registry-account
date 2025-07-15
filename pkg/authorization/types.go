package authorization

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ===== VAULT PERMISSIONS =====

// VaultPermission representa permissões granulares de vault
type VaultPermission string

const (
	VaultPermissionView     VaultPermission = "can_view"
	VaultPermissionRead     VaultPermission = "can_read"
	VaultPermissionCopy     VaultPermission = "can_copy"
	VaultPermissionDownload VaultPermission = "can_download"
	VaultPermissionWrite    VaultPermission = "can_write"
	VaultPermissionDelete   VaultPermission = "can_delete"
	VaultPermissionShare    VaultPermission = "can_share"
	VaultPermissionManage   VaultPermission = "can_manage"
)

// AllVaultPermissions retorna todas as permissões válidas
func AllVaultPermissions() []VaultPermission {
	return []VaultPermission{
		VaultPermissionView, VaultPermissionRead, VaultPermissionCopy,
		VaultPermissionDownload, VaultPermissionWrite, VaultPermissionDelete,
		VaultPermissionShare, VaultPermissionManage,
	}
}

// IsValid verifica se a permissão é válida
func (vp VaultPermission) IsValid() bool {
	for _, valid := range AllVaultPermissions() {
		if vp == valid {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer
func (vp VaultPermission) String() string {
	return string(vp)
}

// ===== SECRET PERMISSIONS =====

// SecretPermission representa permissões granulares de secret
type SecretPermission string

const (
	SecretPermissionView           SecretPermission = "can_view"
	SecretPermissionRead           SecretPermission = "can_read"
	SecretPermissionCopy           SecretPermission = "can_copy"
	SecretPermissionDownload       SecretPermission = "can_download"
	SecretPermissionWrite          SecretPermission = "can_write"
	SecretPermissionDelete         SecretPermission = "can_delete"
	SecretPermissionReadSensitive  SecretPermission = "can_read_sensitive"
	SecretPermissionCopySensitive  SecretPermission = "can_copy_sensitive"
	SecretPermissionCopyProduction SecretPermission = "can_copy_production"
)

// AllSecretPermissions retorna todas as permissões válidas
func AllSecretPermissions() []SecretPermission {
	return []SecretPermission{
		SecretPermissionView, SecretPermissionRead, SecretPermissionCopy,
		SecretPermissionDownload, SecretPermissionWrite, SecretPermissionDelete,
		SecretPermissionReadSensitive, SecretPermissionCopySensitive,
		SecretPermissionCopyProduction,
	}
}

// IsValid verifica se a permissão é válida
func (sp SecretPermission) IsValid() bool {
	for _, valid := range AllSecretPermissions() {
		if sp == valid {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer
func (sp SecretPermission) String() string {
	return string(sp)
}

// ===== TENANT ROLES =====

// TenantRole representa papéis no tenant
type TenantRole string

const (
	TenantRoleOwner   TenantRole = "owner"
	TenantRoleAdmin   TenantRole = "admin"
	TenantRoleMember  TenantRole = "member"
	TenantRoleGuest   TenantRole = "guest"
	TenantRoleManager TenantRole = "manager"
)

// AllTenantRoles retorna todos os papéis válidos
func AllTenantRoles() []TenantRole {
	return []TenantRole{
		TenantRoleOwner, TenantRoleAdmin, TenantRoleMember, TenantRoleGuest,
	}
}

// IsValid verifica se o papel é válido
func (tr TenantRole) IsValid() bool {
	for _, valid := range AllTenantRoles() {
		if tr == valid {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer
func (tr TenantRole) String() string {
	return string(tr)
}

// ===== GROUP ROLES =====

// GroupRole representa papéis no grupo
type GroupRole string

const (
	GroupRoleOwner  GroupRole = "owner"
	GroupRoleAdmin  GroupRole = "admin"
	GroupRoleMember GroupRole = "member"
)

// AllGroupRoles retorna todos os papéis válidos
func AllGroupRoles() []GroupRole {
	return []GroupRole{
		GroupRoleOwner, GroupRoleAdmin, GroupRoleMember,
	}
}

// IsValid verifica se o papel é válido
func (gr GroupRole) IsValid() bool {
	for _, valid := range AllGroupRoles() {
		if gr == valid {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer
func (gr GroupRole) String() string {
	return string(gr)
}

// ===== PLAN FEATURES =====

// PlanFeature representa recursos do plano
type PlanFeature string

const (
	PlanFeatureBasic               PlanFeature = "basic"
	PlanFeatureAdvancedPermissions PlanFeature = "advanced_permissions"
	PlanFeatureCrossTenantSharing  PlanFeature = "cross_tenant_sharing"
	PlanFeatureAuditLogs           PlanFeature = "audit_logs"
	PlanFeatureBackup              PlanFeature = "backup"
	PlanFeatureExternalSharing     PlanFeature = "external_sharing"
	PlanFeatureVaultLimit          PlanFeature = "vault_limit"
	PlanFeatureUserLimit           PlanFeature = "user_limit"
	PlanFeatureUnlimitedVaults     PlanFeature = "unlimited_vaults"
	PlanFeatureUnlimitedUsers      PlanFeature = "unlimited_users"
	PlanFeatureAPIAccess           PlanFeature = "api_access"
	PlanFeatureGroupManagement     PlanFeature = "group_management"
	PlanFeatureSSO                 PlanFeature = "sso"
	PlanFeatureAdvancedSecurity    PlanFeature = "advanced_security"
	PlanFeatureBasicSharing        PlanFeature = "basic_sharing"
	PlanFeatureAdvancedSharing     PlanFeature = "advanced_sharing"
)

// AllPlanFeatures retorna todos os recursos válidos
func AllPlanFeatures() []PlanFeature {
	return []PlanFeature{
		PlanFeatureBasic, PlanFeatureAdvancedPermissions, PlanFeatureCrossTenantSharing,
		PlanFeatureAuditLogs, PlanFeatureBackup, PlanFeatureExternalSharing,
		PlanFeatureVaultLimit, PlanFeatureUserLimit, PlanFeatureUnlimitedVaults,
		PlanFeatureUnlimitedUsers, PlanFeatureAPIAccess, PlanFeatureGroupManagement,
		PlanFeatureSSO, PlanFeatureAdvancedSecurity, PlanFeatureBasicSharing,
		PlanFeatureAdvancedSharing,
	}
}

// IsValid verifica se o recurso é válido
func (pf PlanFeature) IsValid() bool {
	for _, valid := range AllPlanFeatures() {
		if pf == valid {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer
func (pf PlanFeature) String() string {
	return string(pf)
}

// ===== TOKEN PERMISSIONS =====

// TokenPermission representa permissões de token
type TokenPermission string

const (
	TokenPermissionUse          TokenPermission = "can_use"
	TokenPermissionReadSecrets  TokenPermission = "can_read_secrets"
	TokenPermissionWriteSecrets TokenPermission = "can_write_secrets"
	TokenPermissionManageVault  TokenPermission = "can_manage_vault"
	TokenPermissionRevoke       TokenPermission = "can_revoke"
	TokenPermissionRegenerate   TokenPermission = "can_regenerate"
)

// AllTokenPermissions retorna todas as permissões válidas
func AllTokenPermissions() []TokenPermission {
	return []TokenPermission{
		TokenPermissionUse, TokenPermissionReadSecrets, TokenPermissionWriteSecrets,
		TokenPermissionManageVault, TokenPermissionRevoke, TokenPermissionRegenerate,
	}
}

// IsValid verifica se a permissão é válida
func (tp TokenPermission) IsValid() bool {
	for _, valid := range AllTokenPermissions() {
		if tp == valid {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer
func (tp TokenPermission) String() string {
	return string(tp)
}

// ===== REQUEST/RESPONSE TYPES =====

// CheckRequest representa uma solicitação de verificação
type CheckRequest struct {
	User     string `json:"user"`
	Relation string `json:"relation"`
	Object   string `json:"object"`
}

// Validate valida a solicitação
func (cr *CheckRequest) Validate() error {
	if cr.User == "" {
		return errors.New("user cannot be empty")
	}
	if cr.Relation == "" {
		return errors.New("relation cannot be empty")
	}
	if cr.Object == "" {
		return errors.New("object cannot be empty")
	}

	// Validar formato do usuário
	if !strings.HasPrefix(cr.User, "user:") && !strings.HasPrefix(cr.User, "token:") {
		return errors.New("user must start with 'user:' or 'token:'")
	}

	// Validar formato do objeto
	parts := strings.Split(cr.Object, ":")
	if len(parts) != 2 {
		return errors.New("object must be in format 'type:id'")
	}

	return nil
}

// CheckResponse representa a resposta de uma verificação
type CheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// WriteRequest representa uma solicitação para criar tuplas
type WriteRequest struct {
	Tuples []Tuple `json:"tuples"`
}

// Validate valida a solicitação de escrita
func (wr *WriteRequest) Validate() error {
	if len(wr.Tuples) == 0 {
		return errors.New("tuples cannot be empty")
	}

	for i, tuple := range wr.Tuples {
		if err := tuple.Validate(); err != nil {
			return fmt.Errorf("tuple %d: %w", i, err)
		}
	}

	return nil
}

// DeleteRequest representa uma solicitação para deletar tuplas
type DeleteRequest struct {
	Tuples []Tuple `json:"tuples"`
}

// Validate valida a solicitação de deleção
func (dr *DeleteRequest) Validate() error {
	if len(dr.Tuples) == 0 {
		return errors.New("tuples cannot be empty")
	}

	for i, tuple := range dr.Tuples {
		if err := tuple.Validate(); err != nil {
			return fmt.Errorf("tuple %d: %w", i, err)
		}
	}

	return nil
}

// Tuple representa uma tupla OpenFGA
type Tuple struct {
	User     string `json:"user"`
	Relation string `json:"relation"`
	Object   string `json:"object"`
}

// String implementa fmt.Stringer para Tuple
func (t Tuple) String() string {
	return fmt.Sprintf("%s#%s@%s", t.User, t.Relation, t.Object)
}

// Validate valida a tupla
func (t *Tuple) Validate() error {
	if t.User == "" {
		return errors.New("user cannot be empty")
	}
	if t.Relation == "" {
		return errors.New("relation cannot be empty")
	}
	if t.Object == "" {
		return errors.New("object cannot be empty")
	}
	return nil
}

// ListObjectsRequest representa uma solicitação para listar objetos
type ListObjectsRequest struct {
	User     string                 `json:"user"`
	Relation string                 `json:"relation"`
	Type     string                 `json:"type"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// Validate valida a solicitação de listagem
func (lor *ListObjectsRequest) Validate() error {
	if lor.User == "" {
		return errors.New("user cannot be empty")
	}
	if lor.Relation == "" {
		return errors.New("relation cannot be empty")
	}
	if lor.Type == "" {
		return errors.New("type cannot be empty")
	}

	// Validar formato do usuário
	if !strings.HasPrefix(lor.User, "user:") && !strings.HasPrefix(lor.User, "token:") {
		return errors.New("user must start with 'user:' or 'token:'")
	}

	return nil
}

// ListObjectsResponse representa a resposta de uma listagem de objetos
type ListObjectsResponse struct {
	Objects  []string      `json:"objects"`
	Duration time.Duration `json:"duration"`
}

// ===== HEALTH CHECK TYPES =====

// HealthState representa o estado de saúde do serviço
type HealthState string

const (
	HealthStateHealthy   HealthState = "healthy"
	HealthStateUnhealthy HealthState = "unhealthy"
	HealthStateDegraded  HealthState = "degraded"
)

// String implementa fmt.Stringer
func (hs HealthState) String() string {
	return string(hs)
}

// HealthCheckResponse representa a resposta de um health check
type HealthCheckResponse struct {
	Status    HealthState   `json:"status"`
	Message   string        `json:"message"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// ===== BATCH OPERATIONS =====

// BatchCheckResponse representa a resposta de verificações em lote
type BatchCheckResponse struct {
	Results  []*CheckResponse `json:"results"`
	Duration time.Duration    `json:"duration"`
	Error    string           `json:"error,omitempty"`
}

// ===== ENHANCED CHECK TYPES =====

// CheckRequestWithContext representa uma solicitação de verificação com contexto
type CheckRequestWithContext struct {
	User     string                 `json:"user"`
	Relation string                 `json:"relation"`
	Object   string                 `json:"object"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// Validate valida a solicitação
func (cr *CheckRequestWithContext) Validate() error {
	if cr.User == "" {
		return errors.New("user cannot be empty")
	}
	if cr.Relation == "" {
		return errors.New("relation cannot be empty")
	}
	if cr.Object == "" {
		return errors.New("object cannot be empty")
	}

	// Validar formato do usuário
	if !strings.HasPrefix(cr.User, "user:") && !strings.HasPrefix(cr.User, "token:") {
		return errors.New("user must start with 'user:' or 'token:'")
	}

	// Validar formato do objeto
	parts := strings.Split(cr.Object, ":")
	if len(parts) != 2 {
		return errors.New("object must be in format 'type:id'")
	}

	return nil
}

// CheckResponseWithDuration representa a resposta de uma verificação com duração
type CheckResponseWithDuration struct {
	Allowed  bool          `json:"allowed"`
	Reason   string        `json:"reason,omitempty"`
	Duration time.Duration `json:"duration"`
}

// ===== CACHE TYPES =====

// CacheEntry representa uma entrada no cache
type CacheEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
}

// IsExpired verifica se a entrada expirou
func (ce *CacheEntry) IsExpired() bool {
	return time.Now().After(ce.ExpiresAt)
}

// CacheStats representa estatísticas do cache
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Evictions   int64     `json:"evictions"`
	Size        int64     `json:"size"`
	MaxSize     int64     `json:"max_size"`
	HitRate     float64   `json:"hit_rate"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// ===== METRICS TYPES =====

// MetricData representa dados de métricas
type MetricData struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Tags      map[string]string      `json:"tags,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MetricsSnapshot representa um snapshot das métricas
type MetricsSnapshot struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// ===== HELPER FUNCTIONS =====

// FormatUser formata um usuário para o formato OpenFGA
func FormatUser(userID string) string {
	return FormatObject("user", userID)
}

// FormatToken formata um token para o formato OpenFGA
func FormatToken(tokenID string) string {
	return FormatObject("token", tokenID)
}

// FormatTenant formata um tenant para o formato OpenFGA
func FormatTenant(tenantID string) string {
	return FormatObject("tenant", tenantID)
}

// FormatVault formata um vault para o formato OpenFGA
func FormatVault(vaultID string) string {
	return FormatObject("vault", vaultID)
}

// FormatSecret formata um secret para o formato OpenFGA
func FormatSecret(secretID string) string {
	return FormatObject("secret", secretID)
}

// FormatGroup formata um grupo para o formato OpenFGA
func FormatGroup(groupID string) string {
	return FormatObject("group", groupID)
}

// FormatObject formata um objeto para o formato OpenFGA
func FormatObject(objectType, objectID string) string {
	return fmt.Sprintf("%s:%s", objectType, objectID)
}

// ParseObject extrai tipo e ID de um objeto OpenFGA
func ParseObject(object string) (objectType, objectID string, err error) {
	parts := strings.Split(object, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid object format: %s", object)
	}
	return parts[0], parts[1], nil
}

// ParseUser extrai o ID do usuário de um formato OpenFGA
func ParseUser(user string) (userID string, err error) {
	if strings.HasPrefix(user, "user:") {
		return strings.TrimPrefix(user, "user:"), nil
	}
	if strings.HasPrefix(user, "token:") {
		return strings.TrimPrefix(user, "token:"), nil
	}
	return "", fmt.Errorf("invalid user format: %s", user)
}

// ===== AUDIT EVENT TYPES =====

// PermissionCheckEvent representa um evento de verificação de permissão
type PermissionCheckEvent struct {
	User      string        `json:"user"`
	Relation  string        `json:"relation"`
	Object    string        `json:"object"`
	Result    string        `json:"result"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

// PermissionGrantEvent representa um evento de concessão de permissão
type PermissionGrantEvent struct {
	Grantor   string    `json:"grantor"`
	Grantee   string    `json:"grantee"`
	Relation  string    `json:"relation"`
	Object    string    `json:"object"`
	Timestamp time.Time `json:"timestamp"`
}

// PermissionRevokeEvent representa um evento de revogação de permissão
type PermissionRevokeEvent struct {
	Revoker   string    `json:"revoker"`
	Revokee   string    `json:"revokee"`
	Relation  string    `json:"relation"`
	Object    string    `json:"object"`
	Timestamp time.Time `json:"timestamp"`
}

// SuspiciousActivityEvent representa atividade suspeita
type SuspiciousActivityEvent struct {
	User      string    `json:"user"`
	Activity  string    `json:"activity"`
	Details   string    `json:"details"`
	Severity  string    `json:"severity"`
	ClientIP  string    `json:"client_ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditEvent representa um evento de auditoria genérico
type AuditEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	Metadata  map[string]interface{} `json:"metadata"`
	RequestID string                 `json:"request_id"`
	ClientIP  string                 `json:"client_ip"`
	UserAgent string                 `json:"user_agent"`
	Duration  time.Duration          `json:"duration"`
}

// ===== HELPER FUNCTIONS FOR PERMISSIONS =====

// vaultPermissionToRelation converts vault permission to relation
func vaultPermissionToRelation(permission VaultPermission) string {
	return string(permission)
}

// secretPermissionToRelation converts secret permission to relation
func secretPermissionToRelation(permission SecretPermission) string {
	return string(permission)
}

// tenantPermissionToRelation converts tenant permission to relation
func tenantPermissionToRelation(permission TenantRole) string {
	return string(permission)
}

// tokenPermissionToRelation converts token permission to relation
func tokenPermissionToRelation(permission TokenPermission) string {
	return string(permission)
}

// GenerateTokenID generates a unique token ID
func GenerateTokenID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// NewAuditEvent creates a new audit event
func NewAuditEvent(userID, action, resource string, allowed bool, context map[string]interface{}) *AuditEvent {
	result := "allowed"
	if !allowed {
		result = "denied"
	}

	return &AuditEvent{
		ID:        GenerateTokenID(),
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Result:    result,
		Metadata:  context,
	}
}

// ===== ADDITIONAL TENANT PERMISSIONS =====

// TenantPermission represents granular tenant permissions
type TenantPermission string

const (
	TenantPermissionView           TenantPermission = "can_view"
	TenantPermissionCreateVault    TenantPermission = "can_create_vault"
	TenantPermissionManageUsers    TenantPermission = "can_manage_users"
	TenantPermissionManageGroups   TenantPermission = "can_manage_groups"
	TenantPermissionManageTokens   TenantPermission = "can_manage_tokens"
	TenantPermissionViewAudit      TenantPermission = "can_view_audit"
	TenantPermissionManageSettings TenantPermission = "can_manage_settings"
	TenantPermissionExternalShare  TenantPermission = "can_external_share"
	TenantPermissionOwner          TenantPermission = "owner"
	TenantPermissionAdmin          TenantPermission = "admin"
)

// AllTenantPermissions returns all valid tenant permissions
func AllTenantPermissions() []TenantPermission {
	return []TenantPermission{
		TenantPermissionView, TenantPermissionCreateVault, TenantPermissionManageUsers,
		TenantPermissionManageGroups, TenantPermissionManageTokens, TenantPermissionViewAudit,
		TenantPermissionManageSettings, TenantPermissionExternalShare, TenantPermissionOwner,
		TenantPermissionAdmin,
	}
}

// IsValid checks if the tenant permission is valid
func (tp TenantPermission) IsValid() bool {
	for _, valid := range AllTenantPermissions() {
		if tp == valid {
			return true
		}
	}
	return false
}

// String implements fmt.Stringer for TenantPermission
func (tp TenantPermission) String() string {
	return string(tp)
}

// WriteResponse represents the response from a write operation
type WriteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// DeleteResponse represents the response from a delete operation
type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// AuditQuery represents a query for audit events
type AuditQuery struct {
	UserID    string            `json:"user_id,omitempty"`
	Action    string            `json:"action,omitempty"`
	Resource  string            `json:"resource,omitempty"`
	StartTime *time.Time        `json:"start_time,omitempty"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
}

// ===== PLAN TYPES =====

// PlanType representa os tipos de plano
type PlanType string

const (
	PlanFree       PlanType = "free"
	PlanPro        PlanType = "pro"
	PlanEnterprise PlanType = "enterprise"
)

// AllPlanTypes retorna todos os tipos de plano válidos
func AllPlanTypes() []PlanType {
	return []PlanType{PlanFree, PlanPro, PlanEnterprise}
}

// IsValid verifica se o tipo de plano é válido
func (pt PlanType) IsValid() bool {
	for _, valid := range AllPlanTypes() {
		if pt == valid {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer
func (pt PlanType) String() string {
	return string(pt)
}
