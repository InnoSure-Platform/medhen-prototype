package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Config holds the standard telemetry configuration for all InnoGuard services.
type Config struct {
	ServiceName string
	Version     string
	Endpoint    string // OTLP endpoint (e.g., localhost:4317)
}

// Init initializes OpenTelemetry and structured logging.
// It returns a shutdown function that should be called when the service exits.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	// Set up the OTLP trace exporter
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "localhost:4317" // Default local OTel collector
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(), // In a real Tier-0 this might be TLS-secured depending on mesh
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create resource describing the service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Set up the Trace Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Set up structured logging (slog) with a basic handler for now
	// Ideally we'd wrap this with otelslog to extract trace_id automatically.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	return tp.Shutdown, nil
}
