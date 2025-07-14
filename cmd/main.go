package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"github.com/synera-br/lockari-backend-app/config"

	// AUhtneitcation
	entity "github.com/synera-br/lockari-backend-app/internal/core/entity/auth"
	entity_auth "github.com/synera-br/lockari-backend-app/internal/core/entity/auth"
	repo_auth "github.com/synera-br/lockari-backend-app/internal/core/repository/auth"
	svc_auth "github.com/synera-br/lockari-backend-app/internal/core/service/auth"
	webhandler "github.com/synera-br/lockari-backend-app/internal/handler/web/auth"

	// AUDIT
	entity_audit "github.com/synera-br/lockari-backend-app/internal/core/entity/audit"
	repo_audit "github.com/synera-br/lockari-backend-app/internal/core/repository/audit"
	svc_audit "github.com/synera-br/lockari-backend-app/internal/core/service/audit"
	webhandler_audit "github.com/synera-br/lockari-backend-app/internal/handler/web/audit"

	"github.com/synera-br/lockari-backend-app/pkg/authenticator"
	"github.com/synera-br/lockari-backend-app/pkg/authorization"
	"github.com/synera-br/lockari-backend-app/pkg/cache"
	cryptserver "github.com/synera-br/lockari-backend-app/pkg/crypt/crypt_server"
	"github.com/synera-br/lockari-backend-app/pkg/database"
	httpserver "github.com/synera-br/lockari-backend-app/pkg/http_server"
	"github.com/synera-br/lockari-backend-app/pkg/message_queue"
	"github.com/synera-br/lockari-backend-app/pkg/tokengen"
)

func main() {

	cfg, viperCfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	apiResponse, err := loadWebServer(cfg.Fields["webserver"].(map[string]interface{}))
	if err != nil {
		log.Fatal(err)
	}

	crypt, err := initializeCryptData(cfg.Fields["encrypt"])
	if err != nil {
		log.Fatal(err)
	}

	authClient, db, err := initializeFirebase(cfg.Fields["firebase"])
	if err != nil {
		log.Fatal(err)
	}

	_, err = initializeCache(cfg.Fields["cache"])
	if err != nil {
		log.Fatal(err)
	}

	mq, err := initializeMessageQueue(cfg.Fields["message_queue"].(map[string]interface{}))
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

	tokenJWT, err := initializeJWT(cfg.Fields["token"].(map[string]interface{}))
	if err != nil {
		log.Fatal(err)
	}

	authSvc, err := initializeAuth(db)
	if err != nil {
		log.Fatal(err)
	}

	authZ, err := initializeAuthorization(cfg.Fields["openfga"].(map[string]interface{}), viperCfg)
	if err != nil {
		log.Fatal(err)
	}

	authZ.AddUserToTenant(context.Background(), "tenant-id", "user-id", "role")
	authZ.SetupTenant(context.Background(), "tenant-id", "role", []authorization.PlanFeature{})

	log.Println("authorization OpenFGA...", authZ)

	signup, err := initializeSignup(db, authClient, tokenJWT, authZ)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Signup service initialized successfully")

	auditSvc, err := initializeAuditEvent(db, authClient, tokenJWT)
	if err != nil {
		log.Fatal(err)
	}

	webhandler.InitializeLoginHandler(authSvc, crypt, authClient, tokenJWT, apiResponse.RouterGroup, apiResponse.MiddlewareHeader)
	webhandler.InitializeSignupHandler(signup, crypt, authClient, tokenJWT, apiResponse.RouterGroup, apiResponse.MiddlewareHeader)
	webhandler_audit.InitializeAuditSystemEventHandler(auditSvc, crypt, authClient, tokenJWT, apiResponse.RouterGroup, apiResponse.MiddlewareHeader)

	log.Println("Starting Lockari Backend App...")

	apiResponse.Run(apiResponse.Routes)
}

func loadWebServer(fields map[string]interface{}) (*httpserver.RestAPI, error) {

	var apiConfig httpserver.RestAPIConfig
	err := mapstructure.Decode(fields, &apiConfig)
	if err != nil {
		return nil, err
	}

	api, err := httpserver.NewRestApi(apiConfig)
	if err != nil {
		return nil, err
	}
	return api, nil
}

func initializeCryptData(encryptField interface{}) (cryptserver.CryptDataInterface, error) {
	token := fmt.Sprintf("%v", encryptField)
	return cryptserver.InicializationCryptData(&token)
}

func initializeFirebase(firebaseField interface{}) (authenticator.Authenticator, database.FirebaseDBInterface, error) {
	var fConfig authenticator.FirebaseConfig

	b, err := json.Marshal(firebaseField)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal firebase config: %w", err)
	}

	if err := json.Unmarshal(b, &fConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal firebase config: %w", err)
	}

	authClient, err := authenticator.InitializeAuth(context.Background(), &fConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize auth: %w", err)
	}

	db, err := database.InitializeFirebaseDB(database.FirebaseConfig{
		ProjectID:             fConfig.ProjectID,
		APIKey:                fConfig.APIKey,
		DatabaseURL:           fConfig.DatabaseURL,
		StorageBucket:         fConfig.StorageBucket,
		AppID:                 fConfig.AppID,
		AuthDomain:            fConfig.AuthDomain,
		MessagingSenderID:     fConfig.MessagingSenderID,
		ServiceAccountKeyPath: fConfig.ServiceAccountKeyPath,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize firebase DB: %w", err)
	}

	return authClient, db, nil
}

func initializeCache(fields interface{}) (cache.CacheService, error) {

	b, _ := json.Marshal(fields)

	var config cache.CacheConfig
	err := json.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}

	cacheClient, err := cache.NewRedisCacheService(config)
	if err != nil {
		return nil, err
	}

	return cacheClient, nil

}

func initializeMessageQueue(fields map[string]interface{}) (message_queue.MessageQueue, error) {

	b, _ := json.Marshal(fields)

	var config message_queue.Config
	err := json.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}

	mq, err := message_queue.NewRabbitMQ(config)
	if err != nil {
		return nil, err
	}

	err = mq.Setup()
	if err != nil {
		return nil, err
	}

	return mq, nil
}

func initializeJWT(fields map[string]interface{}) (tokengen.TokenGenerator, error) {

	token := tokengen.NewTokenGenerator(
		fields["secret"].(string),
		fields["issuer"].(string),
		time.Duration(time.Hour*4),
	)

	if token == nil {
		return nil, fmt.Errorf("failed to initialize token generator: token secret or issuer is empty")
	}

	return token, nil
}

func initializeAuth(db database.FirebaseDBInterface) (entity.LoginEventService, error) {

	repo, err := repo_auth.InitializeLoginEventRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize login event repository: %w", err)
	}

	svc, err := svc_auth.InitializeLoginEventService(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize login event service: %w", err)
	}

	return svc, nil
}

func initializeAuthorization(config map[string]interface{}, v *viper.Viper) (authorization.LockariAuthorizationService, error) {
	fmt.Println("Initializing OpenFGA client with configuration...")
	cfg, err := authorization.LoadFromViper(v)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenFGA configuration: %w", err)
	}

	if cfg == nil {
		return nil, fmt.Errorf("OpenFGA configuration is nil")
	}

	if cfg.Validate() != nil {
		return nil, fmt.Errorf("invalid OpenFGA configuration: %w", cfg.Validate())
	}

	logger := authorization.NewSlogAdapter(slog.Default())
	logger.Info("Initializing OpenFGA client with configuration", "config", cfg)

	opts := authorization.ClientOptions{
		Config: cfg,
		Logger: logger,
		Cache: &authorization.CacheOptions{
			CleanupInterval: cfg.CacheCleanupInterval,
			MaxSize:         500,
		}}

	client, err := authorization.NewOpenFGAClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenFGA client: %w", err)
	}

	// 4. Criar serviços auxiliares
	cacheService := authorization.NewMemoryCache(authorization.CacheOptions{
		CleanupInterval: cfg.CacheCleanupInterval,
		MaxSize:         500,
	})

	// 5. Criar serviço básico
	service := authorization.NewService(authorization.ServiceOptions{
		Client: client,
		Cache:  cacheService,
		Logger: logger,
	})

	// 6. Criar serviço Lockari (implementação da interface)

	lockariService := authorization.NewLockariAuthorizationService(authorization.LockariServiceOptions{
		Service: service,
		Config:  cfg,
		Logger:  logger,
	})

	if lockariService == nil {
		return nil, fmt.Errorf("failed to create LockariAuthorizationService")
	}

	return lockariService, nil
}

func initializeSignup(db database.FirebaseDBInterface, auth authenticator.Authenticator, tokenJWT tokengen.TokenGenerator, authZ authorization.LockariAuthorizationService) (entity_auth.SignupEventService, error) {

	authorization.NewLockariService(authorization.LockariServiceOptions{
		Service: nil,
		Config:  nil,
		Logger:  nil,
	})
	repo, err := repo_auth.InitializeSignupEventRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signup event repository: %w", err)
	}

	svc, err := svc_auth.InitializeSignupEventService(repo, auth, tokenJWT, authZ)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signup event service: %w", err)
	}

	return svc, nil
}

func initializeAuditEvent(db database.FirebaseDBInterface, auth authenticator.Authenticator, tokenJWT tokengen.TokenGenerator) (entity_audit.AuditSystemEventService, error) {
	repo, err := repo_audit.InicializeAuditSystemEventRepository(db)
	if err != nil {
		return nil, err
	}

	svc, err := svc_audit.InitializeAuditSystemEventService(repo, auth, tokenJWT)
	if err != nil {
		return nil, err
	}

	return svc, nil

}
