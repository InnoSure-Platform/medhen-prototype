package sagas

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"medhen/pc-claims-svc/internal/domain"
)

// SettlementSaga orchestrates the distributed payout
type SettlementSaga struct {
	pool *pgxpool.Pool
}

func NewSettlementSaga(pool *pgxpool.Pool) *SettlementSaga {
	return &SettlementSaga{pool: pool}
}

// Execute kicks off the Outbox commands for approved settlements
func (s *SettlementSaga) Execute(ctx context.Context, settlement *domain.Settlement) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update Settlement Status
	_, err = tx.Exec(ctx, `UPDATE settlements SET status = $1 WHERE id = $2`, domain.SettlementApproved, settlement.ID)
	if err != nil {
		return fmt.Errorf("failed to update settlement: %w", err)
	}

	// Emit Saga Commands to Billing
	for _, disb := range settlement.Disbursements {
		payload, _ := json.Marshal(map[string]interface{}{
			"saga_id":        settlement.ID,
			"claim_id":       settlement.ClaimID,
			"payee_id":       disb.PayeeID,
			"amount_base":    disb.Amount.String(),
			"payment_method": disb.PaymentMethod,
		})

		_, err = tx.Exec(ctx, `
			INSERT INTO outbox (topic, partition_key, payload, headers)
			VALUES ($1, $2, $3, $4)
		`, "pc.billing.disbursement.command.v1", settlement.ClaimID, payload, "{}")
		
		if err != nil {
			return fmt.Errorf("failed to insert billing outbox: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// HandleBillingFailure is the Compensating Action
func (s *SettlementSaga) HandleBillingFailure(ctx context.Context, sagaID string) error {
	// Reverses the Reserve PAYMENT_DRAWDOWN and marks Settlement as FAILED
	// In production, this would load the reserve ledger and execute ledger.AdjustReserve(INCREASE)
	return nil
}
