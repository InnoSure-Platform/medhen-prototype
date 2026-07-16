package query

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Customer360View represents the flattened CQRS read projection 
// aggregating Party, Policies, Claims, and Billing data.
type Customer360View struct {
	PartyID           uuid.UUID   `json:"party_id"`
	FullName          string      `json:"full_name"`
	KYCStatus         string      `json:"kyc_status"`
	RiskScore         float64     `json:"risk_score"`          // Sourced from pc-fincrime-svc
	ActivePolicies    int         `json:"active_policies"`     // Sourced from pc-policy-svc events
	TotalClaimsValue  float64     `json:"total_claims_value"`  // Sourced from pc-claims-svc events
	OutstandingBalance float64     `json:"outstanding_balance"` // Sourced from pc-billing-svc events
}

type Customer360QueryService struct {
	db *pgxpool.Pool
}

func NewCustomer360QueryService(db *pgxpool.Pool) *Customer360QueryService {
	return &Customer360QueryService{db: db}
}

func (s *Customer360QueryService) GetCustomer360(ctx context.Context, tenantID string, partyID uuid.UUID) (*Customer360View, error) {
	// In a real CQRS system, this queries a flat NoSQL table or a JSONB projection table
	// that is continuously updated by a Kafka consumer listening to cross-domain events.
	query := `
		SELECT 
			party_id, full_name, kyc_status, risk_score, 
			active_policies, total_claims_value, outstanding_balance
		FROM customer_360_projections
		WHERE tenant_id = $1 AND party_id = $2
	`
	
	var view Customer360View
	err := s.db.QueryRow(ctx, query, tenantID, partyID).Scan(
		&view.PartyID, &view.FullName, &view.KYCStatus, &view.RiskScore,
		&view.ActivePolicies, &view.TotalClaimsValue, &view.OutstandingBalance,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("customer 360 projection not found")
		}
		return nil, err
	}

	return &view, nil
}
