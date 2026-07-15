package rating

import (
	"time"

	"github.com/InnoSure-Platform/pc-shared-go/i18n"
	"github.com/InnoSure-Platform/pc-shared-go/money"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
)

// Input is the Phase 0 motor rating request.
type Input struct {
	CoverType       string
	Usage           string
	Year            int
	SumInsuredMinor int64
	Locale          i18n.Locale
}

type Result struct {
	Lines      []store.PremiumLine
	TotalMinor int64
	Currency   string
}

// CalculateMotor computes itemized premium for EIC Motor Phase 0 (synthetic tariff).
// Base: comprehensive = 3.5% of SI, third_party = fixed 2,500 ETB + 0.15% SI.
// Factors: age (>10y +15%), commercial +20%. Then VAT 15% and stamp duty.
func CalculateMotor(in Input) Result {
	si := money.FromMinor(in.SumInsuredMinor, money.CurrencyETB)
	var base money.Amount
	switch in.CoverType {
	case "third_party":
		base, _ = money.NewETB(2500).Add(si.MulRate(0.0015))
	default: // comprehensive
		base = si.MulRate(0.035)
	}

	lines := []store.PremiumLine{{
		Code: "BASE", Label: i18n.T("premium.base", i18n.EN), LabelAm: i18n.T("premium.base", i18n.AM), AmountMinor: base.Minor,
	}}
	net := base

	age := time.Now().UTC().Year() - in.Year
	if age > 10 {
		factor := base.MulRate(0.15)
		lines = append(lines, store.PremiumLine{
			Code: "FAC_AGE", Label: i18n.T("premium.factor.age", i18n.EN), LabelAm: i18n.T("premium.factor.age", i18n.AM), AmountMinor: factor.Minor,
		})
		net, _ = net.Add(factor)
	}
	if in.Usage == "commercial" {
		factor := base.MulRate(0.20)
		lines = append(lines, store.PremiumLine{
			Code: "FAC_USAGE", Label: i18n.T("premium.factor.usage", i18n.EN), LabelAm: i18n.T("premium.factor.usage", i18n.AM), AmountMinor: factor.Minor,
		})
		net, _ = net.Add(factor)
	}

	vat := money.VATOn(net)
	stamp := money.StampDutyMotor(net)
	lines = append(lines,
		store.PremiumLine{Code: "VAT", Label: i18n.T("premium.vat", i18n.EN), LabelAm: i18n.T("premium.vat", i18n.AM), AmountMinor: vat.Minor},
		store.PremiumLine{Code: "STAMP", Label: i18n.T("premium.stamp", i18n.EN), LabelAm: i18n.T("premium.stamp", i18n.AM), AmountMinor: stamp.Minor},
	)
	total, _ := net.Add(vat)
	total, _ = total.Add(stamp)

	return Result{Lines: lines, TotalMinor: total.Minor, Currency: money.CurrencyETB}
}
