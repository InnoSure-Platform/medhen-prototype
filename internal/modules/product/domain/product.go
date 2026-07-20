// Package domain is the product-definition bounded context: insurance products,
// their coverages, base rates and rating factors. Pure domain (no framework/DB).
package domain

import (
	"errors"

	"github.com/shopspring/decimal"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// Status is the product lifecycle state.
type Status string

const (
	StatusDraft   Status = "DRAFT"
	StatusActive  Status = "ACTIVE"
	StatusRetired Status = "RETIRED"
)

var (
	ErrCodeRequired    = errors.New("product: code is required")
	ErrNoCoverages     = errors.New("product: at least one coverage is required")
	ErrCoverageInvalid = errors.New("product: coverage code and name are required")
)

// Coverage is a purchasable cover within a product, with its base premium.
type Coverage struct {
	Code        string
	Name        string
	NameAmharic string
	BaseRate    money.Amount
}

// Factor is a rating multiplier keyed by a coverage, factor type and dimension
// value (e.g. AGE / "young" → 1.25).
type Factor struct {
	CoverageCode string
	FactorType   string
	Dimension    string
	Value        decimal.Decimal
}

// Product is the aggregate root.
type Product struct {
	Code        string
	LOB         string
	Name        string
	NameAmharic string
	Status      Status
	RateVersion string
	Coverages   []Coverage
	Factors     []Factor
}

// NewProduct validates and constructs an active product.
func NewProduct(code, lob, name, nameAmharic, rateVersion string, coverages []Coverage, factors []Factor) (*Product, error) {
	if code == "" {
		return nil, ErrCodeRequired
	}
	if len(coverages) == 0 {
		return nil, ErrNoCoverages
	}
	for _, c := range coverages {
		if c.Code == "" || c.Name == "" {
			return nil, ErrCoverageInvalid
		}
	}
	return &Product{
		Code: code, LOB: lob, Name: name, NameAmharic: nameAmharic,
		Status: StatusActive, RateVersion: rateVersion,
		Coverages: coverages, Factors: factors,
	}, nil
}
