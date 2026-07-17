package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"medhen/pc-claims-svc/internal/domain"
)

type ReserveRepository struct {
	pool *pgxpool.Pool
}

func NewReserveRepository(pool *pgxpool.Pool) *ReserveRepository {
	return &ReserveRepository{pool: pool}
}

// SaveLedger implements Bi-Temporal inserts for IFRS-17 auditability
func (r *ReserveRepository) SaveLedger(ctx context.Context, ledger *domain.ReserveLedger, eventPayload []byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// In a real implementation, we only insert the newly appended entries.
	// For bi-temporal tracking, if we were updating a record, we would expire the old one:
	// UPDATE reserve_ledger SET sys_valid_to = NOW() WHERE id = X;
	// Since ReserveLedger entries are append-only immutables, we just insert.
	
	lastEntry := ledger.Entries[len(ledger.Entries)-1]

	_, err = tx.Exec(ctx, `
		INSERT INTO reserve_ledger (
			claim_id, reserve_type, transaction_type, amount, currency, exchange_rate, 
			base_amount_delta, running_balance_base, reason_code, author_id, 
			sys_valid_from, sys_valid_to, biz_valid_from, biz_valid_to
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, 
		lastEntry.ClaimID, lastEntry.ReserveType, lastEntry.TransactionType, 
		lastEntry.Amount.Amount.String(), lastEntry.Amount.Currency, lastEntry.Amount.ExchangeRate.String(), 
		lastEntry.Amount.BaseAmount.String(), lastEntry.RunningBalanceBase.String(), 
		lastEntry.ReasonCode, lastEntry.AuthorID,
		lastEntry.SystemValidFrom, lastEntry.SystemValidTo, lastEntry.BusinessValidFrom, lastEntry.BusinessValidTo,
	)

	if err != nil {
		return fmt.Errorf("failed to insert bi-temporal reserve entry: %w", err)
	}

	// Insert Outbox
	partitionKey := fmt.Sprintf("reserve:%s", lastEntry.ClaimID)
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox (topic, partition_key, payload, headers)
		VALUES ($1, $2, $3, $4)
	`, "pc.claim.financials.v1", partitionKey, eventPayload, "{}")
	if err != nil {
		return fmt.Errorf("failed to insert financial outbox: %w", err)
	}

	return tx.Commit(ctx)
}
