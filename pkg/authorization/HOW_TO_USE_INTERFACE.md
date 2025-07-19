# Como Acessar e Usar a Interface LockariAuthorizationService

## Visão Geral

A interface `LockariAuthorizationService` é a principal interface para operações de autorização no Lockari. Ela estende a interface básica `AuthorizationService` com métodos específicos do domínio.

## Criando uma Instância

### 1. Configuração Básica

```go
package main

import (
    "context"
    "log/slog"
    "time"
    
    "lockari-api-application/pkg/authorization"
)

func main() {
    // 1. Criar configuração
    config := &authorization.Config{
        APIURL:   "http://localhost:8080",
        StoreID:  "01HXSJ9QXKV8J9DK8V7S9QXKV8",
        Timeout:  30 * time.Second,
        CacheEnabled: true,
        CacheTTL:     5 * time.Minute,
        AuditEnabled: true,
    }

    // 2. Criar logger
    logger := slog.Default()

    // 3. Criar cliente OpenFGA
    client, err := authorization.NewClient(config, logger)
    if err != nil {
        panic(err)
    }

    // 4. Criar serviços auxiliares
    auditService := authorization.NewAuditService(logger)
    cacheService := authorization.NewCacheService(config, logger)

    // 5. Criar serviço básico
    service := authorization.NewService(authorization.ServiceOptions{
        Client:       client,
        Cache:        cacheService,
        Audit:        auditService,
        Logger:       logger,
    })

    // 6. Criar serviço Lockari (implementação da interface)
    lockariService := authorization.NewLockariService(authorization.LockariServiceOptions{
        Service:      service,
        AuditService: auditService,
        Config:       config,
        Logger:       logger,
    })

    // 7. Agora você tem acesso à interface LockariAuthorizationService
    useService(lockariService)
}
```

### 2. Usando a Interface

```go
func useService(authService authorization.LockariAuthorizationService) {
    ctx := context.Background()
    
    // Exemplo 1: Verificar acesso a vault
    userID := "user123"
    vaultID := "vault456"
    
    canRead, err := authService.CanAccessVault(ctx, userID, vaultID, authorization.VaultPermissionRead)
    if err != nil {
        log.Printf("Erro: %v", err)
        return
    }
    
    if canRead {
        log.Printf("Usuário %s pode ler vault %s", userID, vaultID)
    } else {
        log.Printf("Usuário %s NÃO pode ler vault %s", userID, vaultID)
    }
    
    // Exemplo 2: Configurar novo vault
    tenantID := "tenant789"
    ownerID := "owner123"
    
    err = authService.SetupVault(ctx, vaultID, tenantID, ownerID)
    if err != nil {
        log.Printf("Erro ao configurar vault: %v", err)
        return
    }
    
    log.Printf("Vault %s configurado com sucesso", vaultID)
    
    // Exemplo 3: Compartilhar vault
    targetUserID := "user456"
    
    err = authService.ShareVault(ctx, vaultID, ownerID, targetUserID, authorization.VaultPermissionRead)
    if err != nil {
        log.Printf("Erro ao compartilhar vault: %v", err)
        return
    }
    
    log.Printf("Vault %s compartilhado com usuário %s", vaultID, targetUserID)
}
```

## Integração com Handlers

### 1. Injeção de Dependência

```go
type VaultHandler struct {
    authService authorization.LockariAuthorizationService
}

func NewVaultHandler(authService authorization.LockariAuthorizationService) *VaultHandler {
    return &VaultHandler{
        authService: authService,
    }
}

func (h *VaultHandler) GetVault(c *gin.Context) {
    userID := c.GetString("user_id")
    vaultID := c.Param("id")
    
    // Verificar permissão
    canAccess, err := h.authService.CanAccessVault(c.Request.Context(), userID, vaultID, authorization.VaultPermissionRead)
    if err != nil {
        c.JSON(500, gin.H{"error": "authorization check failed"})
        return
    }
    
    if !canAccess {
        c.JSON(403, gin.H{"error": "forbidden"})
        return
    }
    
    // Continuar com a lógica do handler...
    c.JSON(200, gin.H{"message": "vault accessed successfully"})
}
```

### 2. Middleware de Autorização

```go
func (h *VaultHandler) RequireVaultPermission(permission authorization.VaultPermission) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        vaultID := c.Param("id")
        
        canAccess, err := h.authService.CanAccessVault(c.Request.Context(), userID, vaultID, permission)
        if err != nil {
            c.JSON(500, gin.H{"error": "authorization check failed"})
            c.Abort()
            return
        }
        
        if !canAccess {
            c.JSON(403, gin.H{"error": "forbidden"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Uso no router
func setupRoutes(router *gin.Engine, handler *VaultHandler) {
    vaultRoutes := router.Group("/api/v1/vaults")
    {
        vaultRoutes.GET("/:id", 
            handler.RequireVaultPermission(authorization.VaultPermissionRead),
            handler.GetVault,
        )
        
        vaultRoutes.PUT("/:id", 
            handler.RequireVaultPermission(authorization.VaultPermissionWrite),
            handler.UpdateVault,
        )
        
        vaultRoutes.DELETE("/:id", 
            handler.RequireVaultPermission(authorization.VaultPermissionDelete),
            handler.DeleteVault,
        )
    }
}
```

## Testando com Mocks

### 1. Criando Mock da Interface

```go
type MockLockariAuthorizationService struct {
    mock.Mock
}

func (m *MockLockariAuthorizationService) CanAccessVault(ctx context.Context, userID, vaultID string, permission authorization.VaultPermission) (bool, error) {
    args := m.Called(ctx, userID, vaultID, permission)
    return args.Bool(0), args.Error(1)
}

func (m *MockLockariAuthorizationService) SetupVault(ctx context.Context, vaultID, tenantID, ownerID string) error {
    args := m.Called(ctx, vaultID, tenantID, ownerID)
    return args.Error(0)
}

// Implementar outros métodos da interface...
```

### 2. Usando Mock em Testes

```go
func TestVaultHandler_GetVault(t *testing.T) {
    // Arrange
    mockAuth := new(MockLockariAuthorizationService)
    handler := NewVaultHandler(mockAuth)
    
    // Configurar expectativas
    mockAuth.On("CanAccessVault", mock.Anything, "user123", "vault456", authorization.VaultPermissionRead).
        Return(true, nil)
    
    // Act
    req, _ := http.NewRequest("GET", "/api/v1/vaults/vault456", nil)
    w := httptest.NewRecorder()
    
    // ... resto do teste
    
    // Assert
    mockAuth.AssertExpectations(t)
}
```

## Principais Métodos da Interface

### Operações de Vault
- `CanAccessVault()` - Verificar permissão de acesso
- `SetupVault()` - Configurar novo vault
- `ShareVault()` - Compartilhar vault com usuário
- `RevokeVaultAccess()` - Revogar acesso ao vault
- `ListAccessibleVaults()` - Listar vaults acessíveis

### Operações de Tenant
- `SetupTenant()` - Configurar novo tenant
- `AddUserToTenant()` - Adicionar usuário ao tenant
- `RemoveUserFromTenant()` - Remover usuário do tenant

### Operações de Grupo
- `CreateGroup()` - Criar novo grupo
- `AddUserToGroup()` - Adicionar usuário ao grupo

### Operações de Token
- `CreateAPIToken()` - Criar token de API
- `CheckTokenPermission()` - Verificar permissão do token
- `RevokeToken()` - Revogar token

### Operações de Auditoria
- `GetAuditLogs()` - Obter logs de auditoria

## Permissões Disponíveis

### VaultPermission
- `VaultPermissionView` - Ver metadados
- `VaultPermissionRead` - Ler conteúdo
- `VaultPermissionCopy` - Copiar para clipboard
- `VaultPermissionDownload` - Baixar/exportar
- `VaultPermissionWrite` - Criar/editar
- `VaultPermissionDelete` - Deletar
- `VaultPermissionShare` - Compartilhar
- `VaultPermissionManage` - Gerenciar completamente

### TenantRole
- `TenantRoleOwner` - Proprietário do tenant
- `TenantRoleAdmin` - Administrador
- `TenantRoleMember` - Membro regular

### GroupRole
- `GroupRoleOwner` - Proprietário do grupo
- `GroupRoleAdmin` - Administrador do grupo
- `GroupRoleMember` - Membro do grupo

### TokenPermission
- `TokenPermissionRead` - Permissão de leitura
- `TokenPermissionWrite` - Permissão de escrita
- `TokenPermissionDelete` - Permissão de deletar
- `TokenPermissionShare` - Permissão de compartilhar

## Configuração Avançada

### 1. Com Cache Redis

```go
// Configurar cache Redis
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

cacheService := authorization.NewRedisCacheService(redisClient, config)
```

### 2. Com Audit Personalizado

```go
// Criar audit service personalizado
auditService := authorization.NewCustomAuditService(
    authorization.AuditServiceOptions{
        Storage:    customStorage,
        Formatter:  customFormatter,
        AsyncMode:  true,
        BatchSize:  100,
    },
)
```

### 3. Com Métricas Prometheus

```go
// Configurar métricas
metricsService := authorization.NewPrometheusMetrics()

service := authorization.NewService(authorization.ServiceOptions{
    Client:  client,
    Cache:   cacheService,
    Audit:   auditService,
    Metrics: metricsService,
    Logger:  logger,
})
```

### 4. Exemplo Completo de Uso

Veja o arquivo `examples/lockari_service_usage.go` para um exemplo completo e funcional.

### 5. Configuração de Usuários/Tenants por Plano

Para configurar permissões de novos usuários de forma automática baseada no plano:

```go
// Exemplo de criação de usuário/tenant
func CreateUser(ctx context.Context, authService authorization.LockariAuthorizationService) {
    userID := "user123"
    tenantID := "tenant456"
    plan := authorization.PlanPro // ou PlanFree, PlanEnterprise
    
    err := authService.CreateNewUserTenant(ctx, userID, tenantID, plan)
    if err != nil {
        log.Printf("Erro ao criar usuário/tenant: %v", err)
        return
    }
    
    log.Printf("Usuário/tenant criado com sucesso para plano %s", plan)
}
```

**Diferenças por plano:**
- **Free**: 1 vault pessoal, compartilhamento básico
- **Pro**: 2 vaults + grupo de equipe, API, auditoria
- **Enterprise**: 4 vaults departamentais + 4 grupos, SSO, compartilhamento externo

Veja o guia completo em `pkg/authorization/SETUP_PERMISSIONS_GUIDE.md`

## Conclusão

A interface `LockariAuthorizationService` fornece uma API completa e type-safe para todas as operações de autorização no Lockari. Ela é facilmente testável, extensível e integra perfeitamente com o ecossistema Gin/HTTP do projeto.

Para usar:
1. Configure as dependências (Client, Cache, Audit, Logger)
2. Crie o serviço básico
3. Crie o serviço Lockari
4. Use a interface nos seus handlers e middlewares
5. Teste com mocks quando necessário
