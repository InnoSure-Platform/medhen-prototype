// Package adapters contains driven adapters for the rating module. StaticRateTable
// is an in-memory RateTableProvider used for local/dev and tests; the product
// module supplies the production implementation once migrated.
package adapters

import (
	"context"
	"fmt"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/shopspring/decimal"
)

type baseKey struct{ product, coverage string }
type factorKey struct{ product, coverage, factorType, dim string }

// StaticRateTable is an in-memory pricing source.
type StaticRateTable struct {
	base    map[baseKey]money.Amount
	baseVer string
	factors map[factorKey]decimal.Decimal
}

// NewMotorRateTable seeds a demo Motor rate table (base rates + an AGE factor).
func NewMotorRateTable() *StaticRateTable {
	return &StaticRateTable{
		baseVer: "MOTOR-2026.1",
		base: map[baseKey]money.Amount{
			{"MOT", "OD"}:  money.FromInt(1200), // own damage
			{"MOT", "TPL"}: money.FromInt(800),  // third-party liability
		},
		factors: map[factorKey]decimal.Decimal{
			{"MOT", "OD", "AGE", "young"}: decimal.NewFromFloat(1.25),
			{"MOT", "OD", "AGE", "adult"}: decimal.NewFromFloat(1.00),
			{"MOT", "OD", "AGE", "senior"}: decimal.NewFromFloat(1.10),
		},
	}
}

// BaseRate returns the base premium for a product/coverage.
func (t *StaticRateTable) BaseRate(_ context.Context, product, coverage string, _ map[string]string) (money.Amount, string, error) {
	if a, ok := t.base[baseKey{product, coverage}]; ok {
		return a, t.baseVer, nil
	}
	return money.Zero(), "", fmt.Errorf("no base rate for %s/%s", product, coverage)
}

// Factor returns a rating factor keyed by a risk dimension; missing factors
// yield a not-found error (the engine defaults such factors to 1.0).
func (t *StaticRateTable) Factor(_ context.Context, product, coverage, factorType string, dims map[string]string) (decimal.Decimal, string, error) {
	dim := dims[dimKeyFor(factorType)]
	if f, ok := t.factors[factorKey{product, coverage, factorType, dim}]; ok {
		return f, t.baseVer, nil
	}
	return decimal.Zero, "", fmt.Errorf("no %s factor for %s/%s dim=%q", factorType, product, coverage, dim)
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
