# Lockari Authorization Package

Este package implementa a interface de autorização para o Lockari usando OpenFGA (Fine-Grained Authorization).

## Visão Geral

O sistema de autorização do Lockari é baseado no modelo OpenFGA refinado que suporta:

- **Multi-tenancy**: Isolamento completo entre organizações
- **Permissões granulares**: Controle fino sobre cada operação
- **Compartilhamento externo**: Sharing seguro entre tenants (Enterprise)
- **Tokens de API**: Automação com permissões limitadas
- **Auditoria completa**: Log de todas as operações

## Arquitetura

### Interfaces

#### `AuthorizationService`
Interface básica para operações de autorização OpenFGA:
- `Check()`: Verificação de permissão única
- `CheckBatch()`: Verificação em lote
- `ListObjects()`: Listar objetos acessíveis
- `Write()`: Criar relações
- `Delete()`: Remover relações

#### `LockariAuthorizationService`
Interface específica do domínio Lockari:
- `CanAccessVault()`: Verificar acesso a vault
- `SetupTenant()`: Configurar novo tenant
- `CreateAPIToken()`: Criar tokens de automação
- `InitiateExternalSharing()`: Compartilhamento entre tenants

### Tipos de Permissões

#### Vault Permissions
```go
VaultPermissionView         // Ver metadados
VaultPermissionRead         // Ler conteúdo
VaultPermissionCopy         // Copiar para clipboard
VaultPermissionDownload     // Baixar/exportar
VaultPermissionWrite        // Criar/editar
VaultPermissionDelete       // Deletar
VaultPermissionShare        // Compartilhar
VaultPermissionManage       // Gerenciar vault
```

#### Secret Permissions
```go
SecretPermissionView           // Ver metadados
SecretPermissionRead           // Ler conteúdo
SecretPermissionCopy           // Copiar normal
SecretPermissionDownload       // Baixar/exportar
SecretPermissionWrite          // Criar/editar
SecretPermissionDelete         // Deletar
SecretPermissionReadSensitive  // Ler secrets sensíveis
SecretPermissionCopySensitive  // Copiar secrets sensíveis
SecretPermissionCopyProduction // Copiar secrets de produção
```

#### Token Permissions
```go
TokenPermissionUse                // Token ativo
TokenPermissionReadSecrets        // Ler via token
TokenPermissionWriteSecrets       // Escrever via token
TokenPermissionManageVault        // Gerenciar via token
TokenPermissionRevoke             // Revogar token
TokenPermissionRegenerate         // Regenerar token
```

### Papéis (Roles)

#### Tenant Roles
- `TenantRoleOwner`: Proprietário do tenant
- `TenantRoleAdmin`: Administrador
- `TenantRoleMember`: Membro regular
- `TenantRoleGuest`: Convidado externo

#### Group Roles
- `GroupRoleOwner`: Proprietário do grupo
- `GroupRoleAdmin`: Administrador do grupo
- `GroupRoleMember`: Membro do grupo

## Configuração

### Inicialização

```go
package main

import (
    "lockari-api-application/pkg/authorization"
)

func main() {
    // Configuração OpenFGA
    authService, err := authorization.NewAuthorizationService(
        "api.us1.fga.dev",              // API URL
        "01JZXBQMPXQB4XBDCVCJ7EMMM2",   // Store ID
        "auth.fga.dev",                 // API Token Issuer
        "https://api.us1.fga.dev/",     // API Audience
        "your-client-id",               // Client ID
        "your-client-secret",           // Client Secret
        "read write",                   // Scopes
        "your-model-id",                // Authorization Model ID
    )
    if err != nil {
        panic(err)
    }

    // Ou usar a versão específica do Lockari
    lockariService, err := authorization.NewLockariAuthorizationService(
        // ... mesmos parâmetros
    )
    if err != nil {
        panic(err)
    }
}
```

### Variáveis de Ambiente

```bash
# OpenFGA Configuration
OPENFGA_API_URL=api.us1.fga.dev
OPENFGA_STORE_ID=01JZXBQMPXQB4XBDCVCJ7EMMM2
OPENFGA_API_TOKEN_ISSUER=auth.fga.dev
OPENFGA_API_AUDIENCE=https://api.us1.fga.dev/
OPENFGA_CLIENT_ID=your-client-id
OPENFGA_CLIENT_SECRET=your-client-secret
OPENFGA_SCOPES=read write
OPENFGA_MODEL_ID=your-model-id
```

## Uso

### Verificação de Permissões

```go
ctx := context.Background()

// Verificar se usuário pode ler um vault
canRead, err := authService.CheckVaultPermission(ctx, "alice", "vault-123", authorization.VaultPermissionRead)
if err != nil {
    log.Printf("Error checking permission: %v", err)
    return
}

if canRead {
    fmt.Println("Alice pode ler o vault")
}

// Verificar múltiplas permissões
requests := []*authorization.CheckRequest{
    {User: "user:alice", Relation: "can_read", Object: "vault:secrets"},
    {User: "user:bob", Relation: "can_write", Object: "vault:secrets"},
}

responses, err := authService.CheckBatch(ctx, requests)
if err != nil {
    log.Printf("Error in batch check: %v", err)
    return
}

for i, resp := range responses {
    fmt.Printf("Request %d: %v\n", i, resp.Allowed)
}
```

### Configuração de Tenant

```go
ctx := context.Background()

// Configurar novo tenant
tenantID := "company-acme"
ownerID := "alice"

err := lockariService.SetupTenant(ctx, tenantID, ownerID, authorization.PlanFeatureAdvancedPermissions)
if err != nil {
    log.Printf("Error setting up tenant: %v", err)
    return
}

// Adicionar usuários ao tenant
err = lockariService.AddUserToTenant(ctx, "bob", tenantID, authorization.TenantRoleAdmin)
if err != nil {
    log.Printf("Error adding user to tenant: %v", err)
    return
}
```

### Gerenciamento de Grupos

```go
ctx := context.Background()

// Criar grupo
groupID := "developers"
err := lockariService.CreateGroup(ctx, groupID, "company-acme", "alice")
if err != nil {
    log.Printf("Error creating group: %v", err)
    return
}

// Adicionar usuários ao grupo
err = lockariService.AddUserToGroup(ctx, "bob", groupID, authorization.GroupRoleAdmin)
if err != nil {
    log.Printf("Error adding user to group: %v", err)
    return
}
```

### Tokens de API

```go
ctx := context.Background()

// Criar token com permissões específicas
permissions := []authorization.TokenPermission{
    authorization.TokenPermissionReadSecrets,
    authorization.TokenPermissionWriteSecrets,
}

tokenID, err := lockariService.CreateAPIToken(ctx, "alice", "vault-123", permissions)
if err != nil {
    log.Printf("Error creating API token: %v", err)
    return
}

fmt.Printf("Token criado: %s\n", tokenID)

// Verificar permissões do token
canUse, err := authService.CheckTokenPermission(ctx, tokenID, authorization.TokenPermissionUse)
if err != nil {
    log.Printf("Error checking token permission: %v", err)
    return
}

if canUse {
    fmt.Println("Token está ativo e pode ser usado")
}
```

### Compartilhamento Externo

```go
ctx := context.Background()

// Iniciar compartilhamento externo (Enterprise)
requestID, err := lockariService.InitiateExternalSharing(ctx, "alice", "vault-123", "company-beta")
if err != nil {
    log.Printf("Error initiating external sharing: %v", err)
    return
}

fmt.Printf("Solicitação de compartilhamento criada: %s\n", requestID)

// Aprovar compartilhamento
err = lockariService.ApproveExternalSharing(ctx, "alice", requestID)
if err != nil {
    log.Printf("Error approving external sharing: %v", err)
    return
}
```

### Listagem de Recursos

```go
ctx := context.Background()

// Listar vaults acessíveis
vaults, err := lockariService.ListAccessibleVaults(ctx, "alice")
if err != nil {
    log.Printf("Error listing accessible vaults: %v", err)
    return
}

fmt.Printf("Vaults acessíveis: %v\n", vaults)

// Listar objetos de um tipo específico
request := &authorization.ListObjectsRequest{
    User:     "user:alice",
    Relation: "can_read",
    Type:     "vault",
}

response, err := authService.ListObjects(ctx, request)
if err != nil {
    log.Printf("Error listing objects: %v", err)
    return
}

fmt.Printf("Objetos encontrados: %v\n", response.Objects)
```

## Modelo de Dados

### Estrutura das Relações

```
user:alice
├── owner → tenant:company-acme
├── admin → group:developers
├── writer → vault:api-secrets
└── can_read → secret:db-password

vault:api-secrets
├── tenant → tenant:company-acme
├── owner → user:alice
├── writer → group:developers
└── can_read → user:bob

tenant:company-acme
├── owner → user:alice
├── admin → user:bob
├── member → user:charlie
└── plan → plan:enterprise
```

### Hierarquia de Permissões

```
VAULT:
VIEW < READ < COPY < DOWNLOAD < WRITE < DELETE < SHARE < MANAGE

SECRET:
VIEW < READ < COPY < DOWNLOAD < WRITE < DELETE
     └── READ_SENSITIVE (owner only)
         └── COPY_SENSITIVE (owner only)
             └── COPY_PRODUCTION (admin/owner only)

TOKEN:
USE → READ_SECRETS / WRITE_SECRETS / MANAGE_VAULT
    └── REVOKE / REGENERATE (owner only)
```

## Testes

Execute os testes unitários:

```bash
go test ./pkg/authorization/...
```

Execute os benchmarks:

```bash
go test -bench=. ./pkg/authorization/...
```

## Exemplos Completos

Veja os arquivos de exemplo:

- `examples.go`: Exemplos de uso completos
- `authorization_test.go`: Testes unitários
- `README.md`: Esta documentação

## Recursos Avançados

### Cache de Permissões

```go
// Pré-carregar permissões comuns
err := authService.PreloadPermissions(ctx, "alice")
if err != nil {
    log.Printf("Error preloading permissions: %v", err)
}

// Invalidar cache após mudanças
err = authService.InvalidateCache(ctx, "alice")
if err != nil {
    log.Printf("Error invalidating cache: %v", err)
}
```

### Auditoria

Todas as operações são automaticamente auditadas e logadas. As informações incluem:

- Usuário que fez a operação
- Recurso acessado
- Tipo de operação
- Timestamp
- Resultado (permitido/negado)
- Contexto adicional

### Segurança

- **Princípio do menor privilégio**: Usuários só têm acesso ao mínimo necessário
- **Isolamento por tenant**: Dados completamente isolados entre organizações
- **Tokens com escopo limitado**: Automação com permissões específicas
- **Aprovação dupla**: Compartilhamento externo requer duas aprovações
- **Auditoria completa**: Log de todas as operações

## Integração com Gin

### Middleware de Autorização

```go
import (
    "github.com/gin-gonic/gin"
    "lockari-api-application/pkg/authorization"
)

func main() {
    router := gin.Default()
    
    // Inicializar o serviço de autorização
    authService, err := authorization.NewLockariAuthorizationService(/* params */)
    if err != nil {
        panic(err)
    }
    
    // Middleware de autorização
    router.Use(authorization.AuthorizationMiddleware(authService))
    
    // Rotas protegidas
    router.GET("/vaults/:id", 
        authorization.RequireVaultPermission(authorization.VaultPermissionRead),
        handleGetVault,
    )
    
    router.POST("/vaults", 
        authorization.RequireTenantPermission(authorization.TenantPermissionCreateVault),
        handleCreateVault,
    )
    
    router.PUT("/vaults/:id/secrets/:secret_id", 
        authorization.RequireSecretPermission(authorization.SecretPermissionWrite),
        handleUpdateSecret,
    )
    
    router.Run(":8080")
}

func handleGetVault(c *gin.Context) {
    // A autorização já foi verificada pelo middleware
    vaultID := c.Param("id")
    // ... implementação
}
```

### Batch Authorization

```go
// Middleware para verificar múltiplas permissões
router.GET("/dashboard", 
    authorization.RequireBatchPermissions(map[string]interface{}{
        "vault:*": authorization.VaultPermissionRead,
        "secret:*": authorization.SecretPermissionView,
    }),
    handleDashboard,
)
```

### Custom Permissions

```go
// Verificação customizada
router.POST("/vaults/:id/share", 
    authorization.RequireCustomPermission(func(c *gin.Context, authService authorization.LockariAuthorizationService) bool {
        vaultID := c.Param("id")
        userID := c.GetString("user_id")
        
        // Verificar se usuário pode compartilhar E se vault permite compartilhamento
        canShare, _ := authService.CheckVaultPermission(c, userID, vaultID, authorization.VaultPermissionShare)
        isShareable, _ := authService.IsVaultShareable(c, vaultID)
        
        return canShare && isShareable
    }),
    handleShareVault,
)
```

## Performance e Otimização

### Cache Strategy

```go
// Configurar cache para alta performance
config := &authorization.CacheConfig{
    MaxSize:    10000,
    TTL:        time.Minute * 15,
    WarmUpSize: 1000,
}

authService.SetCacheConfig(config)

// Pré-carregar permissões críticas
authService.WarmUpCache(ctx, []string{"alice", "bob", "charlie"})
```

### Batch Operations

```go
// Otimizar verificações múltiplas
permissions := []authorization.PermissionCheck{
    {User: "alice", Object: "vault:secrets", Permission: authorization.VaultPermissionRead},
    {User: "alice", Object: "vault:configs", Permission: authorization.VaultPermissionWrite},
    {User: "bob", Object: "vault:secrets", Permission: authorization.VaultPermissionRead},
}

results, err := authService.CheckPermissionsBatch(ctx, permissions)
if err != nil {
    log.Printf("Error in batch check: %v", err)
    return
}

for i, result := range results {
    fmt.Printf("Permission %d: %v\n", i, result.Allowed)
}
```

### Monitoring e Métricas

```go
// Métricas de performance
metrics := authService.GetMetrics()
fmt.Printf("Cache hit rate: %.2f%%\n", metrics.CacheHitRate)
fmt.Printf("Average response time: %v\n", metrics.AverageResponseTime)
fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
```

## Troubleshooting

### Erros Comuns

1. **"FGA client not initialized"**
   - Verifique se as credenciais OpenFGA estão corretas
   - Confirme se o modelo de autorização foi carregado
   - Teste conectividade: `ping api.us1.fga.dev`

2. **"Permission denied"**
   - Verifique se o usuário tem a permissão necessária
   - Confirme se o usuário é membro ativo do tenant
   - Checque hierarquia de permissões
   - Valide formatação de user/object IDs

3. **"Token expired"**
   - Regenere o token de API
   - Verifique se o token não foi revogado
   - Confirme configuração de TTL

4. **"Store not found"**
   - Verifique se o Store ID está correto
   - Confirme se o store existe no OpenFGA
   - Teste com comando curl direto

5. **"Model mismatch"**
   - Verifique se o Model ID está correto
   - Confirme se o modelo foi corretamente carregado
   - Valide estrutura das relações

### Debug

Enable debug logging:

```go
import "log"

// Habilitar logs detalhados
log.SetLevel(log.DebugLevel)

// Ou configurar logger customizado
logger := authorization.NewLogger(authorization.LogLevelDebug)
authService.SetLogger(logger)
```

### Diagnostic Tools

```go
// Verificar configuração
err := authService.HealthCheck(ctx)
if err != nil {
    log.Printf("Health check failed: %v", err)
}

// Dump de relações para debug
relations, err := authService.DumpRelations(ctx, "user:alice")
if err != nil {
    log.Printf("Error dumping relations: %v", err)
    return
}

for _, rel := range relations {
    fmt.Printf("%s -> %s -> %s\n", rel.User, rel.Relation, rel.Object)
}

// Trace de verificação
trace, err := authService.TraceCheck(ctx, "user:alice", "can_read", "vault:secrets")
if err != nil {
    log.Printf("Error tracing check: %v", err)
    return
}

fmt.Printf("Trace steps: %v\n", trace.Steps)
```

### Edge Cases

1. **Multi-tenant isolation**
   - Sempre prefixe recursos com tenant ID
   - Valide isolamento nos testes
   - Monitore vazamentos entre tenants

2. **Token rotation**
   - Implemente rotação automática
   - Mantenha tokens de fallback
   - Monitore expiração

3. **Large datasets**
   - Use paginação em ListObjects
   - Implemente timeouts apropriados
   - Configure limites de rate

4. **Network issues**
   - Implemente retry policy
   - Configure circuit breaker
   - Monitore latência

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Adicione testes para novas funcionalidades
4. Execute todos os testes
5. Submeta um pull request

## Licença

Este projeto está licenciado sob a MIT License.
