package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/medhen/pc-billing-svc/internal/domain/aggregate"
)

type LedgerRepository struct{}

func NewLedgerRepository() *LedgerRepository {
	return &LedgerRepository{}
}

func (r *LedgerRepository) Save(ctx context.Context, ledger *aggregate.LedgerTransaction) error {
	tx := ExtractTx(ctx)
	if tx == nil {
		return errors.New("no active transaction")
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO ledger_transactions (id, tenant_id, reference_id, reference_type, posted_at)
		VALUES ($1, $2, $3, $4, $5)
	`, ledger.ID, ledger.TenantID, ledger.ReferenceID, ledger.ReferenceType, ledger.PostedAt)
	if err != nil {
		return err
	}

	// Batch insert journal entries
	batch := &pgx.Batch{}
	for _, entry := range ledger.Entries {
		batch.Queue(`
			INSERT INTO journal_entries (ledger_transaction_id, account_code, debit_amount, credit_amount)
			VALUES ($1, $2, $3, $4)
		`, ledger.ID, entry.AccountCode, entry.DebitAmount, entry.CreditAmount)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(ledger.Entries); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
