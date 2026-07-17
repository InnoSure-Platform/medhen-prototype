package port

import (
	"context"

	"github.com/google/uuid"
)

type EnrichmentData map[string]interface{}

type EnrichmentProvider interface {
	FetchPriorClaims(ctx context.Context, tenantID string, quoteID uuid.UUID) (EnrichmentData, error)
	FetchCreditScore(ctx context.Context, tenantID string, quoteID uuid.UUID) (EnrichmentData, error)
}
