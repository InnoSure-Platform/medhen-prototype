package saga

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-party-mgmt-svc/internal/application/command"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type KYCRetrySaga struct {
	db    *pgxpool.Pool
	uow   command.UnitOfWork
	fayda command.FaydaClient
}

func NewKYCRetrySaga(db *pgxpool.Pool, uow command.UnitOfWork, fayda command.FaydaClient) *KYCRetrySaga {
	return &KYCRetrySaga{
		db:    db,
		uow:   uow,
		fayda: fayda,
	}
}

func (s *KYCRetrySaga) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down KYC Retry Saga")
			return
		case <-ticker.C:
			s.processPendingKYC(ctx)
		}
	}
}

func (s *KYCRetrySaga) processPendingKYC(ctx context.Context) {
	// 1. Find parties with PENDING KYC
	// In a real saga, we would lock these rows or use a saga orchestrator table.
	query := `
		SELECT id, national_id_number
		FROM parties
		WHERE kyc_status = $1
		LIMIT 10
	`
	rows, err := s.db.Query(ctx, query, string(domain.KYCStatusPending))
	if err != nil {
		slog.Error("Failed to query pending KYC parties", "error", err)
		return
	}
	defer rows.Close()

	var pendingParties []struct {
		ID         string
		NationalID string
	}

	for rows.Next() {
		var p struct {
			ID         string
			NationalID string
		}
		if err := rows.Scan(&p.ID, &p.NationalID); err != nil {
			slog.Error("Failed to scan pending party row", "error", err)
			continue
		}
		pendingParties = append(pendingParties, p)
	}
	
	// Close rows explicitly so we can execute new queries in UoW
	rows.Close()

	// 2. Attempt retry
	for _, p := range pendingParties {
		if p.NationalID == "" {
			continue // No ID to verify against
		}

		// Call circuit breaker client
		verified, err := s.fayda.VerifyIdentity(ctx, p.NationalID)
		if err != nil || !verified {
			// Either Fayda is still down, or returned false.
			continue
		}

		// 3. Update status transactionally with Unit of Work
		err = s.uow.Do(ctx, func(ctx context.Context, repo command.PartyRepository, outbox command.OutboxPublisher) error {
			// A true CQRS implementation would send a "VerifyKYCCommand" through the handler.
			// For this Saga, we bypass it for direct aggregate manipulation.
			updateQuery := `UPDATE parties SET kyc_status = $1, version = version + 1, updated_at = now() WHERE id = $2`
			_, err := s.db.Exec(ctx, updateQuery, string(domain.KYCStatusVerified), p.ID)
			if err != nil {
				return err
			}
			
			// We would normally publish a PartyKYCVerifiedEvent here via the outbox publisher
			slog.Info("Saga successfully verified pending KYC", "party_id", p.ID)
			return nil
		})

		if err != nil {
			slog.Error("Failed to update KYC status in Saga", "party_id", p.ID, "error", err)
		}
	}
}
