package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/medhen/pc-underwriting-svc/internal/application/port"
)

type SLAWorker struct {
	uow          port.UnitOfWork
	referralRepo port.ReferralRepository
	outboxRepo   port.OutboxRepository
}

func NewSLAWorker(uow port.UnitOfWork, rr port.ReferralRepository, or port.OutboxRepository) *SLAWorker {
	return &SLAWorker{
		uow:          uow,
		referralRepo: rr,
		outboxRepo:   or,
	}
}

// Run starts the daemon blocking on the context.
func (w *SLAWorker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Println("SLA Worker started")
	for {
		select {
		case <-ctx.Done():
			log.Println("SLA Worker shutting down")
			return
		case t := <-ticker.C:
			w.sweep(ctx, t)
		}
	}
}

func (w *SLAWorker) sweep(ctx context.Context, now time.Time) {
	// 1. Find breached without tx to minimize lock duration
	breached, err := w.referralRepo.FindBreachedSLAs(ctx, now, 100)
	if err != nil {
		log.Printf("SLA Worker failed to fetch SLAs: %v", err)
		return
	}

	for _, ref := range breached {
		_ = w.uow.Do(ctx, func(txCtx context.Context) error {
			// Reload inside tx for concurrency safety
			txRef, err := w.referralRepo.FindByID(txCtx, ref.ID)
			if err != nil {
				return err
			}

			if txRef.CheckSLA() { // returns true if it transitioned to ESCALATED
				if err := w.referralRepo.Update(txCtx, txRef); err != nil {
					return err
				}

				payload, _ := json.Marshal(map[string]interface{}{
					"referral_id":     txRef.ID,
					"new_status":      txRef.Status,
					"escalation_time": now,
				})
				_ = w.outboxRepo.PublishEvent(txCtx, "pc.underwriting.referral.sla_breached.v1", txRef.ID.String(), payload)
				log.Printf("Escalated referral %s due to SLA breach", txRef.ID)
			}
			return nil
		})
	}
}
