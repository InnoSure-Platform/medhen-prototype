package app

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"medhen.com/pc-observability-svc/internal/domain"
	"medhen.com/pc-observability-svc/internal/ports"
)

var tracer = otel.Tracer("medhen.com/pc-observability-svc/app")

// SLOService orchestrates the creation and syncing of SLOs.
type SLOService struct {
	sloRepo     ports.SLORepository
	mimirClient ports.MimirClient
}

// NewSLOService creates a new SLOService application layer.
func NewSLOService(sloRepo ports.SLORepository, mimirClient ports.MimirClient) *SLOService {
	return &SLOService{
		sloRepo:     sloRepo,
		mimirClient: mimirClient,
	}
}

// CreateSLO handles the use case of creating a new SLO and syncing it to Mimir.
func (s *SLOService) CreateSLO(ctx context.Context, tenantID, name, desc string, target float64, window int, query, policyID string) (*domain.SLO, error) {
	ctx, span := tracer.Start(ctx, "SLOService.CreateSLO", trace.WithAttributes(
		attribute.String("tenant.id", tenantID),
		attribute.String("slo.name", name),
	))
	defer span.End()

	slog.InfoContext(ctx, "Creating new SLO", "tenant_id", tenantID, "name", name)
	// In a real app, generate a UUID
	id := "slo-1234-uuid"

	slo, err := domain.NewSLO(id, tenantID, name, desc, target, window, query, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to create SLO domain object: %w", err)
	}

	// Save initial state to DB (SYNCING)
	if err := s.sloRepo.Save(ctx, slo); err != nil {
		return nil, fmt.Errorf("failed to save SLO to repository: %w", err)
	}

	// Sync to Data Plane
	if err := s.mimirClient.PushRules(ctx, slo); err != nil {
		// Log error, rely on background outbox job to retry
		return slo, nil
	}

	// If successful, mark active
	slo.MarkActive()
	if err := s.sloRepo.Save(ctx, slo); err != nil {
		return nil, fmt.Errorf("failed to update SLO status to ACTIVE: %w", err)
	}

	return slo, nil
}
