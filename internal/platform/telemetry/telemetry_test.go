package telemetry_test

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/telemetry"
)

func TestSetup_NoEndpointIsSafe(t *testing.T) {
	shutdown, err := telemetry.Setup(context.Background(), "medhen-api", "test", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected a non-nil shutdown func")
	}

	// Instrumentation must be safe to call: creating a span does not panic even
	// with no exporter configured.
	_, span := otel.Tracer("test").Start(context.Background(), "unit")
	span.End()

	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestSetup_InstallsPropagator(t *testing.T) {
	if _, err := telemetry.Setup(context.Background(), "svc", "v", ""); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// A composite text-map propagator (tracecontext + baggage) must be installed
	// so trace context flows across the HTTP edge.
	if fields := otel.GetTextMapPropagator().Fields(); len(fields) == 0 {
		t.Fatal("expected propagator fields (traceparent, baggage) to be registered")
	}
}
