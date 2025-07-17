package authorization

import (
	"context"
	"fmt"
	"strings"
)

// LockariService implementa a interface LockariAuthorizationService
type LockariService struct {
	service      *Service
	auditService AuditService
	config       *Config
	logger       Logger
}

// LockariServiceOptions define opções para criar um LockariService
type LockariServiceOptions struct {
	Service *Service
	// AuditService AuditService
	Config *Config
	Logger Logger
}

// NewLockariService cria uma nova instância do LockariService
func NewLockariService(opts LockariServiceOptions) *LockariService {
	return &LockariService{
		service: opts.Service,
		// auditService: opts.AuditService,
		config: opts.Config,
		logger: opts.Logger,
	}
}

// NewLockariAuthorizationService cria uma nova instância do LockariService
func NewLockariAuthorizationService(opts LockariServiceOptions) LockariAuthorizationService {
	return &LockariService{
		service: opts.Service,
		// auditService: opts.AuditService,
		config: opts.Config,
		logger: opts.Logger,
	}
}

// ===== IMPLEMENTING AuthorizationService INTERFACE =====

// Check verifica uma permissão única (delegado para o serviço básico)
func (ls *LockariService) Check(ctx context.Context, req *CheckRequest) (*CheckResponse, error) {
	if ls.service != nil {
		return ls.service.Check(ctx, req)
	}
	return &CheckResponse{Allowed: false}, nil
}

// CheckBatch verifica múltiplas permissões em uma chamada (delegado para o serviço básico)
func (ls *LockariService) CheckBatch(ctx context.Context, reqs []*CheckRequest) (*BatchCheckResponse, error) {
	if ls.service != nil {
		return ls.service.CheckBatch(ctx, reqs)
	}
	return &BatchCheckResponse{}, nil
}

// ListObjects lista todos os objetos que o usuário pode acessar (delegado para o serviço básico)
func (ls *LockariService) ListObjects(ctx context.Context, req *ListObjectsRequest) (*ListObjectsResponse, error) {
	if ls.service != nil {
		return ls.service.ListObjects(ctx, req)
	}
	return &ListObjectsResponse{Objects: []string{}}, nil
}

// Write cria relacionamentos (tuplas) (delegado para o serviço básico)
func (ls *LockariService) Write(ctx context.Context, req *WriteRequest) error {
	// if ls.service != nil {
	// 	return ls.service.Write(ctx, req)
	// }
	return nil
}

// Delete remove relacionamentos (delegado para o serviço básico)
func (ls *LockariService) Delete(ctx context.Context, req *DeleteRequest) error {
	// if ls.service != nil {
	// 	return ls.service.Delete(ctx, req)
	// }
	return nil
}

// Health verifica se o OpenFGA está disponível (delegado para o serviço básico)
func (ls *LockariService) Health(ctx context.Context) error {
	// if ls.service != nil {
	// 	return ls.service.Health(ctx)
	// }
	return nil
}

// ===== IMPLEMENTING LockariAuthorizationService SPECIFIC METHODS =====

// === VAULT OPERATIONS ===

// CanAccessVault verifica se o usuário pode acessar um vault
func (ls *LockariService) CanAccessVault(ctx context.Context, userID, vaultID string, permission VaultPermission) (bool, error) {
	if !permission.IsValid() {
		return false, fmt.Errorf("invalid vault permission: %s", permission)
	}

	req := &CheckRequest{
		User:     formatUser(userID),
		Relation: string(permission),
		Object:   formatVault(vaultID),
	}

	response, err := ls.Check(ctx, req)
	if err != nil {
		return false, fmt.Errorf("error checking vault permission: %w", err)
	}

	return response.Allowed, nil
}

// SetupVault configura um novo vault com permissões iniciais
func (ls *LockariService) SetupVault(ctx context.Context, vaultID, tenantID, ownerID string) error {
	tuples := []Tuple{
		{
			User:     formatUser(ownerID),
			Relation: "owner",
			Object:   formatVault(vaultID),
		},
		{
			User:     formatTenant(tenantID),
			Relation: "parent",
			Object:   formatVault(vaultID),
		},
	}

	req := &WriteRequest{
		Tuples: tuples,
	}

	return ls.Write(ctx, req)
}

// ShareVault compartilha um vault com outro usuário
func (ls *LockariService) ShareVault(ctx context.Context, vaultID, ownerID, targetUserID string, permission VaultPermission) error {
	tuple := Tuple{
		User:     formatUser(targetUserID),
		Relation: string(permission),
		Object:   formatVault(vaultID),
	}

	req := &WriteRequest{
		Tuples: []Tuple{tuple},
	}

	return ls.Write(ctx, req)
}

// RevokeVaultAccess revoga acesso a um vault
func (ls *LockariService) RevokeVaultAccess(ctx context.Context, vaultID, ownerID, targetUserID string, permission VaultPermission) error {
	tuple := Tuple{
		User:     formatUser(targetUserID),
		Relation: string(permission),
		Object:   formatVault(vaultID),
	}

	req := &DeleteRequest{
		Tuples: []Tuple{tuple},
	}

	return ls.Delete(ctx, req)
}

// ListAccessibleVaults lista todos os vaults acessíveis para o usuário
func (ls *LockariService) ListAccessibleVaults(ctx context.Context, userID string) ([]string, error) {
	req := &ListObjectsRequest{
		User:     formatUser(userID),
		Relation: string(VaultPermissionView),
		Type:     "vault",
	}

	response, err := ls.ListObjects(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error listing accessible vaults: %w", err)
	}

	vaultIDs := make([]string, len(response.Objects))
	for i, object := range response.Objects {
		vaultIDs[i] = extractIDFromObject(object)
	}

	return vaultIDs, nil
}

// === SECRET OPERATIONS ===

// CanAccessSecret verifica se o usuário pode acessar um secret
func (ls *LockariService) CanAccessSecret(ctx context.Context, userID, secretID string, permission SecretPermission) (bool, error) {
	return true, nil // Implementação simplificada
}

// SetupSecret configura um novo secret com permissões iniciais
func (ls *LockariService) SetupSecret(ctx context.Context, secretID, vaultID string) error {
	return nil // Implementação simplificada
}

// ListAccessibleSecrets lista todos os secrets acessíveis para o usuário
func (ls *LockariService) ListAccessibleSecrets(ctx context.Context, userID string) ([]string, error) {
	return []string{}, nil // Implementação simplificada
}

// === TENANT OPERATIONS ===

// SetupTenant configura um novo tenant
func (ls *LockariService) SetupTenant(ctx context.Context, tenantID, ownerID string, features []PlanFeature, relations []string) error {

	var feats []string
	for _, feature := range features {
		feats = append(feats, string(feature))
	}

	tuples := make([]Tuple, len(relations)+1)

	for i, relation := range relations {
		tuples[i] = Tuple{
			User:     formatUser(ownerID),
			Relation: relation,
			Object:   formatTenant(tenantID),
		}
	}

	err := ls.Write(ctx, &WriteRequest{
		Tuples: tuples,
	})

	if err != nil {
		return fmt.Errorf("error setting up tenant: %w", err)
	}

	return nil
}

// AddUserToTenant adiciona um usuário ao tenant
func (ls *LockariService) AddUserToTenant(ctx context.Context, userID, tenantID string, role TenantRole) error {

	return nil // Implementação simplificada
}

// RemoveUserFromTenant remove um usuário do tenant
func (ls *LockariService) RemoveUserFromTenant(ctx context.Context, userID, tenantID string) error {
	return nil // Implementação simplificada
}

// ChangeUserTenantRole altera o papel de um usuário no tenant
func (ls *LockariService) ChangeUserTenantRole(ctx context.Context, userID, tenantID string, newRole TenantRole) error {
	return nil // Implementação simplificada
}

// IsTenantMember verifica se o usuário é membro do tenant
func (ls *LockariService) IsTenantMember(ctx context.Context, userID, tenantID string) (bool, error) {
	return true, nil // Implementação simplificada
}

func (ls *LockariService) CanAssignRoleFromTenant(ctx context.Context, userID, tenantID *string, permission TenantRole) (bool, error) {
	if !permission.IsValid() {
		return false, fmt.Errorf("invalid tenant permission: %s", permission)
	}

	req := &CheckRequest{
		User:     formatUser(*userID),
		Relation: string(permission),
		Object:   formatTenant(*tenantID),
	}

	response, err := ls.Check(ctx, req)
	if err != nil {
		return false, fmt.Errorf("error checking tenant permission: %w", err)
	}

	return response.Allowed, nil
}

func (ls *LockariService) ListPermissionFromTenant(ctx context.Context, tenantID string) (interface{}, error) {

	relations := []string{"owner", "admin", "member", "viewer"}

	for _, relation := range relations {
		req := &ListObjectsRequest{
			User:     formatTenant(tenantID),
			Relation: relation,
		}

		response, err := ls.ListObjects(ctx, req)
		if err != nil {
			fmt.Println("\n [ListPermissionFromTenant] Error listing objects:", err)
		}

		if len(response.Objects) > 0 {
			fmt.Println("\n [ListPermissionFromTenant] Found objects for relation:", relation, "in tenant:", tenantID)
		}
		fmt.Println("\n [ListPermissionFromTenant] Objects for relation:", relation, "in tenant:", tenantID, ":", response.Objects)
	}

	return tenantID, nil
}

// === GROUP OPERATIONS ===

// CreateGroup cria um novo grupo
func (ls *LockariService) CreateGroup(ctx context.Context, groupID, tenantID, ownerID string) error {

	// if groupID == "" {
	// 	return fmt.Errorf("groupID is required")
	// }

	// if tenantID == "" {
	// 	return fmt.Errorf("tenantID is required")
	// }
	// if ownerID == "" {
	// 	return fmt.Errorf("ownerID is required")
	// }

	// tuple := Tuple{
	// 	User:     formatUser(ownerID),
	// 	Relation: "owner",
	// 	Object:   formatGroup(groupID),
	// }

	return nil
}

// AssociateGroupToTenant cria um novo grupo
func (ls *LockariService) AssociateGroupToTenant(ctx context.Context, groupID, tenantID, ownerID, relation string) error {

	if groupID == "" {
		return fmt.Errorf("groupID is required")
	}

	if tenantID == "" {
		return fmt.Errorf("tenantID is required")
	}

	if ownerID == "" {
		return fmt.Errorf("ownerID is required")
	}

	if relation == "" {
		return fmt.Errorf("relation is required")
	}

	tuple := Tuple{
		User:     formatGroupToTenant(groupID, tenantID),
		Relation: relation,
		Object:   formatTenant(tenantID),
	}

	err := ls.Write(ctx, &WriteRequest{
		Tuples: []Tuple{tuple},
	})

	if err != nil {
		return fmt.Errorf("error creating group: %w", err)
	}

	return nil
}

func (ls *LockariService) AssociateUserToGroupAsMember(ctx context.Context, userID, groupID, tenantID string) error {
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	if groupID == "" {
		return fmt.Errorf("groupID is required")
	}
	if tenantID == "" {
		return fmt.Errorf("tenantID is required")
	}

	tuple := Tuple{
		User:     formatUser(userID),
		Relation: "member",
		Object:   formatGroup(groupID),
	}

	err := ls.Write(ctx, &WriteRequest{
		Tuples: []Tuple{tuple},
	})

	if err != nil {
		return fmt.Errorf("error associating user to group: %w", err)
	}

	return nil
}

// AddUserToGroup adiciona um usuário ao grupo
func (ls *LockariService) AddUserToGroup(ctx context.Context, userID, groupID string, role GroupRole) error {

	return nil // Implementação simplificada
}

// RemoveUserFromGroup remove um usuário do grupo
func (ls *LockariService) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	return nil // Implementação simplificada
}

// ChangeUserGroupRole altera o papel de um usuário no grupo
func (ls *LockariService) ChangeUserGroupRole(ctx context.Context, userID, groupID string, newRole GroupRole) error {
	return nil // Implementação simplificada
}

// DeleteGroup deleta um grupo
func (ls *LockariService) DeleteGroup(ctx context.Context, groupID string) error {
	return nil // Implementação simplificada
}

// === TOKEN OPERATIONS ===

// CreateAPIToken cria um token de API com permissões específicas
func (ls *LockariService) CreateAPIToken(ctx context.Context, userID, vaultID string, permissions []TokenPermission) (string, error) {
	return "mock-token-123", nil // Implementação simplificada
}

// CheckTokenPermission verifica se um token tem uma permissão específica
func (ls *LockariService) CheckTokenPermission(ctx context.Context, tokenID string, permission TokenPermission) (bool, error) {
	return true, nil // Implementação simplificada
}

// RevokeToken revoga um token de API
func (ls *LockariService) RevokeToken(ctx context.Context, tokenID string) error {
	return nil // Implementação simplificada
}

// RefreshToken atualiza um token de API
func (ls *LockariService) RefreshToken(ctx context.Context, tokenID string) (string, error) {
	return "new-token-456", nil // Implementação simplificada
}

// ListTokens lista todos os tokens de um usuário
func (ls *LockariService) ListTokens(ctx context.Context, userID string) ([]string, error) {
	return []string{"token-1", "token-2"}, nil // Implementação simplificada
}

// === EXTERNAL SHARING (Enterprise) ===

// InitiateExternalSharing inicia processo de compartilhamento externo
func (ls *LockariService) InitiateExternalSharing(ctx context.Context, userID, vaultID, targetTenantID string) (string, error) {
	return "share-request-789", nil // Implementação simplificada
}

// ApproveExternalSharing aprova compartilhamento externo
func (ls *LockariService) ApproveExternalSharing(ctx context.Context, userID, requestID string) error {
	return nil // Implementação simplificada
}

// RejectExternalSharing rejeita compartilhamento externo
func (ls *LockariService) RejectExternalSharing(ctx context.Context, userID, requestID string) error {
	return nil // Implementação simplificada
}

// ListExternalSharingRequests lista solicitações de compartilhamento externo
func (ls *LockariService) ListExternalSharingRequests(ctx context.Context, tenantID string) ([]string, error) {
	return []string{"request-1", "request-2"}, nil // Implementação simplificada
}

// === AUDIT OPERATIONS ===

// GetAuditLogs recupera logs de auditoria
func (ls *LockariService) GetAuditLogs(ctx context.Context, userID, resource string, limit int) ([]AuditEvent, error) {

	objs, err := ls.ListObjects(ctx, &ListObjectsRequest{
		User:     formatUser(userID),
		Relation: "owner",
		Type:     resource,
	})

	if err != nil {
		return nil, fmt.Errorf("error listing objects for audit logs: %w", err)
	}
	auditEvents := make([]AuditEvent, 0, len(objs.Objects))

	return auditEvents, nil
}

// === CACHE OPERATIONS ===

// InvalidateCache invalida o cache para um usuário específico
func (ls *LockariService) InvalidateCache(ctx context.Context, userID string) error {
	return nil // Implementação simplificada
}

// GetCacheStats recupera estatísticas do cache
func (ls *LockariService) GetCacheStats(ctx context.Context) (CacheStats, error) {
	return CacheStats{}, nil // Implementação simplificada
}

// === HELPER METHODS ===

// formatUser formata um ID de usuário
func formatUser(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

// formatVault formata um ID de vault
func formatVault(vaultID string) string {
	return fmt.Sprintf("vault:%s", vaultID)
}

// formatTenant formata um ID de tenant
func formatTenant(tenantID string) string {
	return fmt.Sprintf("tenant:%s", tenantID)
}

func formatGroup(groupID string) string {
	return fmt.Sprintf("group:%s", groupID)
}

func formatGroupToTenant(groupID, tenantID string) string {
	return fmt.Sprintf("group:%s/%s", groupID, tenantID)
}

// extractIDFromObject extrai o ID de um objeto formatado
func extractIDFromObject(object string) string {
	parts := strings.Split(object, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return object
}

// === HELPER METHODS FOR USER/TENANT SETUP ===

// CreateNewUserTenant configura um novo usuário/tenant com plano específico
func (ls *LockariService) CreateNewUserTenant(ctx context.Context, userID, tenantID string, plan PlanType, relations []string) error {
	// 1. Configurar tenant com plano específico
	features := getPlanFeatures(plan)
	err := ls.SetupTenant(ctx, tenantID, userID, features, relations)
	if err != nil {
		return fmt.Errorf("failed to setup tenant: %w", err)
	}

	// 2. Adicionar usuário como owner do tenant
	err = ls.AddUserToTenant(ctx, userID, tenantID, TenantRoleOwner)
	if err != nil {
		return fmt.Errorf("failed to add user to tenant: %w", err)
	}

	// 3. Configurar recursos específicos do plano
	switch plan {
	case PlanFree:
		return ls.setupFreePlanResources(ctx, userID, tenantID)
	case PlanPro:
		return ls.setupProPlanResources(ctx, userID, tenantID)
	case PlanEnterprise:
		return ls.setupEnterprisePlanResources(ctx, userID, tenantID)
	default:
		return fmt.Errorf("unsupported plan type: %s", plan)
	}
}

// setupFreePlanResources configura recursos para plano gratuito
func (ls *LockariService) setupFreePlanResources(ctx context.Context, userID, tenantID string) error {
	// Vault pessoal gratuito
	defaultVaultID := fmt.Sprintf("vault-%s-personal", userID)
	return ls.SetupVault(ctx, defaultVaultID, tenantID, userID)
}

// setupProPlanResources configura recursos para plano pro
func (ls *LockariService) setupProPlanResources(ctx context.Context, userID, tenantID string) error {
	// Múltiplos vaults iniciais
	vaultIDs := []string{
		fmt.Sprintf("vault-%s-personal", userID),
		fmt.Sprintf("vault-%s-business", userID),
	}

	for _, vaultID := range vaultIDs {
		err := ls.SetupVault(ctx, vaultID, tenantID, userID)
		if err != nil {
			return err
		}
	}

	// Grupo padrão para colaboração
	groupID := fmt.Sprintf("group-%s-team", tenantID)
	err := ls.CreateGroup(ctx, groupID, tenantID, userID)
	if err != nil {
		return err
	}

	return ls.AddUserToGroup(ctx, userID, groupID, GroupRoleOwner)
}

// setupEnterprisePlanResources configura recursos para plano enterprise
func (ls *LockariService) setupEnterprisePlanResources(ctx context.Context, userID, tenantID string) error {
	// Estrutura organizacional completa
	vaultIDs := []string{
		fmt.Sprintf("vault-%s-executive", userID),
		fmt.Sprintf("vault-%s-hr", userID),
		fmt.Sprintf("vault-%s-finance", userID),
		fmt.Sprintf("vault-%s-it", userID),
	}

	for _, vaultID := range vaultIDs {
		err := ls.SetupVault(ctx, vaultID, tenantID, userID)
		if err != nil {
			return err
		}
	}

	// Grupos departamentais
	groupIDs := []string{
		fmt.Sprintf("group-%s-executives", tenantID),
		fmt.Sprintf("group-%s-hr", tenantID),
		fmt.Sprintf("group-%s-finance", tenantID),
		fmt.Sprintf("group-%s-it", tenantID),
	}

	for _, groupID := range groupIDs {
		err := ls.CreateGroup(ctx, groupID, tenantID, userID)
		if err != nil {
			return err
		}

		err = ls.AddUserToGroup(ctx, userID, groupID, GroupRoleOwner)
		if err != nil {
			return err
		}
	}

	return nil
}

// === PLAN TYPE AND FEATURES ===

// // PlanType representa os tipos de plano
// type PlanType string

// const (
// 	PlanFree       PlanType = "free"
// 	PlanPro        PlanType = "pro"
// 	PlanEnterprise PlanType = "enterprise"
// )

// getPlanFeatures retorna as funcionalidades para cada plano
func getPlanFeatures(plan PlanType) []PlanFeature {
	switch plan {
	case PlanFree:
		return []PlanFeature{
			PlanFeatureVaultLimit,   // Limite: 3 vaults
			PlanFeatureBasicSharing, // Compartilhamento básico
		}

	case PlanPro:
		return []PlanFeature{
			PlanFeatureVaultLimit,      // Limite: 50 vaults
			PlanFeatureUserLimit,       // Limite: 10 usuários
			PlanFeatureAdvancedSharing, // Compartilhamento avançado
			PlanFeatureAPIAccess,       // Acesso à API
			PlanFeatureGroupManagement, // Gerenciamento de grupos
			PlanFeatureAuditLogs,       // Logs de auditoria
		}

	case PlanEnterprise:
		return []PlanFeature{
			PlanFeatureUnlimitedVaults,  // Vaults ilimitados
			PlanFeatureUnlimitedUsers,   // Usuários ilimitados
			PlanFeatureExternalSharing,  // Compartilhamento externo
			PlanFeatureAPIAccess,        // Acesso à API
			PlanFeatureGroupManagement,  // Gerenciamento de grupos
			PlanFeatureAuditLogs,        // Logs de auditoria
			PlanFeatureSSO,              // Single Sign-On
			PlanFeatureAdvancedSecurity, // Segurança avançada
		}

	default:
		return []PlanFeature{}
	}
}

// === EXISTING HELPER METHODS ===

// ListAllTenantPermissions lista todas as permissions/relacionamentos de um tenant específico
func (ls *LockariService) ListAllTenantPermissions(ctx context.Context, tenantID string) (*TenantPermissionsReport, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenantID is required")
	}

	report := &TenantPermissionsReport{
		TenantID:         tenantID,
		UserPermissions:  make(map[string][]string),
		GroupPermissions: make(map[string][]string),
		VaultPermissions: make(map[string][]string),
		ResourcesByType:  make(map[string][]string),
	}

	// 1. Buscar todos os vaults do tenant
	vaults, err := ls.listVaultsInTenant(ctx, tenantID)
	if err == nil && len(vaults) > 0 {
		report.ResourcesByType["vaults"] = vaults
		fmt.Printf("Found %d vaults for tenant %s: %v\n", len(vaults), tenantID, vaults)
	}

	// 2. Buscar todos os grupos do tenant
	groups, err := ls.listGroupsInTenant(ctx, tenantID)
	if err == nil && len(groups) > 0 {
		report.ResourcesByType["groups"] = groups
		fmt.Printf("Found %d groups for tenant %s: %v\n", len(groups), tenantID, groups)
	}

	// 3. Buscar todos os secrets do tenant
	secrets, err := ls.listSecretsInTenant(ctx, tenantID)
	if err == nil && len(secrets) > 0 {
		report.ResourcesByType["secrets"] = secrets
		fmt.Printf("Found %d secrets for tenant %s: %v\n", len(secrets), tenantID, secrets)
	}

	return report, nil
}

func (ls *LockariService) listGroupsInTenant(ctx context.Context, tenantID string) ([]string, error) {
	// Buscar grupos onde o tenant é parent
	req := &ListObjectsRequest{
		User:     formatTenant(tenantID),
		Relation: "parent",
		Type:     "group",
	}

	resp, err := ls.ListObjects(ctx, req)
	if err != nil {
		return nil, err
	}

	groups := make([]string, len(resp.Objects))
	for i, object := range resp.Objects {
		groups[i] = extractIDFromObject(object)
	}

	return groups, nil
}

func (ls *LockariService) listVaultsInTenant(ctx context.Context, tenantID string) ([]string, error) {
	// Buscar vaults onde o tenant é parent
	req := &ListObjectsRequest{
		User:     formatTenant(tenantID),
		Relation: "parent",
		Type:     "vault",
	}

	resp, err := ls.ListObjects(ctx, req)
	if err != nil {
		return nil, err
	}

	vaults := make([]string, len(resp.Objects))
	for i, object := range resp.Objects {
		vaults[i] = extractIDFromObject(object)
	}

	return vaults, nil
}

func (ls *LockariService) listSecretsInTenant(ctx context.Context, tenantID string) ([]string, error) {
	// Implementação similar para secrets
	req := &ListObjectsRequest{
		User:     formatTenant(tenantID),
		Relation: "parent",
		Type:     "secret",
	}

	resp, err := ls.ListObjects(ctx, req)
	if err != nil {
		return []string{}, nil // Retorna vazio em caso de erro
	}

	secrets := make([]string, len(resp.Objects))
	for i, object := range resp.Objects {
		secrets[i] = extractIDFromObject(object)
	}

	return secrets, nil
}

type TenantPermissionsReport struct {
	TenantID         string              `json:"tenant_id"`
	VaultPermissions map[string][]string `json:"vault_permissions"`
	GroupPermissions map[string][]string `json:"group_permissions"`
	UserPermissions  map[string][]string `json:"user_permissions"`
	ResourcesByType  map[string][]string `json:"resources_by_type"`
}
