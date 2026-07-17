package application

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/engine"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/models"
)

var tracer = otel.Tracer("medhen.com/platform/pc-rating-calc-svc/application")

// AuditEventProducer defines the outbound port to send telemetry to Kafka
type AuditEventProducer interface {
	PublishRatingEvent(ctx context.Context, breakdown *models.PremiumBreakdown, req models.RatingRequest) error
}

// RatingApplicationService is the orchestration layer boundary
type RatingApplicationService struct {
	engine   *engine.RatingEngine
	producer AuditEventProducer
}

// NewRatingApplicationService creates a new application service
func NewRatingApplicationService(provider engine.RateTableProvider, producer AuditEventProducer) *RatingApplicationService {
	return &RatingApplicationService{
		engine:   engine.NewRatingEngine(provider),
		producer: producer,
	}
}

// CalculatePremium handles a rating request, invokes the domain, and fires async audit logs
func (s *RatingApplicationService) CalculatePremium(ctx context.Context, req models.RatingRequest) (*models.PremiumBreakdown, error) {
	ctx, span := tracer.Start(ctx, "CalculatePremium")
	defer span.End()

	// Execute the pure domain math pipeline
	start := time.Now()
	breakdown, err := s.engine.CalculatePremium(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("domain execution failed: %w", err)
	}

	// Add execution time to trace
	execTime := time.Since(start).Milliseconds()
	breakdown.AddTrace("PIPELINE_EXEC_MS", fmt.Sprintf("%d", execTime), "SYSTEM", span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String())

	// Fire async audit event (do not block the hot path response)
	// In a real system, we might use a background worker or goroutine if the producer isn't inherently async
	go func() {
		// Pass a background context since the incoming request context will be cancelled when response returns
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.producer.PublishRatingEvent(bgCtx, breakdown, req)
	}()

	return breakdown, nil
}

// CalculateProRata handles a midterm adjustment / cancellation rating request
func (s *RatingApplicationService) CalculateProRata(ctx context.Context, req models.RatingRequest, daysActive, totalDays int64) (*models.PremiumBreakdown, error) {
	ctx, span := tracer.Start(ctx, "CalculateProRata")
	defer span.End()

	start := time.Now()
	breakdown, err := s.engine.CalculateProRata(ctx, req, daysActive, totalDays)
	if err != nil {
		return nil, fmt.Errorf("pro-rata domain execution failed: %w", err)
	}

	execTime := time.Since(start).Milliseconds()
	breakdown.AddTrace("PIPELINE_EXEC_MS", fmt.Sprintf("%d", execTime), "SYSTEM", span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String())

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.producer.PublishRatingEvent(bgCtx, breakdown, req)
	}()

	return breakdown, nil
}
