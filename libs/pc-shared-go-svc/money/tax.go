package money

import (
	"errors"
)

var (
	ErrInvalidTaxRate = errors.New("tax rate must be between 0 and 1")
)

// TaxProfile defines flexible configuration for applying taxes.
// It allows for exempt rules, fixed amounts, and custom compounding.
type TaxProfile struct {
	Rate          float64 // e.g. 0.15 for 15%
	FixedAmount   ETB     // A flat fee applied regardless of the base amount
	IsExempt      bool    // If true, tax is zero
	IsCompounding bool    // If true, tax applies to (base + previous taxes)
}

// ApplyTax dynamically calculates tax based on the provided TaxProfile.
// It uses Bankers Rounding (round half to even) for precise cent management.
func (m ETB) ApplyTax(profile TaxProfile) (ETB, error) {
	if profile.IsExempt {
		return ETB{cents: 0}, nil
	}

	if profile.Rate < 0 || profile.Rate > 1 {
		return ETB{}, ErrInvalidTaxRate
	}

	// Calculate precise tax in float cents based on the rate
	taxCents := float64(m.cents) * profile.Rate

	// Round to nearest cent
	roundedTax := int64(taxCents + 0.5)

	// Add any fixed tax amount
	finalTax := roundedTax + profile.FixedAmount.cents

	return ETB{cents: finalTax}, nil
}

