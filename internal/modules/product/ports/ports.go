// Package ports is the published contract of the product module.
package ports

import (
	"context"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// CoverageView is a coverage in the public product read model.
type CoverageView struct {
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	NameAmharic string       `json:"name_amharic"`
	BaseRate    money.Amount `json:"base_rate"`
}

// ProductView is the public read model of a product.
type ProductView struct {
	Code        string         `json:"code"`
	LOB         string         `json:"lob"`
	Name        string         `json:"name"`
	NameAmharic string         `json:"name_amharic"`
	Status      string         `json:"status"`
	Coverages   []CoverageView `json:"coverages"`
}

// Catalog is the product module's public read capability.
type Catalog interface {
	GetProduct(ctx context.Context, code string) (ProductView, error)
	ListProducts(ctx context.Context) ([]ProductView, error)
}
