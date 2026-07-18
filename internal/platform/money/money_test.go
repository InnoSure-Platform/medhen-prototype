package money

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFromMinorAndBack(t *testing.T) {
	a := FromMinor(1050)
	if got := a.Minor(); got != 1050 {
		t.Fatalf("Minor() = %d, want 1050", got)
	}
	if a.String() != "10.50 ETB" {
		t.Fatalf("String() = %q, want %q", a.String(), "10.50 ETB")
	}
}

func TestAddSub(t *testing.T) {
	sum := FromMinor(1050).Add(FromMinor(2075))
	if sum.Minor() != 3125 {
		t.Fatalf("Add minor = %d, want 3125", sum.Minor())
	}
	diff := FromMinor(3125).Sub(FromMinor(1075))
	if diff.Minor() != 2050 {
		t.Fatalf("Sub minor = %d, want 2050", diff.Minor())
	}
}

func TestMulKeepsPrecisionThenRounds(t *testing.T) {
	// 100.00 * 1.234567 = 123.4567 → currency round 123.46
	got := FromInt(100).Mul(decimal.NewFromFloat(1.234567)).RoundCurrency()
	if got.Minor() != 12346 {
		t.Fatalf("Mul/round minor = %d, want 12346", got.Minor())
	}
}

func TestBankersRounding(t *testing.T) {
	// 10.125 → 10.12 (round half to even); 10.135 → 10.14
	if got, _ := FromString("10.125"); got.Minor() != 1012 {
		t.Fatalf("10.125 → %d, want 1012 (banker's)", got.Minor())
	}
	if got, _ := FromString("10.135"); got.Minor() != 1014 {
		t.Fatalf("10.135 → %d, want 1014 (banker's)", got.Minor())
	}
}

func TestDivByZero(t *testing.T) {
	if _, err := FromInt(100).Div(decimal.Zero); err == nil {
		t.Fatal("expected division-by-zero error")
	}
	got, err := FromInt(100).Div(decimal.NewFromInt(4))
	if err != nil || got.Minor() != 2500 {
		t.Fatalf("Div = %v (%v), want 25.00", got, err)
	}
}

func TestAllocateNoLostCents(t *testing.T) {
	// 100.00 split 1:1:1 → parts must sum back to 10000 minor, no cent lost.
	parts := FromMinor(10000).Allocate(1, 1, 1)
	var sum int64
	for _, p := range parts {
		sum += p.Minor()
	}
	if sum != 10000 {
		t.Fatalf("allocate sum = %d, want 10000", sum)
	}
	// Largest-remainder gives 3334/3333/3333 in some order.
	if parts[0].Minor() != 3334 {
		t.Fatalf("first part = %d, want 3334 (largest remainder first)", parts[0].Minor())
	}
}

func TestAllocateByWeights(t *testing.T) {
	parts := FromMinor(10000).Allocate(70, 30) // commission split
	if parts[0].Minor() != 7000 || parts[1].Minor() != 3000 {
		t.Fatalf("weighted allocate = %d/%d, want 7000/3000", parts[0].Minor(), parts[1].Minor())
	}
}

func TestAllocatePanicsOnZeroTotal(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on zero total ratio")
		}
	}()
	FromMinor(100).Allocate(0, 0)
}

func TestCmpAndFlags(t *testing.T) {
	if !FromMinor(-5).IsNegative() {
		t.Fatal("expected negative")
	}
	if !Zero().IsZero() {
		t.Fatal("expected zero")
	}
	if FromInt(10).Cmp(FromInt(5)) != 1 {
		t.Fatal("expected 10 > 5")
	}
}

func TestVAT(t *testing.T) {
	// 15% VAT on 1000.00 = 150.00
	tax := VAT(EthiopiaVATRate).Apply(FromInt(1000))
	if tax.Minor() != 15000 {
		t.Fatalf("VAT = %d, want 15000", tax.Minor())
	}
}

func TestStampDuty(t *testing.T) {
	tax := StampDuty(FromInt(35)).Apply(FromInt(1000))
	if tax.Minor() != 3500 {
		t.Fatalf("stamp duty = %d, want 3500", tax.Minor())
	}
}

func TestTaxExempt(t *testing.T) {
	p := VAT(EthiopiaVATRate)
	p.Exempt = true
	if !p.Apply(FromInt(1000)).IsZero() {
		t.Fatal("exempt tax should be zero")
	}
}

func TestCombinedRateAndFixed(t *testing.T) {
	// A profile with both a rate and a fixed component.
	p := TaxProfile{Name: "COMBINED", Rate: decimal.NewFromFloat(0.10), Fixed: FromInt(5)}
	// 10% of 1000 = 100, + 5 fixed = 105.00
	if got := p.Apply(FromInt(1000)); got.Minor() != 10500 {
		t.Fatalf("combined = %d, want 10500", got.Minor())
	}
}
