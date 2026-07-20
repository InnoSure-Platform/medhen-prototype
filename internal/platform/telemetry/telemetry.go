// Package telemetry wires OpenTelemetry tracing for the monolith. Instrumentation
// (the HTTP edge via otelhttp, spans in the event bus) always calls the global
// tracer provider; Setup decides whether spans are exported. This keeps tracing
// zero-config in dev and a single env var away in production.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Setup installs a global tracer provider and W3C propagators. When endpoint is
// non-empty, spans are batch-exported via OTLP/HTTP to that collector; otherwise
// a provider with no exporter is installed so instrumentation is always safe to
// call but nothing is shipped. The returned func flushes and shuts down the
// provider at process exit.
func Setup(ctx context.Context, serviceName, serviceVersion, endpoint string) (func(context.Context) error, error) {
	res, err := resource.New(ctx, resource.WithAttributes(
		attribute.String("service.name", serviceName),
		attribute.String("service.version", serviceVersion),
	))
	if err != nil {
		return nil, fmt.Errorf("telemetry: resource: %w", err)
	}

	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if endpoint != "" {
		exp, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("telemetry: otlp exporter: %w", err)
		}
		opts = append(opts, sdktrace.WithBatcher(exp))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp.Shutdown, nil
}
