package entity

import (
	"time"

	"lockari-api-application/pkg/utils"
)

// Vault representa um cofre de secrets, chaves, certificados ou chaves SSH
// É um contêiner lógico para itens sensíveis de um tenant específico.
type Vault struct {
	ID                   string                 `json:"id" firestore:"id"`
	Name                 string                 `json:"name" firestore:"name"`
	Description          string                 `json:"description,omitempty" firestore:"description,omitempty"`
	TenantID             string                 `json:"tenant_id" firestore:"tenant_id"`         // ID do tenant ao qual este vault pertence
	OwnerUserID          string                 `json:"owner_user_id" firestore:"owner_user_id"` // O UID do Firebase do usuário que criou/é owner principal
	CreatedAt            time.Time              `json:"created_at" firestore:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at" firestore:"updated_at"`
	DeletedAt            *time.Time             `json:"deleted_at,omitempty" firestore:"deleted_at,omitempty"`                             // Para soft delete
	ItemCount            int                    `json:"item_count" firestore:"item_count"`                                                 // Contador de itens dentro do vault (para limites de plano)
	LastAccessedAt       *time.Time             `json:"last_accessed_at,omitempty" firestore:"last_accessed_at,omitempty"`                 // Opcional: para auditoria/dashboard
	LastAccessedByUserID *string                `json:"last_accessed_by_user_id,omitempty" firestore:"last_accessed_by_user_id,omitempty"` // Opcional: para auditoria/dashboard
	Tags                 []string               `json:"tags,omitempty" firestore:"tags,omitempty"`                                         // Tags associadas ao vault
	Metadata             map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty"`                                 // Para metadados adicionais, se necessário
}

func NewDefaultVault(tenantID, userID *string) *Vault {
	if tenantID == nil || *tenantID == "" {
		panic("tenantID cannot be nil or empty")
	}

	if userID == nil || *userID == "" {
		panic("userID cannot be nil or empty")
	}

	return &Vault{
		ID:          utils.GenerateIDv7(),
		Name:        "Default Vault",
		Description: "This is the default vault for the tenant.",
		TenantID:    *tenantID, // Será definido quando o vault for associado a um tenant
		OwnerUserID: *userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ItemCount:   0,
		Tags:        []string{},
		Metadata:    make(map[string]interface{}),
	}
}
