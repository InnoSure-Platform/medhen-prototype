package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-underwriting-svc/internal/application/port"
	"github.com/medhen/pc-underwriting-svc/internal/domain/aggregate"
)

type AssessmentRepo struct {
	db *pgxpool.Pool
}

func NewAssessmentRepo(db *pgxpool.Pool) port.AssessmentRepository {
	return &AssessmentRepo{db: db}
}

func (r *AssessmentRepo) Save(ctx context.Context, assessment *aggregate.UnderwritingAssessment) error {
	// In a real implementation, this would use the pgx.Tx from context if available
	// For brevity, assuming simple insert
	query := `INSERT INTO assessments (id, tenant_id, quote_id, product_id, status, risk_score, rules_triggered, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, assessment.ID, assessment.TenantID, assessment.QuoteID, assessment.ProductID,
		assessment.Status, assessment.RiskScore.Value, assessment.RulesTriggered, assessment.CreatedAt)
	return err
}

func (r *AssessmentRepo) FindByID(ctx context.Context, id uuid.UUID) (*aggregate.UnderwritingAssessment, error) {
	// Mock implementation
	return nil, nil
}

func (r *AssessmentRepo) FindByQuoteID(ctx context.Context, tenantID string, quoteID uuid.UUID) (*aggregate.UnderwritingAssessment, error) {
	// Mock implementation
	return nil, nil
}
