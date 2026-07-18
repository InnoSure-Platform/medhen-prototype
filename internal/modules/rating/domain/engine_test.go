package domain_test

import (
	"context"
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/shopspring/decimal"
)

func newEngine() *domain.Engine {
	return domain.NewEngine(
		adapters.NewMotorRateTable(),
		domain.TaxPolicy{VATRate: decimal.NewFromFloat(0.15), StampDuty: money.FromInt(35)},
	)
}

func req(coverages ...string) ports.PremiumRequest {
	return ports.PremiumRequest{
		TenantID: "eic", ProductCode: "MOT",
		RiskDimensions: map[string]string{"age_band": "adult"},
		Coverages:      coverages,
	}
}

func TestCalculate_ComponentsSumToGross(t *testing.T) {
	bd, err := newEngine().Calculate(context.Background(), req("OD", "TPL"))
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	// OD 1200 * 1.00 (adult) + TPL 800 * default(1.0) = 2000.00 net
	if bd.NetPremium.Minor() != 200000 {
		t.Fatalf("net = %d, want 200000", bd.NetPremium.Minor())
	}
	// VAT 15% of 2000 = 300.00, + stamp duty 35.00 = 335.00 taxes
	if bd.TotalTaxes.Minor() != 33500 {
		t.Fatalf("taxes = %d, want 33500", bd.TotalTaxes.Minor())
	}
	// gross = 2335.00
	if bd.GrossPremium.Minor() != 233500 {
		t.Fatalf("gross = %d, want 233500", bd.GrossPremium.Minor())
	}

	// Regression for L1: displayed components must sum exactly to gross.
	sum := bd.NetPremium.Add(bd.TotalTaxes)
	if !sum.Equal(bd.GrossPremium) {
		t.Fatalf("net+taxes (%s) != gross (%s)", sum, bd.GrossPremium)
	}
}

func TestCalculate_AppliesAgeFactor(t *testing.T) {
	r := req("OD")
	r.RiskDimensions["age_band"] = "young" // 1.25 factor on OD 1200 = 1500.00
	bd, err := newEngine().Calculate(context.Background(), r)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if bd.NetPremium.Minor() != 150000 {
		t.Fatalf("young-driver net = %d, want 150000", bd.NetPremium.Minor())
	}
}

func TestCalculate_IncludesStampDutyLine(t *testing.T) {
	bd, _ := newEngine().Calculate(context.Background(), req("TPL"))
	var hasVAT, hasStamp bool
	for _, tl := range bd.Taxes {
		if tl.Name == "VAT" {
			hasVAT = true
		}
		if tl.Name == "STAMP_DUTY" && tl.Amount.Minor() == 3500 {
			hasStamp = true
		}
	}
	if !hasVAT || !hasStamp {
		t.Fatalf("expected VAT and STAMP_DUTY lines, got %+v", bd.Taxes)
	}
}

func TestCalculate_ValidationErrors(t *testing.T) {
	e := newEngine()
	cases := []ports.PremiumRequest{
		{ProductCode: "MOT", Coverages: []string{"OD"}},        // no tenant
		{TenantID: "eic", Coverages: []string{"OD"}},           // no product
		{TenantID: "eic", ProductCode: "MOT"},                  // no coverages
	}
	for i, c := range cases {
		if _, err := e.Calculate(context.Background(), c); err == nil {
			t.Fatalf("case %d: expected validation error", i)
		}
	}
}

func TestCalculate_UnknownCoverageFails(t *testing.T) {
	if _, err := newEngine().Calculate(context.Background(), req("NOPE")); err == nil {
		t.Fatal("expected error for unknown coverage (no base rate)")
	}
}

func TestCalculate_ProducesAuditTrace(t *testing.T) {
	bd, _ := newEngine().Calculate(context.Background(), req("OD"))
	if len(bd.Trace) == 0 {
		t.Fatal("expected a non-empty calculation trace")
	}
	if bd.CalculationID == "" {
		t.Fatal("expected a calculation ID")
	}
}
