package webserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"registry-account/pkg/telemetry"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Configuração para o servidor Gin.
type WebServerParams struct {
	Port            string `json:"port" yaml:"port"`
	ServiceName     string `json:"service_name" yaml:"service_name"`
	ServiceVersion  string `json:"service_version" yaml:"service_version"`
	UserAuthHeader  string `json:"user_authentication_header" yaml:"user_authentication_header"`
	AppAuthHeader   string `json:"app_authentication_header" yaml:"app_authentication_header"`
	SwaggerPath     string `json:"swagger_path" yaml:"swagger_path"`
	MetricsPort     string `json:"metrics_port" yaml:"metrics_port"`
	HealthCheckPath string `json:"health_check_path" yaml:"health_check_path"`
	MetricsPath     string `json:"metrics_path" yaml:"metrics_path"`
	Environment     string `json:"environment" yaml:"environment"`
}

type Config struct {
	WebServerParams
	Telemetry         telemetry.OtelConfig
	Tracer            telemetry.OtelObservability
	TracerCleanupFunc func()
}

func (c *WebServerParams) Validate() error {
	if c.Port == "" {
		c.Port = "8080"
	}
	if c.ServiceName == "" {
		c.ServiceName = "webserver"
	}
	if c.ServiceVersion == "" {
		c.ServiceVersion = "0.1.0"
	}
	if c.UserAuthHeader == "" {
		c.UserAuthHeader = "Authorization"
	}
	if c.AppAuthHeader == "" {
		c.AppAuthHeader = "Authorization"
	}
	if c.SwaggerPath == "" {
		c.SwaggerPath = "/docs"
	}
	if c.HealthCheckPath == "" {
		c.HealthCheckPath = "/healthz"
	}
	if c.MetricsPath == "" {
		c.MetricsPath = "/metrics"
	}
	if c.Environment == "" {
		c.Environment = "development"
	}
	return nil
}

// Server representa a instância do servidor Gin.
type Server struct {
	Engine            *gin.Engine
	Config            Config
	Tracer            telemetry.OtelObservability
	TracerCleanupFunc func()
	Logger            *zap.Logger
	userAuthFunc      gin.HandlerFunc
	appAuthFunc       gin.HandlerFunc
	customHandlers    map[string]gin.HandlerFunc
	httpServer        *http.Server
}

// ServerInterface define os métodos que o servidor deve implementar.
type ServerInterface interface {
	Initialize(userAuthFunc, appAuthFunc gin.HandlerFunc) error
	Start() error
	RegisterRoute(method, path string, handlers ...gin.HandlerFunc)
	RegisterCustomHandler(name string, handler gin.HandlerFunc)
	GetCustomHandler(name string) gin.HandlerFunc
	Run() error
	Shutdown(ctx context.Context) error
}

// NewServer cria uma nova instância do servidor Gin.
func NewServer(cfg *Config) (ServerInterface, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	srv := &Server{
		Engine:         gin.Default(),
		customHandlers: make(map[string]gin.HandlerFunc),
		Config:         *cfg,
	}

	srv.Tracer = cfg.Tracer
	srv.TracerCleanupFunc = cfg.TracerCleanupFunc
	if cfg.Tracer == nil {
		obs, cleanup, err := telemetry.InitObservability(&cfg.Telemetry)
		if err != nil {
			return nil, fmt.Errorf("failed to setup observability: %w", err)
		}
		srv.Tracer = obs
		srv.TracerCleanupFunc = cleanup
	}

	return srv, nil
}

func (s *Server) Shutdown(ctx context.Context) error {

	if s.httpServer == nil {
		return fmt.Errorf("server not started")
	}

	s.Logger.Info("Shutting down server...")

	// Shutdown graceful do servidor HTTP
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.Logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	// Cleanup do OpenTelemetry
	if s.TracerCleanupFunc != nil {
		s.TracerCleanupFunc()
	}

	s.Logger.Info("Server shutdown completed")
	return nil
}

func (s *Server) Run() error {

	s.Engine.Use(otelgin.Middleware(fmt.Sprintf("%s-auto", s.Config.ServiceName)))
	s.Engine.Use(cors.Default())
	if err := s.Engine.Run(); err != nil {
		return err
	}

	return nil
}

// Initialize inicializa o servidor Gin com as configurações fornecidas.
func (s *Server) Initialize(userAuthFunc, appAuthFunc gin.HandlerFunc) error {

	// Configura o logger
	logger, err := setupLogger(s.Config.Environment)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	s.Logger = logger

	// Define as funções de autenticação
	s.userAuthFunc = userAuthFunc
	s.appAuthFunc = appAuthFunc

	// Configura as rotas padrão
	s.setupRoutes()

	// Registra a função de shutdown para ser executada no encerramento do servidor
	s.RegisterCustomHandler("shutdown_otel", func(c *gin.Context) {
		if s.Tracer != nil {
			if s.TracerCleanupFunc != nil {
				s.TracerCleanupFunc()
			}
		}

		c.String(http.StatusOK, "OpenTelemetry shutdown completed")
	})

	return nil
}

// setupLogger configura o logger com base no ambiente.
func setupLogger(environment string) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if environment == "production" {
		config := zap.NewProductionConfig()
		config.DisableStacktrace = true
		logger, err = config.Build()
	} else {
		config := zap.NewDevelopmentConfig()
		config.DisableStacktrace = true
		logger, err = config.Build()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return logger, nil
}

func (s *Server) setupTracer(ctx context.Context, spanName string) (context.Context, trace.Span) {
	tracer, err := s.Tracer.Tracer()
	if err != nil {
		s.Logger.Error("failed to get tracer", zap.Error(err))
		return ctx, trace.SpanFromContext(ctx)
	}
	return tracer.Start(ctx, spanName)
}

// setupRoutes configura as rotas padrão do servidor.
func (s *Server) setupRoutes() {
	// Health check
	s.Engine.GET("/trace-example", func(c *gin.Context) {
		_, span := s.setupTracer(c.Request.Context(), "trace-example-span")
		defer span.End()

		// Adicione atributos ou eventos ao span se desejar
		span.SetAttributes(semconv.ServiceNameKey.String(s.Config.ServiceName))

		// Simule alguma lógica
		c.String(http.StatusOK, "Span criado! TraceID: %s", span.SpanContext().TraceID().String())
	})
	// Metrics
	s.Engine.GET(s.Config.MetricsPath, gin.WrapH(promhttp.Handler()))

	// Swagger
	s.Engine.GET(s.Config.SwaggerPath+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// Start inicia o servidor Gin.
func (s *Server) Start() error {
	// Salva na struct para uso no Shutdown
	s.httpServer = &http.Server{
		Addr:    ":" + s.Config.Port,
		Handler: s.Engine,
	}

	// Apenas inicia em goroutine - NÃO bloqueia
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.Error("Server failed", zap.Error(err))
		}
	}()

	s.Logger.Info("Server started", zap.String("port", s.Config.Port))
	return nil
}

// RegisterRoute registra uma nova rota no servidor Gin.
func (s *Server) RegisterRoute(method, path string, handlers ...gin.HandlerFunc) {
	if s == nil {
		return
	}

	// Adiciona os handlers de autenticação, se fornecidos
	var allHandlers []gin.HandlerFunc
	if s.userAuthFunc != nil {
		allHandlers = append(allHandlers, s.userAuthFunc)
	}
	if s.appAuthFunc != nil {
		allHandlers = append(allHandlers, s.appAuthFunc)
	}

	// Adiciona os handlers da rota
	allHandlers = append(allHandlers, handlers...)

	switch method {
	case http.MethodGet:
		s.Engine.GET(path, allHandlers...)
	case http.MethodPost:
		s.Engine.POST(path, allHandlers...)
	case http.MethodPut:
		s.Engine.PUT(path, allHandlers...)
	case http.MethodDelete:
		s.Engine.DELETE(path, allHandlers...)
	case http.MethodPatch:
		s.Engine.PATCH(path, allHandlers...)
	case http.MethodOptions:
		s.Engine.OPTIONS(path, allHandlers...)
	case http.MethodHead:
		s.Engine.HEAD(path, allHandlers...)
	default:
		panic("Unsupported HTTP method: " + method)
	}
}

// RegisterCustomHandler registra um handler customizado.
func (s *Server) RegisterCustomHandler(name string, handler gin.HandlerFunc) {
	s.customHandlers[name] = handler
}

// GetCustomHandler retorna um handler customizado.
func (s *Server) GetCustomHandler(name string) gin.HandlerFunc {
	return s.customHandlers[name]
}

// AuthMiddleware é um exemplo de middleware de autenticação.
func AuthMiddleware(headerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(headerName)
		// if token == "" {
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token Unauthorized"})
		// 	return
		// }

		// TODO: Validar o token (ex: verificar no banco de dados, etc.)
		log.Printf("Token recebido: %s", token)

		c.Next()
	}
}
