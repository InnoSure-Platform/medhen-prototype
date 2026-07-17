package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds the standard telemetry configuration for all InnoGuard services.
type Config struct {
	ServiceName string
	Version     string
	Endpoint    string // OTLP endpoint (e.g., localhost:4317)
}

// TraceHandler is a custom slog handler that adds trace context to log records.
type TraceHandler struct {
	slog.Handler
}

// Handle adds trace_id and span_id to the log record if they exist in the context.
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		r.AddAttrs(slog.String("trace_id", spanCtx.TraceID().String()))
		r.AddAttrs(slog.String("span_id", spanCtx.SpanID().String()))
	}
	return h.Handler.Handle(ctx, r)
}

// TraceLoggingMiddleware wraps an existing slog handler with the TraceHandler.
func TraceLoggingMiddleware(next slog.Handler) slog.Handler {
	return &TraceHandler{Handler: next}
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
		otlptracegrpc.WithInsecure(), // Consider securing this for actual prod deployments
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

	// Set up structured logging (slog) with trace correlation
	logger := slog.New(TraceLoggingMiddleware(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	slog.SetDefault(logger)

	return tp.Shutdown, nil
}

// Middleware provides a simple HTTP middleware to extract trace contexts
// For more robust needs, services should use go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
func Middleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := otel.Tracer(serviceName).Start(ctx, r.URL.Path)
			defer span.End()
			
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
