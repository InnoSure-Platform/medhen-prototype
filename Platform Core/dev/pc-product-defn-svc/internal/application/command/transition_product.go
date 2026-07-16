package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-product-defn-svc/internal/domain/product"
	"github.com/medhen/pc-product-defn-svc/internal/infrastructure/kafka"
)

type TransitionProductCommand struct {
	TenantID               string
	ProductID              uuid.UUID
	TargetStatus           product.Status
	HasFairValueAssessment bool
}

type TransitionProductHandler struct {
	db          *pgxpool.Pool
	productRepo product.Repository
	outboxPub   *kafka.OutboxPublisher
}

func NewTransitionProductHandler(db *pgxpool.Pool, repo product.Repository, outbox *kafka.OutboxPublisher) *TransitionProductHandler {
	return &TransitionProductHandler{
		db:          db,
		productRepo: repo,
		outboxPub:   outbox,
	}
}

func (h *TransitionProductHandler) Handle(ctx context.Context, cmd TransitionProductCommand) error {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// In a real UoW we'd pass Tx down, simplifying here
	p, err := h.productRepo.GetByID(ctx, cmd.TenantID, cmd.ProductID)
	if err != nil {
		return err
	}

	if err := p.TransitionTo(cmd.TargetStatus, cmd.HasFairValueAssessment); err != nil {
		return err
	}

	var action product.LifecycleAction
	switch cmd.TargetStatus {
	case product.StatusReview:
		action = product.ActionSubmitted
	case product.StatusApproved:
		action = product.ActionApproved
	case product.StatusActive:
		action = product.ActionActivated
	case product.StatusSuspended:
		action = product.ActionSuspended
	case product.StatusRetired:
		action = product.ActionRetired
	default:
		action = product.ActionSubmitted
	}

	event := product.NewProductLifecycleEvent(p, action)

	if err := h.productRepo.Save(ctx, p); err != nil {
		return err
	}

	if err := h.outboxPub.Publish(ctx, tx, event); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
