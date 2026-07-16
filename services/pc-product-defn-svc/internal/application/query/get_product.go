package query

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/medhen/pc-auth-sdk"
	"github.com/medhen/pc-product-defn-svc/internal/domain/product"
)

var ErrProductNotFound = errors.New("product not found")

type GetProductHandler struct {
	repo product.Repository
}

func NewGetProductHandler(repo product.Repository) *GetProductHandler {
	return &GetProductHandler{
		repo: repo,
	}
}

func (h *GetProductHandler) Handle(ctx context.Context, idStr string) (*product.Product, error) {
	tenantID, err := auth.GetTenantID(ctx)
	if err != nil {
		return nil, errors.New("unauthenticated")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	p, err := h.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}

	return p, nil
}
