package money

import "github.com/shopspring/decimal"

// TaxProfile describes a single levy applied to a base amount. It models both
// ad-valorem taxes (VAT, a rate) and fixed charges (stamp duty, a flat amount);
// a profile may carry both.
type TaxProfile struct {
	// Name identifies the levy (e.g. "VAT", "STAMP_DUTY") for breakdown lines.
	Name string
	// Rate is the ad-valorem rate, e.g. 0.15 for 15% VAT. Zero means no rate component.
	Rate decimal.Decimal
	// Fixed is a flat amount added regardless of the base (e.g. stamp duty).
	Fixed Amount
	// Exempt short-circuits the levy to zero when true.
	Exempt bool
}

// VAT returns a rate-based tax profile.
func VAT(rate decimal.Decimal) TaxProfile {
	return TaxProfile{Name: "VAT", Rate: rate}
}

// StampDuty returns a fixed-amount tax profile.
func StampDuty(fixed Amount) TaxProfile {
	return TaxProfile{Name: "STAMP_DUTY", Fixed: fixed}
}

// Apply computes the tax due on base for this profile, rounded to the currency
// boundary. The rate component and the fixed component are summed.
func (p TaxProfile) Apply(base Amount) Amount {
	if p.Exempt {
		return Zero()
	}
	tax := p.Fixed
	if !p.Rate.IsZero() {
		tax = tax.Add(base.Mul(p.Rate))
	}
	return tax.RoundCurrency()
}

// EthiopiaVATRate is the standard 15% VAT rate.
var EthiopiaVATRate = decimal.NewFromFloat(0.15)
