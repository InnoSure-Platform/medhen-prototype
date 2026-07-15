// Package money provides ETB (Ethiopian Birr) money primitives in minor units (cents).
// All Medhen financial amounts are stored and transmitted as int64 minor units.
package money

import (
	"fmt"
	"math"
)

const CurrencyETB = "ETB"

// Amount is a currency amount in minor units (1 ETB = 100 minor).
type Amount struct {
	Minor    int64  `json:"minor"`
	Currency string `json:"currency"`
}

func NewETB(birr float64) Amount {
	return Amount{Minor: int64(math.Round(birr * 100)), Currency: CurrencyETB}
}

func FromMinor(minor int64, currency string) Amount {
	if currency == "" {
		currency = CurrencyETB
	}
	return Amount{Minor: minor, Currency: currency}
}

func (a Amount) Birr() float64 { return float64(a.Minor) / 100 }

func (a Amount) Add(b Amount) (Amount, error) {
	if a.Currency != b.Currency {
		return Amount{}, fmt.Errorf("currency mismatch: %s vs %s", a.Currency, b.Currency)
	}
	return Amount{Minor: a.Minor + b.Minor, Currency: a.Currency}, nil
}

func (a Amount) MulRate(rate float64) Amount {
	return Amount{Minor: int64(math.Round(float64(a.Minor) * rate)), Currency: a.Currency}
}

func (a Amount) Format() string {
	return fmt.Sprintf("%.2f %s", a.Birr(), a.Currency)
}

// VATEthiopia is the standard VAT rate (15%).
const VATEthiopia = 0.15

// StampDutyMotor is a simplified Phase 0 stamp duty flat rate on net premium (approx).
// Production rates are product-configured; this mirrors a representative EIC-style duty.
func StampDutyMotor(netPremium Amount) Amount {
	// 0.5% of net premium, minimum 25 ETB — illustrative Phase 0 synthetic.
	duty := netPremium.MulRate(0.005)
	min := NewETB(25)
	if duty.Minor < min.Minor {
		return min
	}
	return duty
}

func VATOn(net Amount) Amount {
	return net.MulRate(VATEthiopia)
}
