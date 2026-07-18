// Package domain holds the rating engine — a pure, deterministic premium
// calculation over the money kernel. It has no framework, DB, or HTTP
// dependencies; pricing data arrives through the RateTableProvider port.
package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/shopspring/decimal"
)

// TaxPolicy configures the levies applied to net premium. In a full system this
// comes from product-defn; here it is injected at wiring time.
type TaxPolicy struct {
	VATRate   decimal.Decimal
	StampDuty money.Amount
}

// Engine computes premiums. It is stateless and safe for concurrent use.
type Engine struct {
	provider    ports.RateTableProvider
	tax         TaxPolicy
	factorTypes []string
}

// NewEngine builds an engine. factorTypes defaults to ["AGE"] when empty.
func NewEngine(provider ports.RateTableProvider, tax TaxPolicy, factorTypes ...string) *Engine {
	if len(factorTypes) == 0 {
		factorTypes = []string{"AGE"}
	}
	return &Engine{provider: provider, tax: tax, factorTypes: factorTypes}
}

// Calculate runs the rating pipeline: per-coverage base × factors → net, then
// VAT and stamp duty → gross. All returned amounts are currency-rounded and the
// components sum exactly to GrossPremium.
func (e *Engine) Calculate(ctx context.Context, req ports.PremiumRequest) (*ports.PremiumBreakdown, error) {
	if err := validate(req); err != nil {
		return nil, err
	}

	bd := &ports.PremiumBreakdown{CalculationID: ids.New()}
	net := money.Zero()

	for _, cov := range req.Coverages {
		base, version, err := e.provider.BaseRate(ctx, req.ProductCode, cov, req.RiskDimensions)
		if err != nil {
			return nil, fmt.Errorf("rating: base rate for %s/%s: %w", req.ProductCode, cov, err)
		}
		addTraceStep(bd, fmt.Sprintf("BASE_RATE:%s", cov), base.String(), version)

		covNet := base
		for _, ft := range e.factorTypes {
			factor, fver, ferr := e.provider.Factor(ctx, req.ProductCode, cov, ft, req.RiskDimensions)
			if ferr != nil {
				factor, fver = decimal.NewFromInt(1), "default"
			}
			covNet = covNet.Mul(factor)
			addTraceStep(bd, fmt.Sprintf("FACTOR:%s:%s", ft, cov), factor.String(), fver)
		}

		covNet = covNet.RoundCurrency()
		bd.Coverages = append(bd.Coverages, ports.CoveragePremium{
			Code: cov, Base: base.RoundCurrency(), Net: covNet,
		})
		net = net.Add(covNet)
	}

	net = net.RoundCurrency()
	bd.NetPremium = net

	// VAT (ad-valorem) then stamp duty (fixed) — currency-rounded per line.
	vat := money.VAT(e.tax.VATRate).Apply(net)
	bd.Taxes = append(bd.Taxes, ports.TaxLine{Name: "VAT", Amount: vat})
	addTraceStep(bd, "VAT", e.tax.VATRate.String(), "policy")
	total := vat

	if !e.tax.StampDuty.IsZero() {
		bd.Taxes = append(bd.Taxes, ports.TaxLine{Name: "STAMP_DUTY", Amount: e.tax.StampDuty})
		addTraceStep(bd, "STAMP_DUTY", e.tax.StampDuty.String(), "policy")
		total = total.Add(e.tax.StampDuty)
	}

	bd.TotalTaxes = total.RoundCurrency()
	bd.GrossPremium = net.Add(bd.TotalTaxes).RoundCurrency()
	return bd, nil
}

func validate(req ports.PremiumRequest) error {
	switch {
	case req.TenantID == "":
		return errors.New("rating: tenant_id required")
	case req.ProductCode == "":
		return errors.New("rating: product_code required")
	case len(req.Coverages) == 0:
		return errors.New("rating: at least one coverage required")
	}
	return nil
}

// addTrace is a small helper kept on the breakdown DTO via the domain layer.
func addTraceStep(bd *ports.PremiumBreakdown, op, val, version string) {
	bd.Trace = append(bd.Trace, ports.AuditStep{
		Order: len(bd.Trace) + 1, Operation: op, Value: val, Version: version,
	})
}
