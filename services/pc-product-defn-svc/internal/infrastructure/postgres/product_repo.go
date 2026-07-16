package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-product-defn-svc/internal/domain/product"
)

// ProductRepository implements product.Repository using pgx.
type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

// Save persists the product. It should be executed within a Unit of Work transaction.
func (r *ProductRepository) Save(ctx context.Context, p *product.Product) error {
	// Example impl using a theoretical Tx from context, or falling back to Pool.
	// We will implement UoW extraction later. For now, this is a stub showing the query.
	query := `
		INSERT INTO products (id, tenant_id, code, lob, name, status, version, effective_from, effective_to, require_fair_value, schema_payload, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			version = EXCLUDED.version,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.TenantID, p.Code, p.LOB, p.Name, p.Status, p.Version,
		p.EffectiveFrom, p.EffectiveTo, p.RequireFairValue, p.SchemaPayload,
		p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, tenantID string, id uuid.UUID) (*product.Product, error) {
	// Stub implementation to be expanded
	return nil, nil
}
