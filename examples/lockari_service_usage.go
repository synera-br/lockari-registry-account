package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"lockari-api-application/pkg/authorization"
)

// ExampleLockariServiceUsage demonstra como usar a interface LockariAuthorizationService
func ExampleLockariServiceUsage() {
	fmt.Println("=== Exemplo de Uso da Interface LockariAuthorizationService ===\n")

	// 1. Criar configuração
	config := &authorization.Config{
		APIURL:       "http://localhost:8080",
		StoreID:      "01HXSJ9QXKV8J9DK8V7S9QXKV8",
		Timeout:      30 * time.Second,
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
		AuditEnabled: true,
	}

	// 2. Criar logger
	logger := authorization.NewSlogAdapter(slog.Default())

	// 3. Criar serviço Lockari (implementação da interface)
	lockariService := authorization.NewLockariService(authorization.LockariServiceOptions{
		Config: config,
		Logger: logger,
	})

	// 4. Usar a interface
	demonstrateUsage(lockariService)
}

func demonstrateUsage(authService authorization.LockariAuthorizationService) {
	ctx := context.Background()

	// === EXEMPLO 1: Verificar acesso a vault ===
	fmt.Println("1. Verificando acesso a vault...")
	userID := "user123"
	vaultID := "vault456"

	canRead, err := authService.CanAccessVault(ctx, userID, vaultID, authorization.VaultPermissionRead)
	if err != nil {
		fmt.Printf("   ❌ Erro ao verificar acesso: %v\n", err)
	} else {
		fmt.Printf("   ✅ Usuário %s pode ler vault %s: %v\n", userID, vaultID, canRead)
	}

	// === EXEMPLO 2: Configurar novo vault ===
	fmt.Println("\n2. Configurando novo vault...")
	tenantID := "tenant789"
	ownerID := "owner123"

	err = authService.SetupVault(ctx, vaultID, tenantID, ownerID)
	if err != nil {
		fmt.Printf("   ❌ Erro ao configurar vault: %v\n", err)
	} else {
		fmt.Printf("   ✅ Vault %s configurado com sucesso para tenant %s\n", vaultID, tenantID)
	}

	// === EXEMPLO 3: Compartilhar vault ===
	fmt.Println("\n3. Compartilhando vault...")
	targetUserID := "user456"

	err = authService.ShareVault(ctx, vaultID, ownerID, targetUserID, authorization.VaultPermissionRead)
	if err != nil {
		fmt.Printf("   ❌ Erro ao compartilhar vault: %v\n", err)
	} else {
		fmt.Printf("   ✅ Vault %s compartilhado com usuário %s\n", vaultID, targetUserID)
	}

	// === EXEMPLO 4: Listar vaults acessíveis ===
	fmt.Println("\n4. Listando vaults acessíveis...")

	accessibleVaults, err := authService.ListAccessibleVaults(ctx, userID)
	if err != nil {
		fmt.Printf("   ❌ Erro ao listar vaults: %v\n", err)
	} else {
		fmt.Printf("   ✅ Vaults acessíveis para %s: %v\n", userID, accessibleVaults)
	}

	// === EXEMPLO 5: Verificar saúde do serviço ===
	fmt.Println("\n5. Verificando saúde do serviço...")

	err = authService.Health(ctx)
	if err != nil {
		fmt.Printf("   ❌ Serviço não está saudável: %v\n", err)
	} else {
		fmt.Printf("   ✅ Serviço está funcionando corretamente\n")
	}

	fmt.Println("\n=== Demonstração concluída ===")
}

// ExampleInHandler demonstra como usar o serviço em um handler
func ExampleInHandler() {
	fmt.Println("\n=== Exemplo de Uso em Handler ===\n")

	// Simulando um handler
	type VaultHandler struct {
		authService authorization.LockariAuthorizationService
	}

	// Construtor do handler
	newVaultHandler := func(authService authorization.LockariAuthorizationService) *VaultHandler {
		return &VaultHandler{
			authService: authService,
		}
	}

	// Criar configuração e serviço
	config := &authorization.Config{
		APIURL:       "http://localhost:8080",
		StoreID:      "01HXSJ9QXKV8J9DK8V7S9QXKV8",
		Timeout:      30 * time.Second,
		CacheEnabled: true,
		AuditEnabled: true,
	}

	logger := authorization.NewSlogAdapter(slog.Default())

	lockariService := authorization.NewLockariService(authorization.LockariServiceOptions{
		Config: config,
		Logger: logger,
	})

	// Criar handler
	handler := newVaultHandler(lockariService)

	// Simular método do handler
	getVault := func(ctx context.Context, userID, vaultID string) error {
		// Verificar se o usuário pode acessar o vault
		canAccess, err := handler.authService.CanAccessVault(ctx, userID, vaultID, authorization.VaultPermissionRead)
		if err != nil {
			return fmt.Errorf("authorization check failed: %w", err)
		}

		if !canAccess {
			return fmt.Errorf("user %s cannot access vault %s", userID, vaultID)
		}

		// Continuar com a lógica do handler...
		fmt.Printf("✅ User %s successfully accessed vault %s\n", userID, vaultID)
		return nil
	}

	// Testar o handler
	ctx := context.Background()
	err := getVault(ctx, "user123", "vault456")
	if err != nil {
		fmt.Printf("❌ Handler error: %v\n", err)
	}

	fmt.Println("\n=== Exemplo de Handler concluído ===")
}

// ExamplePermissions demonstra as permissões disponíveis
func ExamplePermissions() {
	fmt.Println("\n=== Exemplo de Permissões ===\n")

	// Vault Permissions
	fmt.Println("Vault Permissions:")
	vaultPermissions := []authorization.VaultPermission{
		authorization.VaultPermissionView,
		authorization.VaultPermissionRead,
		authorization.VaultPermissionCopy,
		authorization.VaultPermissionDownload,
		authorization.VaultPermissionWrite,
		authorization.VaultPermissionDelete,
		authorization.VaultPermissionShare,
		authorization.VaultPermissionManage,
	}

	for _, permission := range vaultPermissions {
		fmt.Printf("  - %s (válida: %v)\n", permission, permission.IsValid())
	}

	// Tenant Roles
	fmt.Println("\nTenant Roles:")
	tenantRoles := []authorization.TenantRole{
		authorization.TenantRoleOwner,
		authorization.TenantRoleAdmin,
		authorization.TenantRoleMember,
	}

	for _, role := range tenantRoles {
		fmt.Printf("  - %s\n", role)
	}

	// Group Roles
	fmt.Println("\nGroup Roles:")
	groupRoles := []authorization.GroupRole{
		authorization.GroupRoleOwner,
		authorization.GroupRoleAdmin,
		authorization.GroupRoleMember,
	}

	for _, role := range groupRoles {
		fmt.Printf("  - %s\n", role)
	}

	fmt.Println("\n=== Permissões listadas ===")
}

func RunExamples() {
	// Executar exemplos
	ExampleLockariServiceUsage()
	ExampleInHandler()
	ExamplePermissions()
}

// Função para demonstrar como integrar com dependency injection
func ExampleWithDI() {
	fmt.Println("\n=== Exemplo com Dependency Injection ===\n")

	// Simulando um container de DI
	type ServiceContainer struct {
		AuthService authorization.LockariAuthorizationService
	}

	// Função para criar container
	createContainer := func() *ServiceContainer {
		config := &authorization.Config{
			APIURL:       "http://localhost:8080",
			StoreID:      "01HXSJ9QXKV8J9DK8V7S9QXKV8",
			Timeout:      30 * time.Second,
			CacheEnabled: true,
			AuditEnabled: true,
		}

		logger := authorization.NewSlogAdapter(slog.Default())

		authService := authorization.NewLockariService(authorization.LockariServiceOptions{
			Config: config,
			Logger: logger,
		})

		return &ServiceContainer{
			AuthService: authService,
		}
	}

	// Criar e usar container
	container := createContainer()

	// Usar o serviço através do container
	ctx := context.Background()
	err := container.AuthService.Health(ctx)
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
	} else {
		fmt.Printf("✅ Service is healthy via DI container\n")
	}

	fmt.Println("\n=== Exemplo com DI concluído ===")
}

// Função principal dos exemplos
func ExampleMain() {
	fmt.Println("🚀 Iniciando exemplos de uso da interface LockariAuthorizationService\n")

	// Executar todos os exemplos
	RunExamples()
	ExampleWithDI()

	fmt.Println("\n🎉 Todos os exemplos foram executados com sucesso!")
	fmt.Println("\nPara usar em seu projeto:")
	fmt.Println("1. Configure as dependências (config, logger)")
	fmt.Println("2. Crie o serviço com NewLockariService()")
	fmt.Println("3. Use a interface LockariAuthorizationService")
	fmt.Println("4. Injete nos seus handlers e middlewares")
	fmt.Println("5. Teste com mocks quando necessário")
}
