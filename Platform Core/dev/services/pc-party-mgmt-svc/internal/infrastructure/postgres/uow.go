package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-party-mgmt-svc/internal/application/command"
)

type UnitOfWork struct {
	db *pgxpool.Pool
}

func NewUnitOfWork(db *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// Do executes the given function within a database transaction.
func (u *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, repo command.PartyRepository, outbox command.OutboxPublisher) error) error {
	tx, err := u.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// We create transactional versions of our repositories.
	// In pgx, a Tx shares the same interface for Exec/Query as a Pool if wrapped, 
	// or we just inject the Tx into specifically created tx-bound repos.
	txRepo := &PartyTxRepository{tx: tx}
	txOutbox := &OutboxTxRepository{tx: tx}

	err = fn(ctx, txRepo, txOutbox)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
