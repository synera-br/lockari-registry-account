package entity

import "strings"

type TenantGroupType string

const (
	TenantGroupOwner   TenantGroupType = "OWNER"
	TenantGroupManager TenantGroupType = "MANAGER"
	TenantGroupWriter  TenantGroupType = "WRITER"
	TenantGroupViewer  TenantGroupType = "VIEWER"
)

type TenantGroup struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	DeletedAt   string            `json:"deleted_at,omitempty"`
	OwnerID     string            `json:"owner_id"`
	OwnerName   string            `json:"owner_name"`
	OwnerEmail  string            `json:"owner_email"`
	MemberCount int               `json:"member_count"`
	GroupType   TenantGroupType   `json:"group_type"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// GetAuthorizationRelationType retorna a string da relação FGA correspondente ao GroupType
// Útil para construir as tuplas no OpenFGA
func (t TenantGroupType) GetAuthorizationRelationType() string {
	switch t {
	case TenantGroupOwner:
		return "tenant_owner_group"
	case TenantGroupManager:
		return "tenant_manager_group"
	case TenantGroupWriter:
		return "tenant_writer_group"
	case TenantGroupViewer:
		return "tenant_viewer_group"
	default:
		return "" // Deve ser tratado como erro ou valor desconhecido
	}
}

// GetAuthorizationRelationType retorna a string da relação FGA correspondente ao GroupType
// Útil para construir as tuplas no OpenFGA
func (t TenantGroupType) GetAuthorizationRelationGroup(permission string) string {

	switch strings.ToUpper(string(permission)) {
	case string(TenantGroupOwner):
		return "tenant_owner_group"
	case string(TenantGroupManager):
		return "tenant_manager_group"
	case string(TenantGroupWriter):
		return "tenant_writer_group"
	case string(TenantGroupViewer):
		return "tenant_viewer_group"
	default:
		return "" // Deve ser tratado como erro ou valor desconhecido
	}
}

type VaultGroupType string

const (
	// Para grupos a nível de Tenant
	TenantOwnerGroup   VaultGroupType = "TENANT_OWNER"
	TenantManagerGroup VaultGroupType = "TENANT_MANAGER"
	TenantWriterGroup  VaultGroupType = "TENANT_WRITER"
	TenantViewerGroup  VaultGroupType = "TENANT_VIEWER"

	// Para grupos a nível de Vault (exemplos)
	VaultOwnerGroup   VaultGroupType = "VAULT_OWNER"
	VaultManagerGroup VaultGroupType = "VAULT_MANAGER"
	VaultAdminGroup   VaultGroupType = "VAULT_ADMIN"
	VaultWriterGroup  VaultGroupType = "VAULT_WRITER"
	VaultViewerGroup  VaultGroupType = "VAULT_VIEWER"

	// Ou um tipo genérico se o grupo não tiver um papel específico pré-definido,
	// mas apenas for um "coleção de usuários" para customização
	CustomGroup VaultGroupType = "CUSTOM"
)

// GetOpenFGARelationTypeForVault mapeia o VaultGroupType para a relação no modelo FGA para Vault
func (gt VaultGroupType) GetOpenFGARelationTypeForVault() string {
	switch gt {
	case VaultOwnerGroup:
		return "vault_owner_group"
	case VaultManagerGroup:
		return "vault_manager_group"
	case VaultAdminGroup:
		return "vault_admin_group" // Note: você tem admin para vault, não manager
	case VaultWriterGroup:
		return "vault_writer_group"
	case VaultViewerGroup:
		return "vault_viewer_group"
	default:
		return "" // Tratar como erro ou tipo desconhecido
	}
}

// GetTenantPermissions retorna as permissões lógicas de ação para o tipo de grupo no Tenant
// Estas devem estar em ALINHAMENTO PERFEITO com as definições de 'can_X_tenant' do OpenFGA.
func (tg *TenantGroupType) GetTenantPermissions() []string {
	switch *tg {
	case TenantGroupOwner:
		// Owner tem permissão para tudo
		return []string{"can_read_tenant", "can_manage_tenant", "can_shared_tenant", "can_update_tenant", "can_delete_tenant"}
	case TenantGroupManager:
		// Manager tem permissão para gerenciar e compartilhar, mas não deletar o tenant
		return []string{"can_read_tenant", "can_manage_tenant", "can_shared_tenant", "can_update_tenant"}
	case TenantGroupWriter:
		// Writer pode ler e fazer operações de escrita/atualização específicas no tenant (se houver)
		return []string{"can_read_tenant", "can_update_tenant"}
	case TenantGroupViewer:
		// Viewer pode apenas ler
		return []string{"can_read_tenant"}
	default:
		return []string{}
	}
}
