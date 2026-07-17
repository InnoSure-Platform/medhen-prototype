package money

import (
	"errors"
	"fmt"
)

var (
	ErrNegativeAmount = errors.New("money amount cannot be negative")
)

// ETB represents Ethiopian Birr in cents (1 ETB = 100 Cents).
// Storing as int64 avoids floating-point inaccuracies.
type ETB struct {
	cents int64
}

// NewETB creates a new ETB amount given the integer cents.
func NewETB(cents int64) ETB {
	return ETB{cents: cents}
}

// FromFloat creates ETB from a float64 representation (e.g., 10.50 -> 1050 cents).
// It is recommended to use NewETB directly to avoid any floating-point issues.
func FromFloat(amount float64) ETB {
	return ETB{cents: int64(amount * 100)}
}

// Cents returns the raw integer cents value.
func (m ETB) Cents() int64 {
	return m.cents
}

// Float returns the float64 representation.
func (m ETB) Float() float64 {
	return float64(m.cents) / 100.0
}

// String returns the formatted ETB value.
func (m ETB) String() string {
	return fmt.Sprintf("%.2f ETB", m.Float())
}

// Add adds another ETB amount.
func (m ETB) Add(other ETB) ETB {
	return ETB{cents: m.cents + other.cents}
}

// Sub subtracts another ETB amount.
func (m ETB) Sub(other ETB) ETB {
	return ETB{cents: m.cents - other.cents}
}

// Mul multiplies the ETB amount by a scalar factor.
func (m ETB) Mul(factor float64) ETB {
	return ETB{cents: int64(float64(m.cents) * factor)}
}

// Allocate splits the ETB amount into portions based on the provided ratios.
// It guarantees that no fractional cents are lost or invented.
func (m ETB) Allocate(ratios []int) []ETB {
	if len(ratios) == 0 {
		return nil
	}

	totalRatio := 0
	for _, r := range ratios {
		totalRatio += r
	}

	if totalRatio == 0 {
		return nil
	}

	remainder := m.cents
	results := make([]ETB, len(ratios))

	for i, ratio := range ratios {
		// Calculate precise share
		share := (m.cents * int64(ratio)) / int64(totalRatio)
		results[i] = ETB{cents: share}
		remainder -= share
	}

	// Distribute remainder cent-by-cent to the largest ratios first to be fair,
	// but a simpler standard approach is just distributing to the first N.
	for i := 0; remainder > 0; i = (i + 1) % len(results) {
		results[i].cents++
		remainder--
	}

	return results
}
