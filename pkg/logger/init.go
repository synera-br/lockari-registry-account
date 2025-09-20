package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// LoggerInterface defines the logging methods.
type LoggerInterface interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) LoggerInterface
}

// Logger is a wrapper around slog.Logger.
type Logger struct {
	slog *slog.Logger
}

// Config holds the logger configuration.
type Config struct {
	Level          string `json:"level" yaml:"level"`  // "debug", "info", "warn", "error"
	Output         string `json:"output" yaml:"output"` // "stdout", "opentelemetry"
	ServiceName    string `json:"service_name" yaml:"service_name"`
	ServiceVersion string `json:"service_version" yaml:"service_version"`
}

// New creates a new logger instance.
func NewLogger() *Logger {
	return &Logger{}
}

// Initialize configures the logger based on the provided configuration.
func (l *Logger) Initialize(cfg Config) error {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo // Default to info level
	}

	var handler slog.Handler
	switch cfg.Output {
	case "stdout":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	case "opentelemetry":
		// Initialize OpenTelemetry trace provider
		tp, err := newTraceProvider(cfg.ServiceName, cfg.ServiceVersion)
		if err != nil {
			return fmt.Errorf("failed to create trace provider: %w", err)
		}
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}) // You might want to use a different handler for OpenTelemetry
	default:
		if level == slog.LevelDebug {
			handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level, AddSource: true}) // Default to stdout
		} else {
			handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		}
	}

	l.slog = slog.New(handler)
	slog.SetDefault(l.slog) // Set the default logger
	return nil
}

// newTraceProvider creates a new trace provider for OpenTelemetry.
func newTraceProvider(serviceName, serviceVersion string) (*sdktrace.TracerProvider, error) {
	// Create stdout exporter to be able to inspect the generated spans.
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(io.Discard))
	if err != nil {
		return nil, fmt.Errorf("creating stdout exporter: %w", err)
	}

	// Ensure that the service name and version are set in the resource.
	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
	)
	return tp, nil
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// With returns a new logger with the given attributes.
func (l *Logger) With(args ...any) LoggerInterface {
	return &Logger{slog: l.slog.With(args...)}
}
