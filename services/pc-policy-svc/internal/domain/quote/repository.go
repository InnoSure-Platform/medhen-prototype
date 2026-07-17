package quote

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, q *Quote) error
	GetByID(ctx context.Context, id uuid.UUID) (*Quote, error)
}
