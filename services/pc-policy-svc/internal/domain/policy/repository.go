package policy

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, p *Policy) error
	GetByID(ctx context.Context, id uuid.UUID) (*Policy, error)
	GetVersionAt(ctx context.Context, policyID uuid.UUID, asOf time.Time) (*PolicyVersion, error)
}
