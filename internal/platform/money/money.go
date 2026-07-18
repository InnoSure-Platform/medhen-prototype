// Package money is the single monetary type for the platform. All amounts are
// Ethiopian Birr (ETB), backed by an arbitrary-precision decimal so there is no
// float64 money anywhere. Intermediate arithmetic keeps extra precision;
// values are rounded to the currency boundary (2 dp, banker's rounding) only at
// the edges (persistence, display, payment).
package money

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	// InternalScale is the precision kept while chaining rating factors.
	InternalScale int32 = 6
	// CurrencyScale is the ETB currency boundary (2 decimal places).
	CurrencyScale int32 = 2
)

// Amount is an ETB monetary value.
type Amount struct {
	d decimal.Decimal
}

// Zero is 0.00 ETB.
func Zero() Amount { return Amount{d: decimal.Zero} }

// FromMinor builds an Amount from integer minor units (santim/cents): 1050 → 10.50.
func FromMinor(minor int64) Amount {
	return Amount{d: decimal.New(minor, -CurrencyScale)}
}

// FromInt builds an Amount from whole Birr.
func FromInt(birr int64) Amount { return Amount{d: decimal.NewFromInt(birr)} }

// FromDecimal wraps a decimal value.
func FromDecimal(d decimal.Decimal) Amount { return Amount{d: d} }

// FromString parses a decimal string such as "1234.56".
func FromString(s string) (Amount, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Amount{}, fmt.Errorf("money: parse %q: %w", s, err)
	}
	return Amount{d: d}, nil
}

// Decimal returns the underlying decimal (for gRPC/JSON boundaries).
func (a Amount) Decimal() decimal.Decimal { return a.d }

// MarshalJSON emits the currency-rounded value as a JSON number (e.g. 1234.56).
func (a Amount) MarshalJSON() ([]byte, error) {
	return a.d.RoundBank(CurrencyScale).MarshalJSON()
}

// UnmarshalJSON parses a JSON number/string into an Amount.
func (a *Amount) UnmarshalJSON(data []byte) error {
	var d decimal.Decimal
	if err := d.UnmarshalJSON(data); err != nil {
		return err
	}
	a.d = d
	return nil
}

// Add returns a+b at internal precision.
func (a Amount) Add(b Amount) Amount {
	return Amount{d: a.d.Add(b.d).Round(InternalScale)}
}

// Sub returns a-b at internal precision.
func (a Amount) Sub(b Amount) Amount {
	return Amount{d: a.d.Sub(b.d).Round(InternalScale)}
}

// Mul multiplies by a scalar factor (e.g. a rating factor), keeping internal precision.
func (a Amount) Mul(factor decimal.Decimal) Amount {
	return Amount{d: a.d.Mul(factor).Round(InternalScale)}
}

// Div divides by a scalar, returning an error rather than panicking on zero.
func (a Amount) Div(divisor decimal.Decimal) (Amount, error) {
	if divisor.IsZero() {
		return Amount{}, fmt.Errorf("money: division by zero")
	}
	return Amount{d: a.d.Div(divisor).Round(InternalScale)}, nil
}

// RoundCurrency rounds to the ETB currency boundary using banker's rounding.
func (a Amount) RoundCurrency() Amount {
	return Amount{d: a.d.RoundBank(CurrencyScale)}
}

// Minor returns the value in integer minor units after currency rounding.
func (a Amount) Minor() int64 {
	return a.d.RoundBank(CurrencyScale).Shift(CurrencyScale).IntPart()
}

// IsZero reports whether the amount is exactly zero.
func (a Amount) IsZero() bool { return a.d.IsZero() }

// IsNegative reports whether the amount is below zero.
func (a Amount) IsNegative() bool { return a.d.IsNegative() }

// Cmp compares two amounts: -1 if a<b, 0 if equal, +1 if a>b.
func (a Amount) Cmp(b Amount) int { return a.d.Cmp(b.d) }

// Equal reports value equality at currency precision.
func (a Amount) Equal(b Amount) bool {
	return a.d.RoundBank(CurrencyScale).Equal(b.d.RoundBank(CurrencyScale))
}

// String renders the currency-rounded value, e.g. "10.50 ETB".
func (a Amount) String() string {
	return a.d.RoundBank(CurrencyScale).StringFixedBank(CurrencyScale) + " ETB"
}

// Allocate splits the amount into len(ratios) parts by the given weights using
// the largest-remainder method on minor units, so no santim is lost or invented
// (sum of parts == RoundCurrency(a)). Panics on non-positive total weight.
func (a Amount) Allocate(ratios ...int) []Amount {
	if len(ratios) == 0 {
		return nil
	}
	total := 0
	for _, r := range ratios {
		total += r
	}
	if total <= 0 {
		panic("money: Allocate requires a positive total ratio")
	}

	whole := a.Minor()
	remainders := make([]struct {
		idx  int
		frac int64
	}, len(ratios))
	parts := make([]int64, len(ratios))
	var allocated int64

	for i, r := range ratios {
		// floor(whole * r / total), tracking the fractional remainder.
		numer := whole * int64(r)
		parts[i] = numer / int64(total)
		remainders[i] = struct {
			idx  int
			frac int64
		}{i, numer % int64(total)}
		allocated += parts[i]
	}

	leftover := whole - allocated
	// Distribute the leftover minor units to the largest remainders first.
	for i := 0; i < len(remainders); i++ {
		for j := i + 1; j < len(remainders); j++ {
			if remainders[j].frac > remainders[i].frac {
				remainders[i], remainders[j] = remainders[j], remainders[i]
			}
		}
	}
	for i := int64(0); i < leftover; i++ {
		parts[remainders[i].idx]++
	}

	out := make([]Amount, len(ratios))
	for i, p := range parts {
		out[i] = FromMinor(p)
	}
	return out
}
