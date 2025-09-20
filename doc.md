# Organização Recomendada para Projetos Go com Múltiplos Bancos de Dados e Handlers HTTP

##
Estrutura Recomendada:
cmd/
├── api/
│   └── main.go           # Servidor HTTP/REST API
├── cli/
│   └── main.go           # Interface de linha de comando
├── worker/
│   └── main.go           # Background jobs/workers
├── migrate/
│   └── main.go           # Migrations de banco
└── seed/
    └── main.go           # Popular banco com dados iniciais

    Implementações:
1. API Server (cmd/api/main.go)
gopackage main

import (
    "log"
    "net/http"
    
    "projeto/internal/core/service"
    "projeto/internal/infrastructure/database"
    "projeto/internal/infrastructure/http/handler"
    "projeto/internal/infrastructure/http/router"
    "projeto/config"
    
    "github.com/gin-gonic/gin"
)

func main() {
    // Load config
    cfg := config.Load()
    
    // Database
    userRepo, err := database.NewUserRepository(cfg)
    if err != nil {
        log.Fatal("Failed to create user repository:", err)
    }
    
    // Services
    userService := service.NewUserService(userRepo)
    
    // Handlers
    userHandler := handler.NewUserHandler(userService)
    
    // Router
    r := gin.Default()
    router.SetupUserRoutes(r, userHandler)
    
    // Start server
    log.Printf("API Server starting on port %s", cfg.Server.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, r))
}
2. CLI Application (cmd/cli/main.go)
gopackage main

import (
    "context"
    "fmt"
    "os"
    
    "projeto/internal/core/service"
    "projeto/internal/infrastructure/database"
    "projeto/config"
    
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "My application CLI",
}

var userCmd = &cobra.Command{
    Use:   "user",
    Short: "User management commands",
}

var createUserCmd = &cobra.Command{
    Use:   "create",
    Short: "Create a new user",
    Run:   createUser,
}

func init() {
    createUserCmd.Flags().String("name", "", "User name")
    createUserCmd.Flags().String("email", "", "User email")
    createUserCmd.MarkFlagRequired("name")
    createUserCmd.MarkFlagRequired("email")
    
    userCmd.AddCommand(createUserCmd)
    rootCmd.AddCommand(userCmd)
}

func createUser(cmd *cobra.Command, args []string) {
    cfg := config.Load()
    
    // Same dependencies as API
    userRepo, err := database.NewUserRepository(cfg)
    if err != nil {
        fmt.Printf("Error creating repository: %v\n", err)
        os.Exit(1)
    }
    
    userService := service.NewUserService(userRepo)
    
    name, _ := cmd.Flags().GetString("name")
    email, _ := cmd.Flags().GetString("email")
    
    user, err := userService.CreateUser(context.Background(), name, email)
    if err != nil {
        fmt.Printf("Error creating user: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("User created successfully: %+v\n", user)
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
3. Worker (cmd/worker/main.go)
gopackage main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "projeto/internal/core/service"
    "projeto/internal/infrastructure/database"
    "projeto/internal/infrastructure/messaging"
    "projeto/config"
)

func main() {
    cfg := config.Load()
    
    // Dependencies
    userRepo, err := database.NewUserRepository(cfg)
    if err != nil {
        log.Fatal("Failed to create repository:", err)
    }
    
    userService := service.NewUserService(userRepo)
    
    // Message queue
    messageHandler := messaging.NewUserMessageHandler(userService)
    consumer, err := messaging.NewConsumer(cfg.Queue.DSN, messageHandler)
    if err != nil {
        log.Fatal("Failed to create consumer:", err)
    }
    
    // Start worker
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go consumer.Start(ctx)
    
    // Graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    log.Println("Worker started. Press Ctrl+C to exit...")
    <-c
    log.Println("Shutting down worker...")
    cancel()
}
4. Migration (cmd/migrate/main.go)
gopackage main

import (
    "flag"
    "log"
    
    "projeto/internal/infrastructure/database/migrations"
    "projeto/config"
)

func main() {
    var action = flag.String("action", "", "up or down")
    var version = flag.Int("version", 0, "Migration version")
    flag.Parse()
    
    if *action == "" {
        log.Fatal("Please specify -action (up or down)")
    }
    
    cfg := config.Load()
    
    migrator, err := migrations.NewMigrator(cfg)
    if err != nil {
        log.Fatal("Failed to create migrator:", err)
    }
    
    switch *action {
    case "up":
        if *version > 0 {
            err = migrator.MigrateToVersion(*version)
        } else {
            err = migrator.MigrateUp()
        }
    case "down":
        if *version > 0 {
            err = migrator.MigrateToVersion(*version)
        } else {
            err = migrator.MigrateDown()
        }
    default:
        log.Fatal("Invalid action. Use 'up' or 'down'")
    }
    
    if err != nil {
        log.Fatal("Migration failed:", err)
    }
    
    log.Println("Migration completed successfully")
}
Como Usar:
Compilar e executar:
bash# Compilar todos os binários
go build -o bin/api cmd/api/main.go
go build -o bin/cli cmd/cli/main.go
go build -o bin/worker cmd/worker/main.go
go build -o bin/migrate cmd/migrate/main.go

# Executar
./bin/api                                    # Inicia API server
./bin/cli user create --name="João" --email="joao@email.com"  # CLI
./bin/worker                                 # Inicia worker
./bin/migrate -action=up                     # Executa migrations
Makefile para automatizar:
makefile# Makefile
.PHONY: build-all build-api build-cli build-worker run-api run-worker

build-all: build-api build-cli build-worker build-migrate

build-api:
	go build -o bin/api cmd/api/main.go

build-cli:
	go build -o bin/cli cmd/cli/main.go

build-worker:
	go build -o bin/worker cmd/worker/main.go

build-migrate:
	go build -o bin/migrate cmd/migrate/main.go

run-api:
	go run cmd/api/main.go

run-cli:
	go run cmd/cli/main.go

run-worker:
	go run cmd/worker/main.go

migrate-up:
	go run cmd/migrate/main.go -action=up

migrate-down:
	go run cmd/migrate/main.go -action=down
Vantagens desta Estrutura:

Separação clara: Cada executável tem seu próprio entry point
Reutilização: Todos compartilham o mesmo core (services, repositories)
Deploy flexível: Pode deployar apenas o que precisa
Manutenibilidade: Fácil de manter e estender

Sua estrutura final ficaria assim:
projeto/
├── cmd/
│   ├── api/main.go
│   ├── cli/main.go
│   └── worker/main.go
├── internal/
│   ├── core/
│   └── infrastructure/
└── pkg/
Isso responde sua dúvida sobre a estrutura do /cmd?
##
internal/handler/ → internal/infrastructure/http/

Estrutura Recomendada (baseada na sua):
internal/
├── core/
│   ├── entity/
│   │   ├── user.go
│   │   └── product.go
│   ├── service/
│   │   ├── user_service.go
│   │   └── product_service.go
│   └── repository/
│       ├── user_repository.go      # Interface
│       └── product_repository.go   # Interface
├── infrastructure/
│   ├── database/
│   │   ├── mongodb/
│   │   │   ├── user_repository.go  # Implementação
│   │   │   └── user_adapter.go
│   │   └── postgresql/
│   │       ├── user_repository.go
│   │       └── user_adapter.go
│   └── http/                       # ← Mova seus handlers aqui
│       ├── handler/
│       │   ├── user_handler.go
│       │   └── product_handler.go
│       ├── middleware/
│       └── router/
└── pkg/                           # Se necessário
    └── database/
        ├── mongodb/
        └── postgresql/
Por que mover handler para infrastructure/http/?
1. Consistência Arquitetural
Handlers HTTP são parte da camada de infrastructure - são adaptadores que traduzem HTTP para chamadas de service.
2. Padrão da Indústria
go// Hexagonal Architecture
infrastructure/
├── web/          # HTTP handlers
├── persistence/  # Database
└── messaging/    # Queues
3. Escalabilidade
goinfrastructure/
├── http/         # REST API
├── grpc/         # gRPC handlers  
├── graphql/      # GraphQL resolvers
├── cli/          # CLI commands
└── database/     # Repositories
Implementação com Sua Estrutura:
Core (Domínio)
go// internal/core/entity/user.go
package entity

type User struct {
    ID        string
    Name      string
    Email     string
    CreatedAt time.Time
}
go// internal/core/repository/user_repository.go
package repository

import "internal/core/entity"

type UserRepository interface {
    Create(ctx context.Context, user *entity.User) error
    GetByID(ctx context.Context, id string) (*entity.User, error)
}
go// internal/core/service/user_service.go
package service

import (
    "internal/core/entity"
    "internal/core/repository"
)

type UserService struct {
    userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
    return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*entity.User, error) {
    user := &entity.User{Name: name, Email: email}
    err := s.userRepo.Create(ctx, user)
    return user, err
}
Infrastructure (Implementações)
go// internal/infrastructure/database/mongodb/user_repository.go
package mongodb

import (
    "internal/core/entity"
    "internal/core/repository"
    mongoClient "pkg/database/mongodb"
)

type userRepository struct {
    client  *mongoClient.Client
    adapter *UserAdapter
}

func NewUserRepository(client *mongoClient.Client) repository.UserRepository {
    return &userRepository{
        client:  client,
        adapter: NewUserAdapter(),
    }
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
    doc := r.adapter.EntityToDocument(user)
    _, err := r.client.Database().Collection("users").InsertOne(ctx, doc)
    return err
}
HTTP Layer
go// internal/infrastructure/http/handler/user_handler.go
package handler

import (
    "internal/core/service"
    "github.com/gin-gonic/gin"
)

type UserHandler struct {
    userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    // HTTP logic here
    user, err := h.userService.CreateUser(c.Request.Context(), name, email)
    // Handle response...
}
Conclusão:
Sua estrutura está 90% correta! Apenas considere:

✅ Manter: internal/core/ (excelente organização)
✅ Manter: Separação entity/service/repository
🔄 Ajustar: internal/handler/ → internal/infrastructure/http/handler/

Esta mudança torna sua arquitetura mais consistente e facilita futuras expansões (gRPC, CLI, etc.).
Sua abordagem com /pkg/ para clientes de banco + adapters na infrastructure está perfeita!