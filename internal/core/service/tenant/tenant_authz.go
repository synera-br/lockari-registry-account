package tenant

import (
	"context"
	"fmt"

	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/tenant"
	"github.com/synera-br/lockari-backend-app/pkg/authorization"
)

func (s *tenantEventService) initializeAuthorizer(ctx context.Context, tenant *entity.Tenant, defaultUser entity.Owner, defaultGroup *entity.UserGroup, defaultVault *entity.Vault) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Debug detalhado da configuração antes do health check
	fmt.Println("\n=== DEBUGGING AUTHORIZATION SERVICE ===")
	s.debugAuthorizerConfig()

	// Verificar se o authorizer foi inicializado
	if s.authorizer == nil {
		return fmt.Errorf("authorizer is nil - service not properly initialized")
	}

	// Tentar health check com mais detalhes
	fmt.Println("Attempting health check...")
	health, err := s.authorizer.Health(ctx)
	if err != nil {
		fmt.Printf("Health check failed with error: %v\n", err)
		fmt.Printf("Error type: %T\n", err)

		// Implementar health check simplificado como fallback
		fmt.Println("Attempting simplified connectivity test...")
		if simpleErr := s.testBasicConnectivity(ctx); simpleErr != nil {
			fmt.Printf("Basic connectivity test also failed: %v\n", simpleErr)

			// Fornecer dicas específicas baseadas no erro
			if fmt.Sprintf("%v", err) == "403 Forbidden" {
				fmt.Println("DICAS PARA RESOLVER 403 FORBIDDEN:")
				fmt.Println("1. Verifique se a URL contém https:// -> api_url: 'https://api.us1.fga.dev'")
				fmt.Println("2. Verifique client_id e client_secret")
				fmt.Println("3. Verifique se o scope está correto -> scopes: 'fga:api'")
				fmt.Println("4. Verifique se o store_id está correto")
			}

			return fmt.Errorf("authorization service completely unreachable: original error: %w", err)
		}

		fmt.Println("Basic connectivity OK, but health check failed")
		// Continuar mesmo com health check falhando para debug
	} else {
		fmt.Printf("Health check successful: %+v\n", health)
	}

	fmt.Println("=== END DEBUGGING ===\n")

	err = s.createTenantInAuthorization(ctx, tenant)
	if err != nil {
		// Rollback tenant creation in database if authorization fails
		originalErr := err
		if rollbackErr := s.authenticator.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.ID); rollbackErr != nil {
			return fmt.Errorf("failed to rollback tenant creation in database: %w", rollbackErr)
		}
		return fmt.Errorf("failed to create tenant in authorization service: %w", originalErr)
	}

	relation := fmt.Sprintf("%s", entity.TenantGroupOwner)
	err = s.AssociateGroupToTenant(ctx, &defaultGroup.ID, &tenant.ID, &defaultUser.Uid, &relation)
	if err != nil {
		return fmt.Errorf("failed to associate group to tenant: %w", err)
	}

	err = s.AssociateUserToVault(ctx, &defaultVault.ID, &tenant.ID, &defaultUser.Uid)
	if err != nil {
		return fmt.Errorf("failed to associate user to vault: %w", err)
	}

	features := make([]authorization.PlanFeature, 0)
	realations := []string{"owner", fmt.Sprintf("%s", entity.TenantGroupOwner)}

	err = s.authorizer.SetupTenant(ctx, tenant.ID, tenant.Owner.GetEmail(), features, realations)
	if err != nil {
		if err := s.authenticator.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.ID); err != nil {
			return fmt.Errorf("failed to set tenant rollback: %w", err)
		}
		return authorization.NewAuthorizationError("SetupTenant", "Failed to setup tenant in authorization service", err)
	}
	return nil
}

func (s *tenantEventService) createTenantInAuthorization(ctx context.Context, tenant *entity.Tenant) error {

	features := make([]authorization.PlanFeature, 0)

	relations := []string{"owner", fmt.Sprintf("%s", entity.TenantGroupOwner)}

	err := s.authorizer.SetupTenant(ctx, tenant.ID, tenant.Owner.GetEmail(), features, relations)
	if err != nil {
		if err := s.authenticator.SetTenantRollback(ctx, tenant.Owner.GetEmail(), tenant.ID); err != nil {
			return fmt.Errorf("failed to set tenant rollback: %w", err)
		}
		return authorization.NewAuthorizationError("SetupTenant", "Failed to setup tenant in authorization service", err)
	}

	return nil
}

func (s *tenantEventService) AssociateGroupToTenant(ctx context.Context, groupID, tenantID, ownerID, relation *string) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if groupID == nil || tenantID == nil || ownerID == nil || relation == nil {
		return fmt.Errorf("all parameters are required")
	}

	if *relation == "" {
		return fmt.Errorf("relation is required")
	}

	if *groupID == "" {
		return fmt.Errorf("groupID is required")
	}

	if *tenantID == "" {
		return fmt.Errorf("tenantID is required")
	}

	if *ownerID == "" {
		return fmt.Errorf("ownerID is required")
	}

	err := s.authorizer.AssociateGroupToTenant(ctx, *groupID, *tenantID, *ownerID, *relation)
	if err != nil {
		return fmt.Errorf("failed to associate group to tenant: %w", err)
	}

	err = s.authorizer.AssociateUserToGroupAsMember(ctx, *ownerID, *groupID, *tenantID)
	if err != nil {
		return fmt.Errorf("failed to associate user to group as member: %w", err)
	}

	return nil
}

func (s *tenantEventService) AssociateUserToVault(ctx context.Context, vaultID, tenantID, userID *string) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if vaultID == nil || tenantID == nil || userID == nil {
		return fmt.Errorf("all parameters are required")
	}

	if *vaultID == "" {
		return fmt.Errorf("vaultID is required")
	}

	if *tenantID == "" {
		return fmt.Errorf("tenantID is required")
	}

	if *userID == "" {
		return fmt.Errorf("userID is required")
	}

	err := s.authorizer.SetupVault(ctx, *vaultID, *tenantID, *userID)
	if err != nil {
		return fmt.Errorf("failed to setup vault: %w", err)
	}

	return nil
}

// testBasicConnectivity testa conectividade básica com o OpenFGA
func (s *tenantEventService) testBasicConnectivity(ctx context.Context) error {
	if s.authorizer == nil {
		return fmt.Errorf("authorizer is nil")
	}

	fmt.Println("Testing basic connectivity with OpenFGA...")

	// Tentar uma operação simples que não requer tuplas existentes
	// Por exemplo, listar objetos vazios para um usuário fake
	result, err := s.authorizer.ListPermissionFromTenant(ctx, "health-check-tenant-id")

	fmt.Printf("ListPermissionFromTenant result: %v\n", result)
	fmt.Printf("ListPermissionFromTenant error: %v\n", err)

	// Se retornar erro 403, é problema de autenticação
	if err != nil {
		errStr := fmt.Sprintf("%v", err)
		fmt.Printf("Error string contains: %s\n", errStr)

		if fmt.Sprintf("%v", err) == "403 Forbidden" ||
			fmt.Sprintf("%v", err) == "403" ||
			fmt.Sprintf("%v", err) == "Forbidden" {
			return fmt.Errorf("authentication failed - 403 Forbidden detected")
		}

		// Outros erros podem ser OK para conectividade básica
		fmt.Printf("Got error but might be OK for connectivity test: %v\n", err)
	}

	fmt.Println("Basic connectivity test completed")
	return nil // Conectividade básica OK
}

// debugAuthorizerConfig imprime informações de debug sobre a configuração
func (s *tenantEventService) debugAuthorizerConfig() {
	fmt.Println("=== AUTHORIZER DEBUG INFO ===")
	fmt.Printf("Authorizer initialized: %t\n", s.authorizer != nil)

	if s.authorizer != nil {
		// Se possível, obter informações de configuração
		fmt.Println("Authorizer type: LockariAuthorizationService")

		// Tentar casting para obter mais detalhes (se necessário)
		// Isso pode variar dependendo da sua implementação
	}
	fmt.Println("=== END DEBUG INFO ===")
}
