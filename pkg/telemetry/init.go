package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	tracer "go.opentelemetry.io/otel/trace"
)

type OtelTypeTracer string

const (
	OtelTypeTracerJaeger    OtelTypeTracer = "jaeger"
	OtelTypeTracerTempo     OtelTypeTracer = "tempo"
	OtelTypeTracerHoneycomb OtelTypeTracer = "honeycomb"
	OtelTypeTracerStdout    OtelTypeTracer = "stdout"
)

type OtelConfig struct {
	EndpointType   OtelTypeTracer    `json:"endpoint_type" yaml:"endpoint_type"`
	Endpoint       string            `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	ServiceName    string            `json:"service_name,omitempty" yaml:"service_name,omitempty"`
	Headers        map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	EnabledTracer  bool              `json:"enabled_tracer,omitempty" yaml:"enabled_tracer,omitempty"`
	EnabledLogger  bool              `json:"enabled_logger,omitempty" yaml:"enabled_logger,omitempty"`
	exporter       trace.SpanExporter
	traceProvider  *trace.TracerProvider
	loggerProvider *log.LoggerProvider
	logger         *slog.Logger
}

type OtelObservability interface {
	CleanUp() func()
	Tracer() (tracer.Tracer, error)
	Span(context.Context, string) (context.Context, tracer.Span)
	Logger() (*slog.Logger, error)
	Trace(ctx context.Context)tracer.Span 
}

func InitObservability(cfg *OtelConfig) (OtelObservability, func(), error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is nil")
	}

	if err := cfg.validate(); err != nil {
		return nil, nil, err
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = "default-app"
	}

	var cleanupFuncs []func()

	if cfg.EnabledTracer {
		cleanupTracer, err := cfg.setTracer()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize tracer: %w", err)
		}
		if cleanupTracer != nil {
			cleanupFuncs = append(cleanupFuncs, cleanupTracer)
		}
	}

	if cfg.EnabledLogger {
		ctx := context.Background()
		cleanupLogger, err := cfg.setLogger(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
		}
		if cleanupLogger != nil {
			cleanupFuncs = append(cleanupFuncs, cleanupLogger)
		}
	}

	// Retorna função que executa todos os cleanups
	cleanup := func() {
		for i := len(cleanupFuncs) - 1; i >= 0; i-- {
			if cleanupFuncs[i] != nil {
				cleanupFuncs[i]()
			}
		}
	}

	return cfg, cleanup, nil
}

func (c *OtelConfig) Trace(ctx context.Context)tracer.Span {
	span := tracer.SpanFromContext(ctx)
	return span
}

func (c *OtelConfig) Tracer() (tracer.Tracer, error) {
	if c.ServiceName == "" {
		c.ServiceName = "lockari-backend-app"
	}

	t := otel.Tracer(c.ServiceName)
	if t == nil {
		return nil, fmt.Errorf("tracer is nil")
	}

	return t, nil
}

func (c *OtelConfig) CleanUp() func() {
	if c.traceProvider == nil {
		fmt.Println("trace provider is nil")
		return nil
	}

	f := func() {
		err := c.traceProvider.Shutdown(context.Background())
		if err != nil {
			fmt.Println("failed to shutdown trace provider: %v", err)
		}
	}

	return f
}

func (c *OtelConfig) validate() error {
	if c.EnabledTracer || c.EnabledLogger {
		if c.Endpoint == "" {
			return fmt.Errorf("endpoint is required when tracer or logger is enabled")
		}
		if c.EndpointType == "" {
			return fmt.Errorf("endpoint type is required when tracer or logger is enabled")
		}
	}
	return nil
}

func (c *OtelConfig) setLogger(ctx context.Context) (func(), error) {
	if c.EndpointType == "" {
		c.EndpointType = OtelTypeTracerStdout
	}

	headers := make(map[string]string)
	for k, v := range c.Headers {
		headers[k] = v
	}

	logExporter, err := otlploghttp.New(
		ctx,
		otlploghttp.WithEndpoint(c.Endpoint),
		otlploghttp.WithHeaders(headers),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize log exporter: %w", err)
	}

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(c.ServiceName),
		semconv.ServiceVersionKey.String("1.0.0"),
		attribute.String("environment", "development"),
	)

	lp := log.NewLoggerProvider(
		log.WithProcessor(
			log.NewBatchProcessor(logExporter),
		),
		log.WithResource(resources),
	)

	// Store logger provider for cleanup
	c.loggerProvider = lp

	// Set the logger provider globally
	global.SetLoggerProvider(lp)

	// Create and store slog logger
	c.logger = otelslog.NewLogger(c.ServiceName)

	// Test log
	c.logger.Debug("Logger initialized successfully", "service", c.ServiceName)

	// Return cleanup function
	cleanup := func() {
		if err := lp.Shutdown(ctx); err != nil {
			if c.logger != nil {
				c.logger.Error("failed to shutdown logger provider", "error", err)
			}
		}
	}

	return cleanup, nil
}

func (c *OtelConfig) setTracer() (func(), error) {

	serviceName := "default-app"

	if c.EndpointType == "" {
		return nil, fmt.Errorf("tracer endpoint is empty")
	}

	if c.EndpointType != OtelTypeTracerJaeger &&
		c.EndpointType != OtelTypeTracerTempo &&
		c.EndpointType != OtelTypeTracerHoneycomb &&
		c.EndpointType != OtelTypeTracerStdout {
		return nil, fmt.Errorf("invalid tracer endpoint")
	}

	headers := make(map[string]string)
	for k, v := range c.Headers {
		headers[k] = v
	}

	var exporter trace.SpanExporter
	// var exporterHttp *otlploghttp.Exporter
	var err error
	switch c.EndpointType {
	case OtelTypeTracerStdout:
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
		}

	case OtelTypeTracerJaeger:
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(c.Endpoint),
		))
		if err != nil {
			return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
		}
	case OtelTypeTracerHoneycomb:
		// Para Honeycomb, headers são passados via OTLP

		headers["x-honeycomb-project-name"] = serviceName
		headers["x-honeycomb-dataset"] = serviceName

		if c.ServiceName != "" {
			serviceName = c.ServiceName
			headers["x-honeycomb-project-name"] = c.ServiceName
			headers["x-honeycomb-dataset"] = c.ServiceName
		}

		exporter, err = otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithEndpoint(c.Endpoint),
			otlptracegrpc.WithHeaders(headers))
		if err != nil {
			return nil, fmt.Errorf("failed to create honeycomb exporter: %w", err)
		}

	default:
		return nil, fmt.Errorf("invalid tracer endpoint")
	}

	if exporter == nil {
		return nil, fmt.Errorf("span exporter cannot be nil")
	}

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String("1.0.0"),
	)

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resources),
	)

	otel.SetTracerProvider(traceProvider)

	c.traceProvider = traceProvider
	f := func() {
		err := c.traceProvider.Shutdown(context.Background())
		if err != nil {
			fmt.Printf("failed to shutdown trace provider: %w", err)
		}
	}

	return f, nil
}

func (c *OtelConfig) Span(ctx context.Context, name string) (context.Context, tracer.Span) {
	if !c.EnabledTracer {
		return ctx, tracer.SpanFromContext(ctx)
	}

	t, err := c.Tracer()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("failed to get tracer for span", "error", err, "span_name", name)
		}
		return ctx, tracer.SpanFromContext(ctx)
	}

	return t.Start(ctx, name)
}

func (c *OtelConfig) Logger() (*slog.Logger, error) {
	if !c.EnabledLogger {
		return nil, fmt.Errorf("logger is not enabled")
	}

	if c.logger == nil {
		return nil, fmt.Errorf("logger not initialized")
	}

	return c.logger, nil
}
