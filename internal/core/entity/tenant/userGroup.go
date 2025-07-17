package entity

import (
	"time"

	"github.com/synera-br/lockari-backend-app/pkg/utils"
)

// GroupMember representa um membro de um grupo de usuários
// Este é o metadado do membro. A associação com users e tenants/vaults
// é feita através de tuplas no OpenFGA e outras coleções no Firestore.
type GroupMember struct {
	ID        string `json:"id"`         // ID do Grupo
	GroupName string `json:"group_name"` // Nome do grupo
	Name      string `json:"name"`       // Nome do usuário no grupo
}

// UserGroup representa um grupo de usuários armazenado no Firestore
// Este é o metadado do grupo. A associação com users e tenants/vaults
// é feita através de tuplas no OpenFGA e outras coleções no Firestore.
type UserGroup struct {
	ID              string                 `json:"id" firestore:"id"`
	Name            string                 `json:"name" firestore:"name"`
	Description     string                 `json:"description,omitempty" firestore:"description,omitempty"`
	CreatedAt       time.Time              `json:"created_at" firestore:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" firestore:"updated_at"`
	DeletedAt       *time.Time             `json:"deleted_at,omitempty" firestore:"deleted_at,omitempty"` // Para soft delete
	CreatedByUserID string                 `json:"created_by_user_id" firestore:"created_by_user_id"`     // ID do usuário que criou o grupo
	TenantID        string                 `json:"tenant_id" firestore:"tenant_id"`                       // ID do tenant a que este grupo pertence
	MemberCount     int                    `json:"member_count" firestore:"member_count"`                 // Contador de membros (mantido pelo backend)
	GroupType       TenantGroupType        `json:"group_type" firestore:"group_type"`                     // O tipo funcional/papel deste grupo
	Metadata        map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty"`     // Para metadados adicionais, ex: "is_default_group": true
}

func NewDefaultUserGroup() *UserGroup {

	return &UserGroup{
		ID:          utils.GenerateIDv7(),
		Name:        "Default Owner Group",
		Description: "This is the default group for tenant owners.",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		GroupType:   TenantGroupOwner,
		Metadata:    map[string]interface{}{"is_default_group": true},
	}
}

func (g *UserGroup) GetID() string {
	return g.ID
}

func (g *UserGroup) GetName() string {
	return g.Name
}

func (g *UserGroup) GetDescription() string {
	if g.Description == "" {
		return "No description provided"
	}
	return g.Description
}

func (g *UserGroup) GetCreatedAt() time.Time {
	return g.CreatedAt
}

func (g *UserGroup) GetUpdatedAt() time.Time {
	return g.UpdatedAt
}

func (g *UserGroup) GetGroupType() TenantGroupType {
	return g.GroupType
}

func (g *UserGroup) GetMetadata() map[string]interface{} {
	return g.Metadata
}
