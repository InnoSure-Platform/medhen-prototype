package rating_test

import (
	"testing"

	"github.com/InnoSure-Platform/pc-platform/internal/rating"
	"github.com/InnoSure-Platform/pc-shared-go/i18n"
)

func TestCalculateMotorComprehensive(t *testing.T) {
	res := rating.CalculateMotor(rating.Input{
		CoverType: "comprehensive", Usage: "private", Year: 2022,
		SumInsuredMinor: 100_000_000, // 1,000,000 ETB
		Locale:          i18n.EN,
	})
	if res.TotalMinor <= 0 {
		t.Fatal("expected positive premium")
	}
	if res.Currency != "ETB" {
		t.Fatalf("currency %s", res.Currency)
	}
	// base 3.5% of 1M = 35,000 → 3,500,000 cents; + VAT 15% + stamp
	if len(res.Lines) < 3 {
		t.Fatalf("expected itemized lines, got %d", len(res.Lines))
	}
}
