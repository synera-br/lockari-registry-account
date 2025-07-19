package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"lockari-api-application/pkg/authorization"
)

func demonstratePlanLimitsExample() {
	// Simulação de uso do sistema de limites
	demonstratePlanLimits()
}

func demonstratePlanLimits() {
	// Simular conexão com banco (opcional)
	var db *sql.DB // pode ser nil para usar apenas valores padrão

	// Criar serviço de limites
	planService := authorization.NewPlanLimitService(db)

	ctx := context.Background()

	// Cenário 1: Usuário Free tentando criar 4º vault
	fmt.Println("=== CENÁRIO 1: USUÁRIO FREE ===")
	tenantID := "tenant-user-123"
	planType := authorization.PlanFree

	// Verificar limites atuais
	limits, err := planService.GetPlanLimits(ctx, tenantID, planType)
	if err != nil {
		log.Printf("Erro ao obter limites: %v", err)
		return
	}

	fmt.Printf("Plano: %s\n", planType)
	fmt.Printf("Limite de Vaults: %d\n", limits.VaultLimit)
	fmt.Printf("Limite de Usuários: %d\n", limits.UserLimit)
	fmt.Printf("Ilimitado: %v\n", limits.IsUnlimited)

	// Verificar se pode criar mais vaults
	currentVaults := 3
	canCreate, err := planService.IsWithinLimits(ctx, tenantID, planType, "vault", currentVaults)
	if err != nil {
		log.Printf("Erro ao verificar limites: %v", err)
		return
	}

	if canCreate {
		fmt.Printf("✅ Pode criar mais vaults (atual: %d)\n", currentVaults)
	} else {
		fmt.Printf("❌ Limite de vaults atingido (atual: %d)\n", currentVaults)
	}

	// Verificar quantos restam
	remaining, err := planService.GetRemainingLimits(ctx, tenantID, planType, "vault", currentVaults)
	if err != nil {
		log.Printf("Erro ao obter limites restantes: %v", err)
		return
	}

	if remaining == -1 {
		fmt.Printf("♾️ Vaults restantes: Ilimitado\n")
	} else {
		fmt.Printf("📊 Vaults restantes: %d\n", remaining)
	}

	// Verificar se tem funcionalidade específica
	hasAPI, err := planService.HasFeature(ctx, tenantID, planType, authorization.PlanFeatureAPIAccess)
	if err != nil {
		log.Printf("Erro ao verificar funcionalidade: %v", err)
		return
	}

	if hasAPI {
		fmt.Printf("✅ Tem acesso à API\n")
	} else {
		fmt.Printf("❌ Não tem acesso à API\n")
	}

	fmt.Println()

	// Cenário 2: Usuário Pro
	fmt.Println("=== CENÁRIO 2: USUÁRIO PRO ===")
	tenantID2 := "tenant-company-456"
	planType2 := authorization.PlanPro

	limits2, err := planService.GetPlanLimits(ctx, tenantID2, planType2)
	if err != nil {
		log.Printf("Erro ao obter limites: %v", err)
		return
	}

	fmt.Printf("Plano: %s\n", planType2)
	fmt.Printf("Limite de Vaults: %d\n", limits2.VaultLimit)
	fmt.Printf("Limite de Usuários: %d\n", limits2.UserLimit)

	// Verificar API
	hasAPI2, err := planService.HasFeature(ctx, tenantID2, planType2, authorization.PlanFeatureAPIAccess)
	if err != nil {
		log.Printf("Erro ao verificar funcionalidade: %v", err)
		return
	}

	if hasAPI2 {
		fmt.Printf("✅ Tem acesso à API\n")
	} else {
		fmt.Printf("❌ Não tem acesso à API\n")
	}

	// Verificar auditoria
	hasAudit, err := planService.HasFeature(ctx, tenantID2, planType2, authorization.PlanFeatureAuditLogs)
	if err != nil {
		log.Printf("Erro ao verificar funcionalidade: %v", err)
		return
	}

	if hasAudit {
		fmt.Printf("✅ Tem logs de auditoria\n")
	} else {
		fmt.Printf("❌ Não tem logs de auditoria\n")
	}

	fmt.Println()

	// Cenário 3: Usuário Enterprise
	fmt.Println("=== CENÁRIO 3: USUÁRIO ENTERPRISE ===")
	tenantID3 := "tenant-enterprise-789"
	planType3 := authorization.PlanEnterprise

	limits3, err := planService.GetPlanLimits(ctx, tenantID3, planType3)
	if err != nil {
		log.Printf("Erro ao obter limites: %v", err)
		return
	}

	fmt.Printf("Plano: %s\n", planType3)
	fmt.Printf("Ilimitado: %v\n", limits3.IsUnlimited)

	// Verificar se pode criar 1000 vaults
	currentVaults3 := 1000
	canCreate3, err := planService.IsWithinLimits(ctx, tenantID3, planType3, "vault", currentVaults3)
	if err != nil {
		log.Printf("Erro ao verificar limites: %v", err)
		return
	}

	if canCreate3 {
		fmt.Printf("✅ Pode criar mais vaults (atual: %d)\n", currentVaults3)
	} else {
		fmt.Printf("❌ Limite de vaults atingido (atual: %d)\n", currentVaults3)
	}

	// Verificar SSO
	hasSSO, err := planService.HasFeature(ctx, tenantID3, planType3, authorization.PlanFeatureSSO)
	if err != nil {
		log.Printf("Erro ao verificar funcionalidade: %v", err)
		return
	}

	if hasSSO {
		fmt.Printf("✅ Tem SSO\n")
	} else {
		fmt.Printf("❌ Não tem SSO\n")
	}

	fmt.Println()

	// Cenário 4: Customização específica (necessita banco)
	fmt.Println("=== CENÁRIO 4: CUSTOMIZAÇÃO ESPECÍFICA ===")
	if db != nil {
		// Criar limites customizados para um cliente específico
		customLimits := authorization.PlanLimits{
			VaultLimit:  100, // Cliente especial tem 100 vaults
			UserLimit:   25,  // com 25 usuários
			IsUnlimited: false,
			Features: []authorization.PlanFeature{
				authorization.PlanFeatureBasic,
				authorization.PlanFeatureVaultLimit,
				authorization.PlanFeatureUserLimit,
				authorization.PlanFeatureAPIAccess,
				authorization.PlanFeatureAuditLogs,
				authorization.PlanFeatureGroupManagement,
				authorization.PlanFeatureSSO, // SSO customizado no plano Pro
			},
			Description: "Plano Pro customizado para cliente VIP",
		}

		err = planService.SetCustomLimits(ctx, tenantID2, planType2, customLimits)
		if err != nil {
			log.Printf("Erro ao definir limites customizados: %v", err)
			return
		}

		fmt.Printf("✅ Limites customizados definidos para %s\n", tenantID2)

		// Verificar os novos limites
		newLimits, err := planService.GetPlanLimits(ctx, tenantID2, planType2)
		if err != nil {
			log.Printf("Erro ao obter novos limites: %v", err)
			return
		}

		fmt.Printf("Novos limites - Vaults: %d, Usuários: %d\n", newLimits.VaultLimit, newLimits.UserLimit)

		// Verificar se agora tem SSO
		hasCustomSSO, err := planService.HasFeature(ctx, tenantID2, planType2, authorization.PlanFeatureSSO)
		if err != nil {
			log.Printf("Erro ao verificar funcionalidade customizada: %v", err)
			return
		}

		if hasCustomSSO {
			fmt.Printf("✅ Agora tem SSO customizado\n")
		} else {
			fmt.Printf("❌ Ainda não tem SSO\n")
		}
	} else {
		fmt.Printf("ℹ️ Customização requer conexão com banco de dados\n")
	}
}

// Exemplo de uso em um handler HTTP
func createVaultHandler(planService *authorization.PlanLimitService) {
	// Simular um handler Gin
	fmt.Println("\n=== EXEMPLO DE HANDLER HTTP ===")

	ctx := context.Background()
	tenantID := "tenant-user-123"
	planType := authorization.PlanFree

	// Verificar se pode criar vault
	currentVaults := 3 // Obtido do banco
	canCreate, err := planService.IsWithinLimits(ctx, tenantID, planType, "vault", currentVaults)
	if err != nil {
		fmt.Printf("❌ Erro interno: %v\n", err)
		return
	}

	if !canCreate {
		fmt.Printf("❌ HTTP 403: Limite de vaults atingido\n")
		return
	}

	// Criar vault
	fmt.Printf("✅ HTTP 201: Vault criado com sucesso\n")
}

// Exemplo de middleware para verificar funcionalidades
func featureMiddleware(planService *authorization.PlanLimitService, requiredFeature authorization.PlanFeature) {
	fmt.Println("\n=== EXEMPLO DE MIDDLEWARE ===")

	ctx := context.Background()
	tenantID := "tenant-user-123"
	planType := authorization.PlanFree

	hasFeature, err := planService.HasFeature(ctx, tenantID, planType, requiredFeature)
	if err != nil {
		fmt.Printf("❌ Erro interno: %v\n", err)
		return
	}

	if !hasFeature {
		fmt.Printf("❌ HTTP 403: Funcionalidade '%s' não disponível no plano %s\n", requiredFeature, planType)
		return
	}

	fmt.Printf("✅ Funcionalidade '%s' autorizada\n", requiredFeature)
}
