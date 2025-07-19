# OpenFGA Integration Documentation

## Arquitetura de Autenticação/Autorização

### Visão Geral

O OpenFGA **não precisa** integrar diretamente com o Firebase Authentication. A arquitetura funciona da seguinte forma:

```
Frontend (Next.js) 
    ↓ [Firebase JWT Token]
Backend (Go) 
    ↓ [User ID + Tenant ID + Resource]
OpenFGA Server
    ↓ [Check Permissions]
Backend (Go)
    ↓ [Allow/Deny Response]
Frontend (Next.js)
```

### Fluxo de Autenticação/Autorização

1. **Frontend**: Usuário faz login via Firebase Auth
2. **Frontend**: Envia requisições com JWT token do Firebase
3. **Backend**: Valida JWT token e extrai `userID` e `tenantID`
4. **Backend**: Consulta OpenFGA para verificar permissões específicas
5. **Backend**: Retorna resposta baseada nas permissões

### Por que essa separação?

- **Firebase Auth**: Especializado em autenticação (identidade)
- **OpenFGA**: Especializado em autorização (permissões)
- **Backend**: Orquestra a comunicação entre ambos

## Configuração Docker Compose

### docker-compose.yml

```yaml
version: '3.8'

services:
  # Backend Lockari
  lockari-backend:
    build:
      context: .
      dockerfile: build/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - CONFIG_PATH=/app/config/config.yaml
      - OPENFGA_API_URL=http://openfga:8080
      - POSTGRES_URL=postgres://openfga:openfga_password@postgres:5432/openfga?sslmode=disable
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis
      - openfga
    volumes:
      - ./cmd/config:/app/config
    networks:
      - lockari-network

  # OpenFGA Server
  openfga:
    image: openfga/openfga:latest
    ports:
      - "8081:8080"
      - "8082:8081"  # gRPC port
      - "3000:3000"  # Playground
    environment:
      - OPENFGA_DATASTORE_ENGINE=postgres
      - OPENFGA_DATASTORE_URI=postgres://openfga:openfga_password@postgres:5432/openfga?sslmode=disable
      - OPENFGA_LOG_LEVEL=info
      - OPENFGA_HTTP_ADDR=0.0.0.0:8080
      - OPENFGA_GRPC_ADDR=0.0.0.0:8081
      - OPENFGA_PROFILER_ENABLED=false
      - OPENFGA_METRICS_ENABLED=true
      - OPENFGA_METRICS_ADDR=0.0.0.0:2112
      - OPENFGA_PLAYGROUND_ENABLED=true
      - OPENFGA_PLAYGROUND_PORT=3000
    depends_on:
      - postgres
    networks:
      - lockari-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3

  # PostgreSQL para OpenFGA
  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=openfga
      - POSTGRES_PASSWORD=openfga_password
      - POSTGRES_DB=openfga
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./docker/openfga/postgres-init.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - lockari-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U openfga"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Redis para cache
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - lockari-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  # OpenFGA CLI para setup inicial
  openfga-cli:
    image: openfga/cli:latest
    depends_on:
      - openfga
    volumes:
      - ./docker/openfga:/workspace
    networks:
      - lockari-network
    profiles:
      - setup
    command: >
      sh -c "
        echo 'Waiting for OpenFGA to be ready...' &&
        sleep 10 &&
        cd /workspace &&
        fga store create --name lockari-store --api-url http://openfga:8080 &&
        export STORE_ID=$$(fga store list --api-url http://openfga:8080 --format json | jq -r '.stores[0].id') &&
        echo 'Store ID:' $$STORE_ID &&
        fga model write --api-url http://openfga:8080 --store-id $$STORE_ID --file model.json &&
        fga tuple write --api-url http://openfga:8080 --store-id $$STORE_ID --file init-data.json &&
        echo 'OpenFGA setup completed!'
      "

volumes:
  postgres_data:
  redis_data:

networks:
  lockari-network:
    driver: bridge
```

## Arquivos de Configuração

### 1. PostgreSQL Initialization

**Arquivo**: `docker/openfga/postgres-init.sql`

```sql
-- Criar extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Criar índices adicionais para performance
CREATE INDEX IF NOT EXISTS idx_authorization_models_store_id ON authorization_models(store_id);
CREATE INDEX IF NOT EXISTS idx_tuples_store_id ON tuples(store_id);
CREATE INDEX IF NOT EXISTS idx_tuples_object_type ON tuples(object_type);
CREATE INDEX IF NOT EXISTS idx_tuples_user_object ON tuples(user_object);

-- Configurações de performance
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

-- Recarregar configurações
SELECT pg_reload_conf();
```

### 2. OpenFGA Authorization Model

**Arquivo**: `docker/openfga/model.json`

```json
{
  "schema_version": "1.1",
  "type_definitions": [
    {
      "type": "tenant",
      "relations": {
        "owner": {
          "this": {}
        },
        "admin": {
          "this": {}
        },
        "member": {
          "this": {}
        }
      },
      "metadata": {
        "relations": {
          "owner": {
            "directly_related_user_types": [
              {
                "type": "user"
              }
            ]
          },
          "admin": {
            "directly_related_user_types": [
              {
                "type": "user"
              }
            ]
          },
          "member": {
            "directly_related_user_types": [
              {
                "type": "user"
              }
            ]
          }
        }
      }
    },
    {
      "type": "vault",
      "relations": {
        "owner": {
          "this": {}
        },
        "admin": {
          "this": {}
        },
        "writer": {
          "this": {}
        },
        "viewer": {
          "this": {}
        },
        "tenant": {
          "this": {}
        },
        "read": {
          "union": {
            "child": [
              {
                "this": {}
              },
              {
                "computedUserset": {
                  "relation": "viewer"
                }
              },
              {
                "computedUserset": {
                  "relation": "writer"
                }
              },
              {
                "computedUserset": {
                  "relation": "admin"
                }
              },
              {
                "computedUserset": {
                  "relation": "owner"
                }
              }
            ]
          }
        },
        "write": {
          "union": {
            "child": [
              {
                "computedUserset": {
                  "relation": "writer"
                }
              },
              {
                "computedUserset": {
                  "relation": "admin"
                }
              },
              {
                "computedUserset": {
                  "relation": "owner"
                }
              }
            ]
          }
        },
        "delete": {
          "union": {
            "child": [
              {
                "computedUserset": {
                  "relation": "admin"
                }
              },
              {
                "computedUserset": {
                  "relation": "owner"
                }
              }
            ]
          }
        },
        "share": {
          "union": {
            "child": [
              {
                "computedUserset": {
                  "relation": "admin"
                }
              },
              {
                "computedUserset": {
                  "relation": "owner"
                }
              }
            ]
          }
        }
      },
      "metadata": {
        "relations": {
          "owner": {
            "directly_related_user_types": [
              {
                "type": "user"
              }
            ]
          },
          "admin": {
            "directly_related_user_types": [
              {
                "type": "user"
              }
            ]
          },
          "writer": {
            "directly_related_user_types": [
              {
                "type": "user"
              }
            ]
          },
          "viewer": {
            "directly_related_user_types": [
              {
                "type": "user"
              }
            ]
          },
          "tenant": {
            "directly_related_user_types": [
              {
                "type": "tenant"
              }
            ]
          }
        }
      }
    },
    {
      "type": "secret",
      "relations": {
        "vault": {
          "this": {}
        },
        "read": {
          "tupleToUserset": {
            "tupleset": {
              "relation": "vault"
            },
            "computedUserset": {
              "relation": "read"
            }
          }
        },
        "write": {
          "tupleToUserset": {
            "tupleset": {
              "relation": "vault"
            },
            "computedUserset": {
              "relation": "write"
            }
          }
        },
        "delete": {
          "tupleToUserset": {
            "tupleset": {
              "relation": "vault"
            },
            "computedUserset": {
              "relation": "delete"
            }
          }
        }
      },
      "metadata": {
        "relations": {
          "vault": {
            "directly_related_user_types": [
              {
                "type": "vault"
              }
            ]
          }
        }
      }
    },
    {
      "type": "user"
    }
  ]
}
```

### 3. Initial Data Setup

**Arquivo**: `docker/openfga/init-data.json`

```json
{
  "tuples": [
    {
      "user": "user:demo-user-1",
      "relation": "owner",
      "object": "tenant:demo-tenant-1"
    },
    {
      "user": "user:demo-user-1",
      "relation": "owner",
      "object": "vault:demo-vault-1"
    },
    {
      "user": "tenant:demo-tenant-1",
      "relation": "tenant",
      "object": "vault:demo-vault-1"
    },
    {
      "user": "vault:demo-vault-1",
      "relation": "vault",
      "object": "secret:demo-secret-1"
    }
  ]
}
```

## Configuração do Environment

### .env.dev

```bash
# OpenFGA Configuration
OPENFGA_API_URL=http://localhost:8081
OPENFGA_STORE_ID=
OPENFGA_AUTH_MODEL_ID=

# Database Configuration
POSTGRES_URL=postgres://openfga:openfga_password@localhost:5432/openfga?sslmode=disable

# Redis Configuration
REDIS_URL=redis://localhost:6379

# Firebase Configuration
FIREBASE_PROJECT_ID=your-firebase-project
FIREBASE_PRIVATE_KEY_PATH=./config/firebase-service-account.json

# Application Configuration
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=debug
```

## Setup Scripts

### scripts/setup-openfga.sh

```bash
#!/bin/bash

set -e

echo "🚀 Setting up OpenFGA for Lockari..."

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para log
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Verificar se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    error "Docker não está rodando. Por favor, inicie o Docker."
    exit 1
fi

# Verificar se docker-compose está disponível
if ! command -v docker-compose &> /dev/null; then
    error "docker-compose não está instalado."
    exit 1
fi

log "Parando containers existentes..."
docker-compose down -v

log "Construindo e iniciando containers..."
docker-compose up -d postgres redis

log "Aguardando PostgreSQL ficar pronto..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U openfga > /dev/null 2>&1; then
        break
    fi
    if [ $i -eq 30 ]; then
        error "PostgreSQL não ficou pronto em 30 segundos"
        exit 1
    fi
    sleep 1
done

log "Aguardando Redis ficar pronto..."
for i in {1..30}; do
    if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
        break
    fi
    if [ $i -eq 30 ]; then
        error "Redis não ficou pronto em 30 segundos"
        exit 1
    fi
    sleep 1
done

log "Iniciando OpenFGA..."
docker-compose up -d openfga

log "Aguardando OpenFGA ficar pronto..."
for i in {1..60}; do
    if curl -sf http://localhost:8081/healthz > /dev/null 2>&1; then
        break
    fi
    if [ $i -eq 60 ]; then
        error "OpenFGA não ficou pronto em 60 segundos"
        exit 1
    fi
    sleep 1
done

log "Verificando se CLI do OpenFGA está disponível..."
if ! command -v fga &> /dev/null; then
    warn "CLI do OpenFGA não encontrado. Usando container..."
    FGA_CMD="docker-compose run --rm openfga-cli fga"
else
    FGA_CMD="fga"
fi

log "Criando store no OpenFGA..."
STORE_OUTPUT=$(docker-compose run --rm openfga-cli fga store create --name lockari-store --api-url http://openfga:8080 --format json)
STORE_ID=$(echo $STORE_OUTPUT | jq -r '.id')

if [ "$STORE_ID" == "null" ] || [ -z "$STORE_ID" ]; then
    error "Falha ao criar store no OpenFGA"
    echo "Output: $STORE_OUTPUT"
    exit 1
fi

log "Store criado com ID: $STORE_ID"

log "Carregando modelo de autorização..."
MODEL_OUTPUT=$(docker-compose run --rm openfga-cli fga model write --api-url http://openfga:8080 --store-id $STORE_ID --file /workspace/model.json --format json)
MODEL_ID=$(echo $MODEL_OUTPUT | jq -r '.authorization_model_id')

if [ "$MODEL_ID" == "null" ] || [ -z "$MODEL_ID" ]; then
    error "Falha ao carregar modelo de autorização"
    echo "Output: $MODEL_OUTPUT"
    exit 1
fi

log "Modelo carregado com ID: $MODEL_ID"

log "Carregando dados iniciais..."
docker-compose run --rm openfga-cli fga tuple write --api-url http://openfga:8080 --store-id $STORE_ID --file /workspace/init-data.json

log "Atualizando arquivo .env.dev com IDs..."
sed -i "s/OPENFGA_STORE_ID=.*/OPENFGA_STORE_ID=$STORE_ID/" .env.dev
sed -i "s/OPENFGA_AUTH_MODEL_ID=.*/OPENFGA_AUTH_MODEL_ID=$MODEL_ID/" .env.dev

log "Iniciando todos os serviços..."
docker-compose up -d

echo ""
log "✅ Setup do OpenFGA concluído com sucesso!"
echo ""
echo -e "${YELLOW}Informações importantes:${NC}"
echo -e "  • Store ID: ${GREEN}$STORE_ID${NC}"
echo -e "  • Model ID: ${GREEN}$MODEL_ID${NC}"
echo -e "  • OpenFGA API: ${GREEN}http://localhost:8081${NC}"
echo -e "  • OpenFGA Playground: ${GREEN}http://localhost:3000${NC}"
echo -e "  • PostgreSQL: ${GREEN}localhost:5432${NC}"
echo -e "  • Redis: ${GREEN}localhost:6379${NC}"
echo ""
echo -e "${YELLOW}Para testar a autorização:${NC}"
echo -e "  curl -X POST http://localhost:8081/stores/$STORE_ID/check \\"
echo -e "    -H 'Content-Type: application/json' \\"
echo -e "    -d '{\"tuple_key\":{\"user\":\"user:demo-user-1\",\"relation\":\"read\",\"object\":\"vault:demo-vault-1\"}}'"
echo ""
```

## Integração com Backend Go

### 1. Configuração no config.yaml

```yaml
# config/config.yaml
openfga:
  api_url: "http://localhost:8081"
  store_id: "${OPENFGA_STORE_ID}"
  auth_model_id: "${OPENFGA_AUTH_MODEL_ID}"
  timeout: "5s"
  retry_attempts: 3
  retry_delay: "1s"
```

### 2. Client OpenFGA

**Arquivo**: `pkg/authorization/openfga/client.go`

```go
package openfga

import (
    "context"
    "fmt"
    "time"

    openfga "github.com/openfga/go-sdk"
    "github.com/openfga/go-sdk/client"
    "github.com/openfga/go-sdk/credentials"
)

type Client struct {
    fgaClient  *client.OpenFgaClient
    storeID    string
    modelID    string
}

type Config struct {
    APIURL         string        `yaml:"api_url"`
    StoreID        string        `yaml:"store_id"`
    AuthModelID    string        `yaml:"auth_model_id"`
    Timeout        time.Duration `yaml:"timeout"`
    RetryAttempts  int          `yaml:"retry_attempts"`
    RetryDelay     time.Duration `yaml:"retry_delay"`
}

func NewClient(cfg *Config) (*Client, error) {
    configuration := &client.Configuration{
        ApiUrl: cfg.APIURL,
        StoreId: cfg.StoreID,
        AuthorizationModelId: cfg.AuthModelID,
        Credentials: &credentials.Credentials{
            Method: credentials.CredentialsMethodNone,
        },
    }

    fgaClient := client.NewSdkClient(configuration)

    return &Client{
        fgaClient: fgaClient,
        storeID:   cfg.StoreID,
        modelID:   cfg.AuthModelID,
    }, nil
}

// CheckPermission verifica se um usuário tem permissão para um recurso
func (c *Client) CheckPermission(ctx context.Context, userID, relation, objectType, objectID string) (bool, error) {
    body := client.ClientCheckRequest{
        User:     fmt.Sprintf("user:%s", userID),
        Relation: relation,
        Object:   fmt.Sprintf("%s:%s", objectType, objectID),
    }

    data, err := c.fgaClient.Check(ctx).Body(body).Execute()
    if err != nil {
        return false, fmt.Errorf("failed to check permission: %w", err)
    }

    return data.GetAllowed(), nil
}

// WriteTuple adiciona uma relação de autorização
func (c *Client) WriteTuple(ctx context.Context, userID, relation, objectType, objectID string) error {
    body := client.ClientWriteRequest{
        Writes: []openfga.TupleKey{
            {
                User:     openfga.PtrString(fmt.Sprintf("user:%s", userID)),
                Relation: openfga.PtrString(relation),
                Object:   openfga.PtrString(fmt.Sprintf("%s:%s", objectType, objectID)),
            },
        },
    }

    _, err := c.fgaClient.Write(ctx).Body(body).Execute()
    if err != nil {
        return fmt.Errorf("failed to write tuple: %w", err)
    }

    return nil
}

// DeleteTuple remove uma relação de autorização
func (c *Client) DeleteTuple(ctx context.Context, userID, relation, objectType, objectID string) error {
    body := client.ClientWriteRequest{
        Deletes: []openfga.TupleKeyWithoutCondition{
            {
                User:     fmt.Sprintf("user:%s", userID),
                Relation: relation,
                Object:   fmt.Sprintf("%s:%s", objectType, objectID),
            },
        },
    }

    _, err := c.fgaClient.Write(ctx).Body(body).Execute()
    if err != nil {
        return fmt.Errorf("failed to delete tuple: %w", err)
    }

    return nil
}

// ListObjects lista objetos que um usuário tem acesso
func (c *Client) ListObjects(ctx context.Context, userID, relation, objectType string) ([]string, error) {
    body := client.ClientListObjectsRequest{
        User:     fmt.Sprintf("user:%s", userID),
        Relation: relation,
        Type:     objectType,
    }

    data, err := c.fgaClient.ListObjects(ctx).Body(body).Execute()
    if err != nil {
        return nil, fmt.Errorf("failed to list objects: %w", err)
    }

    return data.GetObjects(), nil
}
```

### 3. Middleware de Autorização

**Arquivo**: `internal/middleware/authorization.go`

```go
package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "your-project/pkg/authorization/openfga"
)

type AuthorizationMiddleware struct {
    fgaClient *openfga.Client
}

func NewAuthorizationMiddleware(fgaClient *openfga.Client) *AuthorizationMiddleware {
    return &AuthorizationMiddleware{
        fgaClient: fgaClient,
    }
}

// RequirePermission middleware que verifica permissões específicas
func (m *AuthorizationMiddleware) RequirePermission(relation, objectType string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
            c.Abort()
            return
        }

        // Extrair object ID do path parameter
        objectID := c.Param("id")
        if objectID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "object id required"})
            c.Abort()
            return
        }

        // Verificar permissão no OpenFGA
        allowed, err := m.fgaClient.CheckPermission(
            context.Background(),
            userID,
            relation,
            objectType,
            objectID,
        )
        
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization check failed"})
            c.Abort()
            return
        }

        if !allowed {
            c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// RequireVaultPermission middleware específico para vaults
func (m *AuthorizationMiddleware) RequireVaultPermission(relation string) gin.HandlerFunc {
    return m.RequirePermission(relation, "vault")
}

// RequireSecretPermission middleware específico para secrets
func (m *AuthorizationMiddleware) RequireSecretPermission(relation string) gin.HandlerFunc {
    return m.RequirePermission(relation, "secret")
}
```

### 4. Exemplo de Uso nos Handlers

**Arquivo**: `internal/handler/vault.go`

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "your-project/internal/middleware"
    "your-project/pkg/authorization/openfga"
)

type VaultHandler struct {
    fgaClient *openfga.Client
    authMW    *middleware.AuthorizationMiddleware
}

func NewVaultHandler(fgaClient *openfga.Client) *VaultHandler {
    return &VaultHandler{
        fgaClient: fgaClient,
        authMW:    middleware.NewAuthorizationMiddleware(fgaClient),
    }
}

func (h *VaultHandler) RegisterRoutes(r *gin.RouterGroup) {
    vaults := r.Group("/vaults")
    {
        vaults.GET("/", h.ListVaults)
        vaults.POST("/", h.CreateVault)
        vaults.GET("/:id", h.authMW.RequireVaultPermission("read"), h.GetVault)
        vaults.PUT("/:id", h.authMW.RequireVaultPermission("write"), h.UpdateVault)
        vaults.DELETE("/:id", h.authMW.RequireVaultPermission("delete"), h.DeleteVault)
        vaults.POST("/:id/share", h.authMW.RequireVaultPermission("share"), h.ShareVault)
    }
}

func (h *VaultHandler) CreateVault(c *gin.Context) {
    userID := c.GetString("user_id")
    
    // ... lógica de criação do vault ...
    vaultID := "new-vault-id" // ID do vault criado
    
    // Dar ownership ao usuário criador
    err := h.fgaClient.WriteTuple(c.Request.Context(), userID, "owner", "vault", vaultID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set permissions"})
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{"message": "vault created"})
}
```

### 5. Inicialização no main.go

**Arquivo**: `cmd/main.go`

```go
package main

import (
    "log"
    
    "your-project/config"
    "your-project/pkg/authorization/openfga"
    "your-project/internal/handler"
)

func main() {
    // Carregar configuração
    cfg := config.Load()
    
    // Inicializar cliente OpenFGA
    fgaClient, err := openfga.NewClient(&cfg.OpenFGA)
    if err != nil {
        log.Fatalf("Failed to initialize OpenFGA client: %v", err)
    }
    
    // Inicializar handlers
    vaultHandler := handler.NewVaultHandler(fgaClient)
    
    // ... resto da inicialização ...
}
```

## Comandos Úteis

### Inicializar o ambiente

```bash
# Dar permissão de execução ao script
chmod +x scripts/setup-openfga.sh

# Executar setup
./scripts/setup-openfga.sh
```

### Gerenciar containers

```bash
# Iniciar todos os serviços
docker-compose up -d

# Ver logs
docker-compose logs -f openfga

# Parar todos os serviços
docker-compose down

# Resetar completamente (dados serão perdidos)
docker-compose down -v
```

### Testar OpenFGA

```bash
# Verificar se um usuário pode ler um vault
curl -X POST http://localhost:8081/stores/$STORE_ID/check \
  -H 'Content-Type: application/json' \
  -d '{
    "tuple_key": {
      "user": "user:demo-user-1",
      "relation": "read",
      "object": "vault:demo-vault-1"
    }
  }'

# Listar objetos que um usuário pode acessar
curl -X POST http://localhost:8081/stores/$STORE_ID/list-objects \
  -H 'Content-Type: application/json' \
  -d '{
    "user": "user:demo-user-1",
    "relation": "read",
    "type": "vault"
  }'
```

## Fluxos de Autorização

### 1. Criação de Vault

```mermaid
sequenceDiagram
    participant F as Frontend
    participant B as Backend
    participant FGA as OpenFGA
    participant FS as Firestore

    F->>B: POST /vaults (JWT Token)
    B->>B: Validate JWT & Extract UserID
    B->>FS: Create vault document
    B->>FGA: WriteTuple(userID, "owner", "vault", vaultID)
    FGA-->>B: Success
    B-->>F: Vault created
```

### 2. Acesso a Secret

```mermaid
sequenceDiagram
    participant F as Frontend
    participant B as Backend
    participant FGA as OpenFGA
    participant FS as Firestore

    F->>B: GET /vaults/123/secrets/456 (JWT Token)
    B->>B: Validate JWT & Extract UserID
    B->>FGA: Check(userID, "read", "secret", secretID)
    FGA-->>B: Allowed: true
    B->>FS: Fetch secret data
    FS-->>B: Encrypted secret
    B->>B: Decrypt secret
    B-->>F: Secret data
```

### 3. Compartilhamento de Vault

```mermaid
sequenceDiagram
    participant F as Frontend
    participant B as Backend
    participant FGA as OpenFGA

    F->>B: POST /vaults/123/share (JWT Token)
    B->>B: Validate JWT & Extract UserID
    B->>FGA: Check(userID, "share", "vault", vaultID)
    FGA-->>B: Allowed: true
    B->>FGA: WriteTuple(targetUserID, "viewer", "vault", vaultID)
    FGA-->>B: Success
    B-->>F: Vault shared successfully
```

## Monitoramento e Métricas

### Prometheus Metrics

OpenFGA expõe métricas no endpoint `:2112/metrics`:

- `openfga_datastore_query_count_total`
- `openfga_request_duration_ms`
- `openfga_grpc_request_duration_ms`

### Health Checks

```bash
# OpenFGA Health
curl http://localhost:8081/healthz

# Verificar store
curl http://localhost:8081/stores

# Verificar modelo
curl http://localhost:8081/stores/$STORE_ID/authorization-models
```

## Considerações de Produção

### 1. Segurança

- Use HTTPS em produção
- Configure autenticação adequada para OpenFGA
- Implemente rate limiting
- Use TLS para comunicação entre serviços

### 2. Performance

- Configure índices adequados no PostgreSQL
- Use conexão pooling
- Implemente cache para consultas frequentes
- Monitore latência das queries

### 3. Backup e Recovery

```bash
# Backup do PostgreSQL
docker-compose exec postgres pg_dump -U openfga openfga > backup.sql

# Restore
docker-compose exec -T postgres psql -U openfga openfga < backup.sql
```

### 4. Escalabilidade

- Use cluster PostgreSQL para alta disponibilidade
- Configure Redis Cluster para cache distribuído
- Deploy OpenFGA em múltiplas instâncias com load balancer

## Próximos Passos

1. **Implementar client Go completo** no backend
2. **Criar testes de integração** entre backend e OpenFGA
3. **Documentar exemplos** de uso para diferentes cenários
4. **Implementar cache** para melhorar performance
5. **Configurar monitoramento** em produção

Esta configuração fornece uma base sólida para autorização baseada em relacionamentos usando OpenFGA, mantendo a separação clara entre autenticação (Firebase) e autorização (OpenFGA).
      # Database Configuration
      OPENFGA_DATASTORE_ENGINE: postgres
      OPENFGA_DATASTORE_URI: postgres://openfga:${OPENFGA_DB_PASSWORD}@postgres-openfga:5432/openfga?sslmode=disable
      
      # Server Configuration
      OPENFGA_HTTP_ADDR: 0.0.0.0:8080
      OPENFGA_GRPC_ADDR: 0.0.0.0:8081
      OPENFGA_GRPC_TLS_ENABLED: false
      
      # Authentication (Opcional para desenvolvimento)
      OPENFGA_AUTHN_METHOD: none  # ou 'preshared' para produção
      OPENFGA_AUTHN_PRESHARED_KEYS: ${OPENFGA_API_KEY}
      
      # Logging
      OPENFGA_LOG_LEVEL: info
      OPENFGA_LOG_FORMAT: json
      
      # Performance
      OPENFGA_DATASTORE_MAX_OPEN_CONNS: 30
      OPENFGA_DATASTORE_MAX_IDLE_CONNS: 10
      OPENFGA_DATASTORE_CONN_MAX_IDLE_TIME: 30s
      OPENFGA_DATASTORE_CONN_MAX_LIFETIME: 300s
      
      # Request limits
      OPENFGA_MAX_TUPLES_PER_WRITE: 100
      OPENFGA_MAX_TYPES_PER_AUTHORIZATION_MODEL: 100
      
    ports:
      - "8080:8080"  # HTTP API
      - "8081:8081"  # gRPC API
    depends_on:
      postgres-openfga:
        condition: service_healthy
    networks:
      - lockari-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
      interval: 10s
      timeout: 5s
      retries: 3
    restart: unless-stopped

  # Redis (já existente no seu projeto)
  redis:
    image: redis:7-alpine
    container_name: lockari-redis
    ports:
      - "6379:6379"
    networks:
      - lockari-network
    volumes:
      - redis_data:/data

  # Seu Backend Golang
  lockari-backend:
    build:
      context: ..
      dockerfile: build/Dockerfile
    container_name: lockari-backend
    environment:
      # Firebase Config
      FIREBASE_PROJECT_ID: ${FIREBASE_PROJECT_ID}
      FIREBASE_SERVICE_ACCOUNT_PATH: /app/config/firebase-service-account.json
      
      # OpenFGA Config
      OPENFGA_API_URL: http://openfga:8080
      OPENFGA_STORE_ID: ${OPENFGA_STORE_ID}
      OPENFGA_MODEL_ID: ${OPENFGA_MODEL_ID}
      OPENFGA_API_KEY: ${OPENFGA_API_KEY}
      
      # Database URLs
      REDIS_URL: redis://redis:6379
      
    ports:
      - "3000:3000"
    depends_on:
      - openfga
      - redis
    networks:
      - lockari-network
    volumes:
      - ../cmd/config/auth.json:/app/config/firebase-service-account.json:ro

networks:
  lockari-network:
    driver: bridge

volumes:
  postgres_openfga_data:
  redis_data:
```

### **3. Configuração para Desenvolvimento**

```yaml
# docker/docker-compose.dev.yml
version: '3.8'

services:
  openfga:
    environment:
      # Desenvolvimento: sem autenticação
      OPENFGA_AUTHN_METHOD: none
      OPENFGA_LOG_LEVEL: debug
    volumes:
      # Volume para facilitar debug
      - ./openfga/logs:/tmp/openfga-logs

  # OpenFGA CLI para testes e debug
  openfga-cli:
    image: openfga/cli:v0.2.5
    container_name: lockari-openfga-cli
    networks:
      - lockari-network
    volumes:
      - ./openfga:/workspace
    working_dir: /workspace
    # Mantém container rodando para execução de comandos
    command: tail -f /dev/null
```

### **4. Configuração para Produção**

```yaml
# docker/docker-compose.prod.yml
version: '3.8'

services:
  postgres-openfga:
    environment:
      # Produção: senha forte obrigatória
      POSTGRES_HOST_AUTH_METHOD: md5
    volumes:
      # Backup automático
      - /opt/backups/openfga:/backups

  openfga:
    environment:
      # Produção: autenticação obrigatória
      OPENFGA_AUTHN_METHOD: preshared
      OPENFGA_AUTHN_PRESHARED_KEYS: ${OPENFGA_API_KEY}
      
      # TLS habilitado
      OPENFGA_GRPC_TLS_ENABLED: true
      OPENFGA_HTTP_TLS_ENABLED: true
      OPENFGA_HTTP_TLS_CERT: /certs/server.crt
      OPENFGA_HTTP_TLS_KEY: /certs/server.key
      
      # Logging estruturado
      OPENFGA_LOG_LEVEL: warn
      OPENFGA_LOG_FORMAT: json
      
    volumes:
      - /etc/ssl/certs:/certs:ro
    
    # Não expor portas diretamente (usar reverse proxy)
    ports: []
```

---

## 📋 Arquivos de Configuração {#arquivos-config}

### **1. Variáveis de Ambiente (.env.dev)**

```bash
# .env.dev
# PostgreSQL OpenFGA
OPENFGA_DB_PASSWORD=dev_password_123

# OpenFGA Server
OPENFGA_API_KEY=dev_api_key_12345
OPENFGA_STORE_ID=01HXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
OPENFGA_MODEL_ID=01HYYY-YYYY-YYYY-YYYY-YYYYYYYYYYYY

# Firebase
FIREBASE_PROJECT_ID=lockari-vault

# Backend
BACKEND_PORT=3000
REDIS_URL=redis://localhost:6379
```

### **2. Inicialização PostgreSQL**

```sql
-- docker/openfga/postgres-init.sql
-- Criar database e usuário para OpenFGA
CREATE DATABASE openfga;
CREATE USER openfga WITH PASSWORD 'dev_password_123';
GRANT ALL PRIVILEGES ON DATABASE openfga TO openfga;

-- Configurações de performance
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET max_connections = 100;
ALTER SYSTEM SET shared_buffers = '256MB';

-- Criar extensões necessárias
\c openfga;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
```

---

## 🔧 Modelo de Autorização OpenFGA {#modelo-autorizacao}

### **1. Modelo DSL para Lockari**

```json
{
  "schema_version": "1.1",
  "type_definitions": [
    {
      "type": "tenant",
      "relations": {
        "member": {
          "this": {}
        },
        "admin": {
          "this": {}
        },
        "owner": {
          "this": {}
        }
      },
      "metadata": {
        "relations": {
          "member": {
            "directly_related_user_types": [{"type": "user"}]
          },
          "admin": {
            "directly_related_user_types": [{"type": "user"}]
          },
          "owner": {
            "directly_related_user_types": [{"type": "user"}]
          }
        }
      }
    },
    {
      "type": "vault",
      "relations": {
        "owner": {
          "this": {}
        },
        "admin": {
          "this": {}
        },
        "writer": {
          "this": {}
        },
        "viewer": {
          "this": {}
        },
        "tenant": {
          "this": {}
        },
        "can_read": {
          "union": {
            "child": [
              {"this": {}},
              {"computedUserset": {"relation": "viewer"}},
              {"computedUserset": {"relation": "writer"}},
              {"computedUserset": {"relation": "admin"}},
              {"computedUserset": {"relation": "owner"}},
              {
                "tupleToUserset": {
                  "tupleset": {"relation": "tenant"},
                  "computedUserset": {"relation": "member"}
                }
              }
            ]
          }
        },
        "can_write": {
          "union": {
            "child": [
              {"this": {}},
              {"computedUserset": {"relation": "writer"}},
              {"computedUserset": {"relation": "admin"}},
              {"computedUserset": {"relation": "owner"}}
            ]
          }
        },
        "can_delete": {
          "union": {
            "child": [
              {"computedUserset": {"relation": "admin"}},
              {"computedUserset": {"relation": "owner"}}
            ]
          }
        },
        "can_share": {
          "union": {
            "child": [
              {"computedUserset": {"relation": "admin"}},
              {"computedUserset": {"relation": "owner"}}
            ]
          }
        }
      },
      "metadata": {
        "relations": {
          "owner": {"directly_related_user_types": [{"type": "user"}]},
          "admin": {"directly_related_user_types": [{"type": "user"}]},
          "writer": {"directly_related_user_types": [{"type": "user"}]},
          "viewer": {"directly_related_user_types": [{"type": "user"}]},
          "tenant": {"directly_related_user_types": [{"type": "tenant"}]}
        }
      }
    },
    {
      "type": "secret",
      "relations": {
        "vault": {
          "this": {}
        },
        "can_read": {
          "tupleToUserset": {
            "tupleset": {"relation": "vault"},
            "computedUserset": {"relation": "can_read"}
          }
        },
        "can_write": {
          "tupleToUserset": {
            "tupleset": {"relation": "vault"},
            "computedUserset": {"relation": "can_write"}
          }
        },
        "can_delete": {
          "tupleToUserset": {
            "tupleset": {"relation": "vault"},
            "computedUserset": {"relation": "can_delete"}
          }
        }
      },
      "metadata": {
        "relations": {
          "vault": {"directly_related_user_types": [{"type": "vault"}]}
        }
      }
    },
    {
      "type": "certificate",
      "relations": {
        "vault": {
          "this": {}
        },
        "can_read": {
          "tupleToUserset": {
            "tupleset": {"relation": "vault"},
            "computedUserset": {"relation": "can_read"}
          }
        },
        "can_write": {
          "tupleToUserset": {
            "tupleset": {"relation": "vault"},
            "computedUserset": {"relation": "can_write"}
          }
        },
        "can_delete": {
          "tupleToUserset": {
            "tupleset": {"relation": "vault"},
            "computedUserset": {"relation": "can_delete"}
          }
        }
      },
      "metadata": {
        "relations": {
          "vault": {"directly_related_user_types": [{"type": "vault"}]}
        }
      }
    },
    {
      "type": "ssh_key",
      "relations": {
        "vault": {
          "this": {}
        },
        "can_read": {
          "tupleToUserset": {
            "tupleset": {"relation": "vault"},
            "computedUserset": {"relation": "can_read"}
          }
        },
        "can_write": {
          "tupleToUserset": {
            "tupleset": {"relation": "vault"},
            "computedUserset": {"relation": "can_write"}
          }
        },
        "can_delete": {
          "tupleToUserset": {
            "tupleset": {"relation": "vault"},
            "computedUserset": {"relation": "can_delete"}
          }
        }
      },
      "metadata": {
        "relations": {
          "vault": {"directly_related_user_types": [{"type": "vault"}]}
        }
      }
    }
  ]
}
```

### **2. Dados Iniciais de Teste**

```json
{
  "tuples": [
    {
      "key": {
        "user": "user:alice",
        "relation": "owner",
        "object": "tenant:company-abc"
      }
    },
    {
      "key": {
        "user": "user:bob",
        "relation": "member", 
        "object": "tenant:company-abc"
      }
    },
    {
      "key": {
        "user": "user:alice",
        "relation": "owner",
        "object": "vault:marketing-secrets"
      }
    },
    {
      "key": {
        "user": "tenant:company-abc",
        "relation": "tenant",
        "object": "vault:marketing-secrets"
      }
    },
    {
      "key": {
        "user": "vault:marketing-secrets",
        "relation": "vault",
        "object": "secret:api-key-prod"
      }
    }
  ]
}
```

---

## 🔗 Integração Backend Golang {#integracao-backend}

### **1. Configuração no config.yaml**

```yaml
# Adicionar ao config.yaml existente
openfga:
  api_url: "http://localhost:8080"
  store_id: "${OPENFGA_STORE_ID}"
  model_id: "${OPENFGA_MODEL_ID}"
  api_key: "${OPENFGA_API_KEY}"
  timeout: "30s"
  retry_attempts: 3
  cache_ttl: "300s"  # Cache de verificações por 5 minutos
```

### **2. Cliente OpenFGA**

```go
// pkg/authorization/openfga/client.go
package openfga

import (
    "context"
    "fmt"
    "time"
    
    openfga "github.com/openfga/go-sdk"
    "github.com/openfga/go-sdk/client"
)

type Config struct {
    APIURL    string `mapstructure:"api_url"`
    StoreID   string `mapstructure:"store_id"`
    ModelID   string `mapstructure:"model_id"`
    APIKey    string `mapstructure:"api_key"`
    Timeout   string `mapstructure:"timeout"`
    CacheTTL  string `mapstructure:"cache_ttl"`
}

type Client struct {
    fgaClient *client.OpenFgaClient
    config    *Config
    cache     map[string]cacheEntry // Cache simples para verificações
}

type cacheEntry struct {
    result    bool
    timestamp time.Time
}

func NewClient(config *Config) (*Client, error) {
    configuration, err := client.NewConfiguration(client.Configuration{
        ApiUrl:  config.APIURL,
        StoreId: config.StoreID,
        AuthorizationModelId: config.ModelID,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenFGA configuration: %w", err)
    }

    // Adicionar autenticação se necessário
    if config.APIKey != "" {
        configuration.Credentials = &client.Credentials{
            Method: client.CredentialsMethodApiToken,
            Config: &client.Config{
                ApiToken: config.APIKey,
            },
        }
    }

    fgaClient := client.NewSdkClient(configuration)

    return &Client{
        fgaClient: fgaClient,
        config:    config,
        cache:     make(map[string]cacheEntry),
    }, nil
}

// Check verifica se um usuário tem uma relação específica com um objeto
func (c *Client) Check(ctx context.Context, user, relation, object string) (bool, error) {
    // Verificar cache primeiro
    cacheKey := fmt.Sprintf("%s:%s:%s", user, relation, object)
    if entry, exists := c.cache[cacheKey]; exists {
        cacheTTL, _ := time.ParseDuration(c.config.CacheTTL)
        if time.Since(entry.timestamp) < cacheTTL {
            return entry.result, nil
        }
    }

    body := &client.ClientCheckRequest{
        User:     user,
        Relation: relation,
        Object:   object,
    }

    data, err := c.fgaClient.Check(ctx).Body(*body).Execute()
    if err != nil {
        return false, fmt.Errorf("OpenFGA check failed: %w", err)
    }

    result := data.GetAllowed()
    
    // Armazenar no cache
    c.cache[cacheKey] = cacheEntry{
        result:    result,
        timestamp: time.Now(),
    }

    return result, nil
}

// WriteTuple escreve uma nova tupla de relacionamento
func (c *Client) WriteTuple(ctx context.Context, user, relation, object string) error {
    body := &client.ClientWriteRequest{
        Writes: []client.ClientTupleKey{
            {
                User:     user,
                Relation: relation,
                Object:   object,
            },
        },
    }

    _, err := c.fgaClient.Write(ctx).Body(*body).Execute()
    if err != nil {
        return fmt.Errorf("failed to write tuple: %w", err)
    }

    // Limpar cache relacionado
    c.clearRelatedCache(user, relation, object)
    
    return nil
}

// DeleteTuple remove uma tupla de relacionamento
func (c *Client) DeleteTuple(ctx context.Context, user, relation, object string) error {
    body := &client.ClientWriteRequest{
        Deletes: []client.ClientTupleKey{
            {
                User:     user,
                Relation: relation,
                Object:   object,
            },
        },
    }

    _, err := c.fgaClient.Write(ctx).Body(*body).Execute()
    if err != nil {
        return fmt.Errorf("failed to delete tuple: %w", err)
    }

    // Limpar cache relacionado
    c.clearRelatedCache(user, relation, object)
    
    return nil
}

// ListObjects lista objetos que um usuário tem uma relação específica
func (c *Client) ListObjects(ctx context.Context, user, relation, objectType string) ([]string, error) {
    body := &client.ClientListObjectsRequest{
        User:     user,
        Relation: relation,
        Type:     objectType,
    }

    data, err := c.fgaClient.ListObjects(ctx).Body(*body).Execute()
    if err != nil {
        return nil, fmt.Errorf("failed to list objects: %w", err)
    }

    return data.GetObjects(), nil
}

func (c *Client) clearRelatedCache(user, relation, object string) {
    // Implementar lógica de limpeza de cache mais sofisticada se necessário
    for key := range c.cache {
        if contains(key, user) || contains(key, object) {
            delete(c.cache, key)
        }
    }
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && 
           (s == substr || 
            s[:len(substr)+1] == substr+":" ||
            s[len(s)-len(substr)-1:] == ":"+substr)
}
```

### **3. Serviço de Autorização**

```go
// pkg/authorization/service.go
package authorization

import (
    "context"
    "fmt"
    
    "lockari-api-application/pkg/authorization/openfga"
)

type Service struct {
    fgaClient *openfga.Client
}

func NewService(fgaClient *openfga.Client) *Service {
    return &Service{
        fgaClient: fgaClient,
    }
}

// Vault Permissions
func (s *Service) CanReadVault(ctx context.Context, userID, vaultID string) (bool, error) {
    return s.fgaClient.Check(ctx, 
        fmt.Sprintf("user:%s", userID), 
        "can_read", 
        fmt.Sprintf("vault:%s", vaultID))
}

func (s *Service) CanWriteVault(ctx context.Context, userID, vaultID string) (bool, error) {
    return s.fgaClient.Check(ctx, 
        fmt.Sprintf("user:%s", userID), 
        "can_write", 
        fmt.Sprintf("vault:%s", vaultID))
}

func (s *Service) CanDeleteVault(ctx context.Context, userID, vaultID string) (bool, error) {
    return s.fgaClient.Check(ctx, 
        fmt.Sprintf("user:%s", userID), 
        "can_delete", 
        fmt.Sprintf("vault:%s", vaultID))
}

func (s *Service) CanShareVault(ctx context.Context, userID, vaultID string) (bool, error) {
    return s.fgaClient.Check(ctx, 
        fmt.Sprintf("user:%s", userID), 
        "can_share", 
        fmt.Sprintf("vault:%s", vaultID))
}

// Secret Permissions
func (s *Service) CanReadSecret(ctx context.Context, userID, secretID string) (bool, error) {
    return s.fgaClient.Check(ctx, 
        fmt.Sprintf("user:%s", userID), 
        "can_read", 
        fmt.Sprintf("secret:%s", secretID))
}

// Tenant Permissions
func (s *Service) IsTenantMember(ctx context.Context, userID, tenantID string) (bool, error) {
    return s.fgaClient.Check(ctx, 
        fmt.Sprintf("user:%s", userID), 
        "member", 
        fmt.Sprintf("tenant:%s", tenantID))
}

// Setup Methods for New Resources
func (s *Service) SetupNewTenant(ctx context.Context, ownerID, tenantID string) error {
    return s.fgaClient.WriteTuple(ctx, 
        fmt.Sprintf("user:%s", ownerID), 
        "owner", 
        fmt.Sprintf("tenant:%s", tenantID))
}

func (s *Service) SetupNewVault(ctx context.Context, ownerID, vaultID, tenantID string) error {
    // Usuário é owner do vault
    err := s.fgaClient.WriteTuple(ctx, 
        fmt.Sprintf("user:%s", ownerID), 
        "owner", 
        fmt.Sprintf("vault:%s", vaultID))
    if err != nil {
        return err
    }

    // Vault pertence ao tenant
    return s.fgaClient.WriteTuple(ctx, 
        fmt.Sprintf("tenant:%s", tenantID), 
        "tenant", 
        fmt.Sprintf("vault:%s", vaultID))
}

func (s *Service) SetupNewSecret(ctx context.Context, secretID, vaultID string) error {
    return s.fgaClient.WriteTuple(ctx, 
        fmt.Sprintf("vault:%s", vaultID), 
        "vault", 
        fmt.Sprintf("secret:%s", secretID))
}

// List Methods
func (s *Service) ListUserVaults(ctx context.Context, userID string) ([]string, error) {
    return s.fgaClient.ListObjects(ctx, 
        fmt.Sprintf("user:%s", userID), 
        "can_read", 
        "vault")
}
```

### **4. Middleware de Autorização**

```go
// internal/handler/middleware/authorization.go
package middleware

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "lockari-api-application/pkg/authorization"
)

type AuthorizationMiddleware struct {
    authzService *authorization.Service
}

func NewAuthorizationMiddleware(authzService *authorization.Service) *AuthorizationMiddleware {
    return &AuthorizationMiddleware{
        authzService: authzService,
    }
}

// RequireVaultAccess middleware para verificar acesso a vault
func (m *AuthorizationMiddleware) RequireVaultAccess(action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        vaultID := c.Param("vaultId")
        
        if userID == "" || vaultID == "" {
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                "error": "Missing user_id or vault_id",
            })
            return
        }

        var hasAccess bool
        var err error

        switch action {
        case "read":
            hasAccess, err = m.authzService.CanReadVault(c.Request.Context(), userID, vaultID)
        case "write":
            hasAccess, err = m.authzService.CanWriteVault(c.Request.Context(), userID, vaultID)
        case "delete":
            hasAccess, err = m.authzService.CanDeleteVault(c.Request.Context(), userID, vaultID)
        case "share":
            hasAccess, err = m.authzService.CanShareVault(c.Request.Context(), userID, vaultID)
        default:
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                "error": "Invalid action",
            })
            return
        }

        if err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                "error": "Authorization check failed",
            })
            return
        }

        if !hasAccess {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error": "Access denied",
            })
            return
        }

        c.Next()
    }
}

// RequireTenantMember middleware para verificar se usuário é membro do tenant
func (m *AuthorizationMiddleware) RequireTenantMember() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        tenantID := c.GetString("tenant_id")
        
        if userID == "" || tenantID == "" {
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                "error": "Missing user_id or tenant_id",
            })
            return
        }

        isMember, err := m.authzService.IsTenantMember(c.Request.Context(), userID, tenantID)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                "error": "Tenant membership check failed",
            })
            return
        }

        if !isMember {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error": "Not a member of this tenant",
            })
            return
        }

        c.Next()
    }
}
```

### **5. Inicialização no main.go**

```go
// Adicionar ao main.go existente
func initializeOpenFGA(fields interface{}) (*authorization.Service, error) {
    var config openfga.Config
    
    b, err := json.Marshal(fields)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal OpenFGA config: %w", err)
    }

    if err := json.Unmarshal(b, &config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal OpenFGA config: %w", err)
    }

    fgaClient, err := openfga.NewClient(&config)
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenFGA client: %w", err)
    }

    authzService := authorization.NewService(fgaClient)
    
    return authzService, nil
}

// No main() function
func main() {
    // ...código existente...

    // Inicializar OpenFGA
    authzService, err := initializeOpenFGA(cfg.Fields["openfga"])
    if err != nil {
        log.Fatal(err)
    }

    // Criar middleware de autorização
    authzMiddleware := middleware.NewAuthorizationMiddleware(authzService)

    // Passar para os handlers
    webhandler.InitializeVaultHandler(vaultSvc, crypt, authClient, tokenJWT, authzService, apiResponse.RouterGroup, authzMiddleware)
}
```

---

## 🔄 Fluxo de Autenticação Completo {#fluxo-autenticacao}

### **1. Registro de Novo Usuário**

```
1. Frontend → Firebase Auth (criar usuário)
2. Frontend → Backend (/api/v1/auth/signup)
3. Backend → Validate Firebase JWT
4. Backend → Generate tenantId (UUID v7)
5. Backend → Firebase (add tenantId to custom claims)
6. Backend → OpenFGA (user:alice#owner@tenant:abc123)
7. Backend → Firestore (create tenant document)
8. Backend → Response (success)
```

### **2. Login de Usuário**

```
1. Frontend → Firebase Auth (login)
2. Frontend → Receive JWT with tenantId
3. Frontend → Backend (with JWT header)
4. Backend → Validate JWT + extract tenantId
5. Backend → Store in context (user_id, tenant_id)
6. Middleware → OpenFGA check tenant membership
7. Continue to protected routes
```

### **3. Operação em Vault**

```
1. Frontend → GET /api/v1/vaults/123
2. Backend → Extract user_id from JWT
3. Backend → OpenFGA.Check(user:alice, can_read, vault:123)
4. OpenFGA → Check computed relationships
5. OpenFGA → Return true/false
6. Backend → Allow/deny operation
7. Backend → Log audit event
```

---

## 🚀 Scripts de Setup {#scripts}

### **1. Script de Inicialização**

```bash
#!/bin/bash
# scripts/setup-openfga.sh

set -e

echo "🚀 Setting up OpenFGA for Lockari..."

# Verificar se docker-compose está disponível
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose not found. Please install it first."
    exit 1
fi

# Definir ambiente
ENV=${1:-dev}
echo "📁 Environment: $ENV"

# Carregar variáveis de ambiente
if [ -f ".env.$ENV" ]; then
    export $(cat .env.$ENV | grep -v '#' | xargs)
    echo "✅ Loaded .env.$ENV"
else
    echo "❌ .env.$ENV file not found"
    exit 1
fi

# Iniciar serviços
echo "🐳 Starting services..."
docker-compose -f docker/docker-compose.yml -f docker/docker-compose.$ENV.yml up -d

# Aguardar OpenFGA estar pronto
echo "⏳ Waiting for OpenFGA to be ready..."
timeout=60
while [ $timeout -gt 0 ]; do
    if curl -f -s http://localhost:8080/healthz > /dev/null 2>&1; then
        echo "✅ OpenFGA is ready!"
        break
    fi
    sleep 2
    timeout=$((timeout-2))
done

if [ $timeout -le 0 ]; then
    echo "❌ OpenFGA failed to start within 60 seconds"
    exit 1
fi

# Criar store se não existir
echo "🏪 Setting up OpenFGA store..."
STORE_RESPONSE=$(curl -s -X POST http://localhost:8080/stores \
    -H "Content-Type: application/json" \
    -d '{"name": "lockari-vault"}')

STORE_ID=$(echo $STORE_RESPONSE | jq -r '.id // empty')

if [ -z "$STORE_ID" ]; then
    echo "❌ Failed to create OpenFGA store"
    echo "Response: $STORE_RESPONSE"
    exit 1
fi

echo "✅ Store created with ID: $STORE_ID"

# Atualizar .env com STORE_ID
sed -i "s/OPENFGA_STORE_ID=.*/OPENFGA_STORE_ID=$STORE_ID/" .env.$ENV

# Carregar modelo de autorização
echo "📋 Loading authorization model..."
if [ -f "docker/openfga/model.json" ]; then
    MODEL_RESPONSE=$(curl -s -X POST "http://localhost:8080/stores/$STORE_ID/authorization-models" \
        -H "Content-Type: application/json" \
        -d @docker/openfga/model.json)
    
    MODEL_ID=$(echo $MODEL_RESPONSE | jq -r '.authorization_model_id // empty')
    
    if [ -z "$MODEL_ID" ]; then
        echo "❌ Failed to load authorization model"
        echo "Response: $MODEL_RESPONSE"
        exit 1
    fi
    
    echo "✅ Model loaded with ID: $MODEL_ID"
    
    # Atualizar .env com MODEL_ID
    sed -i "s/OPENFGA_MODEL_ID=.*/OPENFGA_MODEL_ID=$MODEL_ID/" .env.$ENV
else
    echo "❌ Authorization model file not found: docker/openfga/model.json"
    exit 1
fi

# Carregar dados iniciais se existirem
if [ -f "docker/openfga/init-data.json" ]; then
    echo "📊 Loading initial data..."
    curl -s -X POST "http://localhost:8080/stores/$STORE_ID/write" \
        -H "Content-Type: application/json" \
        -d @docker/openfga/init-data.json > /dev/null
    echo "✅ Initial data loaded"
fi

echo ""
echo "🎉 OpenFGA setup completed successfully!"
echo ""
echo "📋 Configuration:"
echo "   Store ID: $STORE_ID"
echo "   Model ID: $MODEL_ID"
echo "   API URL:  http://localhost:8080"
echo ""
echo "🔧 Next steps:"
echo "   1. Update your backend configuration with these IDs"
echo "   2. Test the setup with: ./scripts/test-openfga.sh"
echo "   3. View logs with: docker-compose logs openfga"
echo ""
```

### **2. Script de Teste**

```bash
#!/bin/bash
# scripts/test-openfga.sh

set -e

echo "🧪 Testing OpenFGA setup..."

# Verificar se OpenFGA está rodando
if ! curl -f -s http://localhost:8080/healthz > /dev/null; then
    echo "❌ OpenFGA is not running. Run ./scripts/setup-openfga.sh first"
    exit 1
fi

# Carregar variáveis
if [ -f ".env.dev" ]; then
    export $(cat .env.dev | grep -v '#' | xargs)
else
    echo "❌ .env.dev file not found"
    exit 1
fi

echo "✅ OpenFGA is running"
echo "🏪 Store ID: $OPENFGA_STORE_ID"

# Teste 1: Verificar store
echo "🔍 Testing store access..."
STORE_INFO=$(curl -s "http://localhost:8080/stores/$OPENFGA_STORE_ID")
STORE_NAME=$(echo $STORE_INFO | jq -r '.name // "unknown"')
echo "✅ Store name: $STORE_NAME"

# Teste 2: Listar modelos
echo "🔍 Testing authorization models..."
MODELS=$(curl -s "http://localhost:8080/stores/$OPENFGA_STORE_ID/authorization-models")
MODEL_COUNT=$(echo $MODELS | jq '.authorization_models | length')
echo "✅ Found $MODEL_COUNT authorization model(s)"

# Teste 3: Escrever tupla de teste
echo "🔍 Testing tuple write..."
WRITE_RESPONSE=$(curl -s -X POST "http://localhost:8080/stores/$OPENFGA_STORE_ID/write" \
    -H "Content-Type: application/json" \
    -d '{
        "writes": [
            {
                "user": "user:test-user",
                "relation": "member",
                "object": "tenant:test-tenant"
            }
        ]
    }')

if echo $WRITE_RESPONSE | jq -e '.code' > /dev/null; then
    echo "❌ Failed to write test tuple:"
    echo $WRITE_RESPONSE | jq '.'
    exit 1
fi

echo "✅ Test tuple written successfully"

# Teste 4: Verificar permissão
echo "🔍 Testing permission check..."
CHECK_RESPONSE=$(curl -s -X POST "http://localhost:8080/stores/$OPENFGA_STORE_ID/check" \
    -H "Content-Type: application/json" \
    -d '{
        "user": "user:test-user",
        "relation": "member",
        "object": "tenant:test-tenant"
    }')

ALLOWED=$(echo $CHECK_RESPONSE | jq -r '.allowed // false')

if [ "$ALLOWED" = "true" ]; then
    echo "✅ Permission check successful (allowed: $ALLOWED)"
else
    echo "❌ Permission check failed (allowed: $ALLOWED)"
    exit 1
fi

# Teste 5: Limpar tupla de teste
echo "🧹 Cleaning up test data..."
curl -s -X POST "http://localhost:8080/stores/$OPENFGA_STORE_ID/write" \
    -H "Content-Type: application/json" \
    -d '{
        "deletes": [
            {
                "user": "user:test-user",
                "relation": "member",
                "object": "tenant:test-tenant"
            }
        ]
    }' > /dev/null

echo "✅ Test data cleaned up"

echo ""
echo "🎉 All OpenFGA tests passed!"
echo ""
echo "🔧 You can now:"
echo "   1. Start your backend with OpenFGA integration"
echo "   2. Use the OpenFGA CLI: docker exec -it lockari-openfga-cli fga"
echo "   3. View OpenFGA logs: docker-compose logs openfga"
echo ""
```

---

## 🔍 Monitoramento e Debug {#monitoramento}

### **1. Health Checks**

```go
// internal/handler/web/health/openfga.go
func (h *healthHandler) CheckOpenFGA(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    // Teste simples de conectividade
    allowed, err := h.authzService.Check(ctx, "user:health-check", "member", "tenant:health-check")
    
    status := gin.H{
        "service": "openfga",
        "status":  "healthy",
        "timestamp": time.Now().UTC(),
    }

    if err != nil {
        status["status"] = "unhealthy"
        status["error"] = err.Error()
        c.JSON(503, status)
        return
    }

    status["check_result"] = allowed
    c.JSON(200, status)
}
```

### **2. Métricas Prometheus**

```go
// pkg/authorization/metrics.go
package authorization

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    openfgaCheckDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "openfga_check_duration_seconds",
            Help: "Duration of OpenFGA check operations",
        },
        []string{"relation", "object_type", "result"},
    )

    openfgaCheckTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "openfga_check_total",
            Help: "Total number of OpenFGA check operations",
        },
        []string{"relation", "object_type", "result"},
    )

    openfgaTupleWriteTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "openfga_tuple_write_total",
            Help: "Total number of OpenFGA tuple write operations",
        },
        []string{"operation", "object_type", "result"},
    )
)

// Instrumentar as operações
func (c *Client) Check(ctx context.Context, user, relation, object string) (bool, error) {
    start := time.Now()
    
    result, err := c.checkInternal(ctx, user, relation, object)
    
    duration := time.Since(start)
    objectType := extractObjectType(object)
    resultStatus := "success"
    if err != nil {
        resultStatus = "error"
    }

    openfgaCheckDuration.WithLabelValues(relation, objectType, resultStatus).Observe(duration.Seconds())
    openfgaCheckTotal.WithLabelValues(relation, objectType, resultStatus).Inc()

    return result, err
}
```

### **3. Logging Estruturado**

```go
// pkg/authorization/logging.go
import (
    "log/slog"
)

func (c *Client) Check(ctx context.Context, user, relation, object string) (bool, error) {
    logger := slog.With(
        "service", "openfga",
        "operation", "check",
        "user", user,
        "relation", relation,
        "object", object,
    )

    logger.Debug("Starting permission check")

    start := time.Now()
    result, err := c.checkInternal(ctx, user, relation, object)
    duration := time.Since(start)

    if err != nil {
        logger.Error("Permission check failed",
            "error", err,
            "duration_ms", duration.Milliseconds())
        return false, err
    }

    logger.Info("Permission check completed",
        "allowed", result,
        "duration_ms", duration.Milliseconds())

    return result, nil
}
```

---

## 🔒 Configuração de Produção {#configuracao-producao}

### **1. Deploy no Google Cloud Run**

```yaml
# deploy/openfga-cloudrun.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: lockari-openfga
  annotations:
    run.googleapis.com/ingress: internal  # Apenas tráfego interno
spec:
  template:
    metadata:
      annotations:
        run.googleapis.com/max-scale: "10"
        run.googleapis.com/min-scale: "1"
        run.googleapis.com/memory: "512Mi"
        run.googleapis.com/cpu: "1000m"
    spec:
      containers:
      - image: openfga/openfga:v1.4.3
        env:
        - name: OPENFGA_DATASTORE_ENGINE
          value: postgres
        - name: OPENFGA_DATASTORE_URI
          valueFrom:
            secretKeyRef:
              name: openfga-db-connection
              key: uri
        - name: OPENFGA_AUTHN_METHOD
          value: preshared
        - name: OPENFGA_AUTHN_PRESHARED_KEYS
          valueFrom:
            secretKeyRef:
              name: openfga-auth
              key: api-key
        - name: OPENFGA_LOG_LEVEL
          value: warn
        - name: OPENFGA_LOG_FORMAT
          value: json
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "256Mi"
            cpu: "500m"
          limits:
            memory: "512Mi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

### **2. Terraform para Infraestrutura**

```hcl
# terraform/openfga.tf
resource "google_cloud_run_service" "openfga" {
  name     = "lockari-openfga"
  location = var.region

  template {
    spec {
      containers {
        image = "openfga/openfga:v1.4.3"
        
        env {
          name  = "OPENFGA_DATASTORE_ENGINE"
          value = "postgres"
        }
        
        env {
          name = "OPENFGA_DATASTORE_URI"
          value_from {
            secret_key_ref {
              name = google_secret_manager_secret_version.openfga_db_uri.secret_id
              key  = "latest"
            }
          }
        }

        resources {
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
          requests = {
            cpu    = "500m"
            memory = "256Mi"
          }
        }

        ports {
          container_port = 8080
        }
      }
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = "10"
        "autoscaling.knative.dev/minScale" = "1"
        "run.googleapis.com/ingress"       = "internal"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

# Banco PostgreSQL Cloud SQL
resource "google_sql_database_instance" "openfga" {
  name             = "lockari-openfga-db"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = "db-f1-micro"  # Para produção, usar tier maior

    database_flags {
      name  = "max_connections"
      value = "100"
    }

    backup_configuration {
      enabled    = true
      start_time = "03:00"
    }

    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.vpc.id
    }
  }
}

resource "google_sql_database" "openfga" {
  name     = "openfga"
  instance = google_sql_database_instance.openfga.name
}
```

### **3. Secrets Management**

```bash
# Criar secrets no Google Secret Manager
gcloud secrets create openfga-api-key --data-file=- <<< "your-secure-api-key-here"
gcloud secrets create openfga-db-uri --data-file=- <<< "postgres://user:pass@db:5432/openfga"

# Dar permissão para Cloud Run acessar os secrets
gcloud secrets add-iam-policy-binding openfga-api-key \
    --member="serviceAccount:your-cloudrun-sa@project.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"
```

---

## ✅ Checklist de Implementação

### **Fase 1: Setup Básico**
- [ ] Configurar Docker Compose com PostgreSQL e OpenFGA
- [ ] Criar modelo de autorização DSL
- [ ] Implementar cliente OpenFGA em Go
- [ ] Configurar health checks
- [ ] Testes básicos de conectividade

### **Fase 2: Integração Backend**
- [ ] Integrar OpenFGA ao signup process
- [ ] Criar middleware de autorização
- [ ] Implementar verificações de vault/secret
- [ ] Adicionar logging e métricas
- [ ] Testes de integração

### **Fase 3: Produção**
- [ ] Deploy OpenFGA no Cloud Run
- [ ] Configurar Cloud SQL PostgreSQL
- [ ] Setup secrets management
- [ ] Monitoramento com Prometheus
- [ ] Backup e disaster recovery

### **Fase 4: Otimização**
- [ ] Cache de verificações frequentes
- [ ] Rate limiting
- [ ] Performance tuning
- [ ] Documentação completa
- [ ] Treinamento da equipe

---

## 🚀 Comandos Rápidos

```bash
# Setup inicial
./scripts/setup-openfga.sh dev

# Testar configuração
./scripts/test-openfga.sh

# Logs OpenFGA
docker-compose logs -f openfga

# CLI OpenFGA para debug
docker exec -it lockari-openfga-cli fga model get --store-id=$OPENFGA_STORE_ID

# Verificar tuplas
docker exec -it lockari-openfga-cli fga tuple read --store-id=$OPENFGA_STORE_ID

# Testar permissão manualmente
curl -X POST http://localhost:8080/stores/$OPENFGA_STORE_ID/check \
  -H "Content-Type: application/json" \
  -d '{"user":"user:alice","relation":"can_read","object":"vault:123"}'
```

---

Este documento fornece uma configuração completa para integrar OpenFGA ao seu projeto Lockari. O OpenFGA funciona como um serviço de autorização separado, complementando (não substituindo) o Firebase Authentication, e seu backend Golang atua como a ponte entre todos os sistemas.
