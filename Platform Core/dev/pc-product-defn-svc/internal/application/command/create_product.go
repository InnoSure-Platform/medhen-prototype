package command

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-product-defn-svc/internal/domain/product"
	"github.com/medhen/pc-product-defn-svc/internal/infrastructure/kafka"
)

type CreateProductCommand struct {
	TenantID         string
	Code             string
	LOB              string
	Name             string
	RequireFairValue bool
}

type CreateProductHandler struct {
	db          *pgxpool.Pool
	productRepo product.Repository
	outboxPub   *kafka.OutboxPublisher
}

func NewCreateProductHandler(db *pgxpool.Pool, repo product.Repository, outbox *kafka.OutboxPublisher) *CreateProductHandler {
	return &CreateProductHandler{
		db:          db,
		productRepo: repo,
		outboxPub:   outbox,
	}
}

func (h *CreateProductHandler) Handle(ctx context.Context, cmd CreateProductCommand) (*product.Product, error) {
	p := product.NewProduct(cmd.TenantID, cmd.Code, cmd.LOB, cmd.Name, cmd.RequireFairValue)

	// Create the event
	event := product.NewProductLifecycleEvent(p, product.ActionCreated)

	// Unit of Work: Start transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Save Aggregate
	if err := h.productRepo.Save(ctx, p); err != nil {
		return nil, err
	}

	// 2. Publish to Outbox
	if err := h.outboxPub.Publish(ctx, tx, event); err != nil {
		return nil, err
	}

	// 3. Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return p, nil
}
