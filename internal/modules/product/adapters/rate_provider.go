package adapters

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/app"
	ratingports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// RateProvider adapts the product repository to the rating module's
// RateTableProvider port. This is the in-process cross-module wiring: rating
// depends on the interface it defines, product supplies the implementation.
type RateProvider struct {
	repo app.Repository
}

// NewRateProvider builds the adapter.
func NewRateProvider(repo app.Repository) *RateProvider { return &RateProvider{repo: repo} }

var _ ratingports.RateTableProvider = (*RateProvider)(nil)

// BaseRate delegates to the product catalog.
func (p *RateProvider) BaseRate(ctx context.Context, productCode, coverageCode string, _ map[string]string) (money.Amount, string, error) {
	return p.repo.BaseRate(ctx, productCode, coverageCode)
}

// Factor resolves the factor for the dimension driving the given factor type.
func (p *RateProvider) Factor(ctx context.Context, productCode, coverageCode, factorType string, dims map[string]string) (decimal.Decimal, string, error) {
	return p.repo.Factor(ctx, productCode, coverageCode, factorType, dims[dimKeyFor(factorType)])
}

// dimKeyFor maps a factor type to the risk-dimension key that drives it.
func dimKeyFor(factorType string) string {
	switch factorType {
	case "AGE":
		return "age_band"
	default:
		return factorType
	}
}
