package math_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/math"
)

func TestRoundInternal(t *testing.T) {
	d, _ := decimal.NewFromString("1.23456789")
	rounded := math.RoundInternal(d)

	if rounded.String() != "1.234568" {
		t.Errorf("expected 1.234568, got %s", rounded.String())
	}
}

func TestRoundFinal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.234", "1.23"},
		{"1.235", "1.24"}, // Half-even rounding (towards nearest even when equidistant)
		{"1.245", "1.24"},
		{"1.255", "1.26"},
	}

	for _, tc := range tests {
		d, _ := decimal.NewFromString(tc.input)
		rounded := math.RoundFinal(d)
		if rounded.String() != tc.expected {
			t.Errorf("input %s: expected %s, got %s", tc.input, tc.expected, rounded.String())
		}
	}
}

func TestSafeDiv(t *testing.T) {
	a := decimal.NewFromInt(10)
	b := decimal.NewFromInt(3)

	result, err := math.SafeDiv(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "3.333333"
	if result.String() != expected {
		t.Errorf("expected %s, got %s", expected, result.String())
	}

	_, err = math.SafeDiv(a, decimal.Zero)
	if err == nil {
		t.Errorf("expected error on division by zero")
	}
}

func TestMultiplyFactors(t *testing.T) {
	base := decimal.NewFromInt(1000)
	f1, _ := decimal.NewFromString("1.5")
	f2, _ := decimal.NewFromString("1.2")

	result := math.MultiplyFactors(base, f1, f2)
	expected := "1800"

	if !result.Equal(decimal.NewFromInt(1800)) {
		t.Errorf("expected %s, got %s", expected, result.String())
	}
}

// Fuzzing protects Tier-0 math from precision overflow panics
func FuzzMultiplyFactors(f *testing.F) {
	// Seed corpus
	f.Add(1000.0, 1.5, 1.2)
	f.Add(0.0, 0.0, 0.0)
	f.Add(-100.5, 2.0, -1.0)
	
	f.Fuzz(func(t *testing.T, baseFloat, f1Float, f2Float float64) {
		base := decimal.NewFromFloat(baseFloat)
		f1 := decimal.NewFromFloat(f1Float)
		f2 := decimal.NewFromFloat(f2Float)

		// The goal of this fuzz test is to ensure the pipeline NEVER panics under arbitrary floats
		_ = math.MultiplyFactors(base, f1, f2)
	})
}

func FuzzSafeDiv(f *testing.F) {
	f.Add(1000.0, 3.0)
	f.Add(100.0, 0.0) // Expected to return error, not panic
	
	f.Fuzz(func(t *testing.T, num, den float64) {
		a := decimal.NewFromFloat(num)
		b := decimal.NewFromFloat(den)

		// Must not panic, must handle 0 gracefully
		_, _ = math.SafeDiv(a, b)
	})
}
