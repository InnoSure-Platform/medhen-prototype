// Package ports is the published contract of the rating module. Other modules
// (e.g. policy) depend only on this package — never on rating's domain or
// adapters — and receive a Calculator implementation via the composition root.
package ports

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// PremiumRequest is the input to a premium calculation.
type PremiumRequest struct {
	RequestID      string            `json:"request_id"`
	TenantID       string            `json:"tenant_id"`
	ProductCode    string            `json:"product_code"`
	AsOfDate       time.Time         `json:"as_of_date"`
	RiskDimensions map[string]string `json:"risk_dimensions"`
	Coverages      []string          `json:"coverages"`
}

// CoveragePremium is the premium attributed to one coverage.
type CoveragePremium struct {
	Code string       `json:"code"`
	Base money.Amount `json:"base"`
	Net  money.Amount `json:"net"`
}

// TaxLine is a single tax/levy applied to the net premium.
type TaxLine struct {
	Name   string       `json:"name"`
	Amount money.Amount `json:"amount"`
}

// AuditStep records one deterministic step of the calculation for explainability.
type AuditStep struct {
	Order     int    `json:"order"`
	Operation string `json:"operation"`
	Value     string `json:"value"`
	Version   string `json:"version"`
}

// PremiumBreakdown is the immutable result of a calculation. Components are
// currency-rounded and always sum to GrossPremium.
type PremiumBreakdown struct {
	CalculationID string            `json:"calculation_id"`
	Coverages     []CoveragePremium `json:"coverages"`
	NetPremium    money.Amount      `json:"net_premium"`
	Taxes         []TaxLine         `json:"taxes"`
	TotalTaxes    money.Amount      `json:"total_taxes"`
	GrossPremium  money.Amount      `json:"gross_premium"`
	Trace         []AuditStep       `json:"trace"`
}

// Calculator is the rating module's public capability, consumed in-process by
// other modules (this replaces the pre-refactor stubbed gRPC client).
type Calculator interface {
	Calculate(ctx context.Context, req PremiumRequest) (*PremiumBreakdown, error)
}

// RateTableProvider is the outbound port the rating engine depends on to fetch
// pricing data. It is satisfied by a static adapter today and by the product
// module once that migrates.
type RateTableProvider interface {
	BaseRate(ctx context.Context, productCode, coverageCode string, dims map[string]string) (money.Amount, string, error)
	Factor(ctx context.Context, productCode, coverageCode, factorType string, dims map[string]string) (decimal.Decimal, string, error)
}
