package math

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	// InternalPrecision defines the precision kept during intermediate pipeline steps
	InternalPrecision int32 = 6
	// FinalPrecision defines the currency boundary rounding for Ethiopian Birr (ETB)
	FinalPrecision int32 = 2
)

func init() {
	// Set global shopspring/decimal rules if necessary, but we prefer explicit rounding
}

// RoundInternal retains 6 decimal places for safe chaining of multiplicative factors
func RoundInternal(d decimal.Decimal) decimal.Decimal {
	return d.Round(InternalPrecision)
}

// RoundFinal rounds to 2 decimal places using bank-standard HalfEven rounding
func RoundFinal(d decimal.Decimal) decimal.Decimal {
	return d.RoundBank(FinalPrecision)
}

// SafeDiv performs a division, avoiding panics on division by zero
func SafeDiv(a, b decimal.Decimal) (decimal.Decimal, error) {
	if b.IsZero() {
		return decimal.Zero, fmt.Errorf("division by zero attempted in financial math")
	}
	return RoundInternal(a.Div(b)), nil
}

// MultiplyFactors takes a base and variadic factors, multiplying them with internal precision
func MultiplyFactors(base decimal.Decimal, factors ...decimal.Decimal) decimal.Decimal {
	result := base
	for _, f := range factors {
		result = RoundInternal(result.Mul(f))
	}
	return result
}
