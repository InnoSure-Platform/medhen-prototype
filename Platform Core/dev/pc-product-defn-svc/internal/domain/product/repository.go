package product

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the persistence interface for the Product Aggregate.
// It is designed to work within a Unit of Work to ensure atomic commits alongside the Outbox.
type Repository interface {
	Save(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, tenantID string, id uuid.UUID) (*Product, error)
	// Additional search/list methods could go here.
}
