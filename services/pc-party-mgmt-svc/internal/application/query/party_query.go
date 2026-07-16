package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type PartyQueryService interface {
	GetPartySummary(ctx context.Context, tenantID string, partyID uuid.UUID) (*domain.Party, error)
	GetKYCStatus(ctx context.Context, tenantID string, partyID uuid.UUID) (string, error)
}
