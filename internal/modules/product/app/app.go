// Package app holds the product use cases and its repository port.
package app

import (
	"context"
	"errors"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/shopspring/decimal"
)

// ErrNotFound is returned when a product does not exist.
var ErrNotFound = errors.New("product: not found")

// Repository persists and loads products, and answers the pricing lookups the
// rating engine needs.
type Repository interface {
	Upsert(ctx context.Context, p *domain.Product) error
	Get(ctx context.Context, code string) (*domain.Product, error)
	List(ctx context.Context) ([]*domain.Product, error)
	BaseRate(ctx context.Context, productCode, coverageCode string) (money.Amount, string, error)
	Factor(ctx context.Context, productCode, coverageCode, factorType, dimension string) (decimal.Decimal, string, error)
}

// Service implements the product use cases.
type Service struct {
	repo Repository
}

// NewService builds the service.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Seed upserts a product definition (idempotent), used to bootstrap the catalog.
func (s *Service) Seed(ctx context.Context, p *domain.Product) error {
	return s.repo.Upsert(ctx, p)
}

// GetProduct returns the public view of a product.
func (s *Service) GetProduct(ctx context.Context, code string) (ports.ProductView, error) {
	p, err := s.repo.Get(ctx, code)
	if err != nil {
		return ports.ProductView{}, err
	}
	return toView(p), nil
}

// ListProducts returns all products as public views.
func (s *Service) ListProducts(ctx context.Context) ([]ports.ProductView, error) {
	ps, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ports.ProductView, 0, len(ps))
	for _, p := range ps {
		out = append(out, toView(p))
	}
	return out, nil
}

// Repo exposes the repository to the module's rate-provider adapter.
func (s *Service) Repo() Repository { return s.repo }

func toView(p *domain.Product) ports.ProductView {
	covs := make([]ports.CoverageView, 0, len(p.Coverages))
	for _, c := range p.Coverages {
		covs = append(covs, ports.CoverageView{
			Code: c.Code, Name: c.Name, NameAmharic: c.NameAmharic, BaseRate: c.BaseRate,
		})
	}
	return ports.ProductView{
		Code: p.Code, LOB: p.LOB, Name: p.Name, NameAmharic: p.NameAmharic,
		Status: string(p.Status), Coverages: covs,
	}
}
