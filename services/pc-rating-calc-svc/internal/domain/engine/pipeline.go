package engine

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel"

	"medhen.com/platform/pc-rating-calc-svc/internal/domain/math"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/models"
)

var tracer = otel.Tracer("medhen.com/platform/pc-rating-calc-svc/domain/engine")

// RateTableProvider defines the outbound port to fetch pricing matrices (Cache or gRPC)
type RateTableProvider interface {
	GetBaseRate(ctx context.Context, productCode string, coverageCode string, dims map[string]string) (decimal.Decimal, string, error)
	GetFactor(ctx context.Context, productCode string, coverageCode string, factorType string, dims map[string]string) (decimal.Decimal, string, error)
}

// RatingEngine orchestrates the pure domain calculation pipeline
type RatingEngine struct {
	provider RateTableProvider
}

// NewRatingEngine creates a new stateless engine
func NewRatingEngine(p RateTableProvider) *RatingEngine {
	return &RatingEngine{provider: p}
}

// CalculatePremium is the core pipeline executing the rating algorithm
func (e *RatingEngine) CalculatePremium(ctx context.Context, req models.RatingRequest) (*models.PremiumBreakdown, error) {
	ctx, span := tracer.Start(ctx, "CalculatePremium")
	defer span.End()
	
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	breakdown := &models.PremiumBreakdown{
		CalculationID: uuid.New().String(),
		NetPremium:    decimal.Zero,
		TotalTaxes:    decimal.Zero,
		GrossPremium:  decimal.Zero,
	}

	for _, cov := range req.SelectedCoverages {
		covNet, err := e.calculateCoverage(ctx, req, cov, breakdown, traceID, spanID)
		if err != nil {
			return nil, err
		}
		
		breakdown.CoverageBreakdowns = append(breakdown.CoverageBreakdowns, models.CoveragePremium{
			CoverageCode: cov,
			NetPremium:   covNet,
			BasePremium:  covNet, // Simplified for now
		})
		breakdown.NetPremium = breakdown.NetPremium.Add(covNet)
	}

	// Calculate Taxes (VAT 15%)
	vatRate := decimal.NewFromFloat(0.15)
	breakdown.TotalTaxes = math.RoundInternal(breakdown.NetPremium.Mul(vatRate))
	breakdown.AddTrace("CALCULATE_VAT", "15%", "SYSTEM_TAX", traceID, spanID)

	// Final Gross (Net + Taxes) - Round to Bank standard
	gross := breakdown.NetPremium.Add(breakdown.TotalTaxes)
	breakdown.GrossPremium = math.RoundFinal(gross)

	return breakdown, nil
}

// CalculateProRata executes the rating algorithm and applies a pro-rata factor
func (e *RatingEngine) CalculateProRata(ctx context.Context, req models.RatingRequest, daysActive, totalDays int64) (*models.PremiumBreakdown, error) {
	ctx, span := tracer.Start(ctx, "CalculateProRata")
	defer span.End()

	breakdown, err := e.CalculatePremium(ctx, req)
	if err != nil {
		return nil, err
	}

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	if totalDays == 0 {
		return nil, fmt.Errorf("totalDays cannot be zero")
	}

	proRataFactor, _ := math.SafeDiv(decimal.NewFromInt(daysActive), decimal.NewFromInt(totalDays))
	
	breakdown.NetPremium = math.RoundInternal(breakdown.NetPremium.Mul(proRataFactor))
	breakdown.TotalTaxes = math.RoundInternal(breakdown.TotalTaxes.Mul(proRataFactor))
	breakdown.GrossPremium = math.RoundFinal(breakdown.NetPremium.Add(breakdown.TotalTaxes))
	
	breakdown.AddTrace("PRO_RATA_ADJUSTMENT", fmt.Sprintf("%d/%d", daysActive, totalDays), "SYSTEM", traceID, spanID)

	return breakdown, nil
}

// calculateCoverage handles base rate and factor multiplication for a specific coverage
func (e *RatingEngine) calculateCoverage(ctx context.Context, req models.RatingRequest, cov string, bd *models.PremiumBreakdown, traceID, spanID string) (decimal.Decimal, error) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("calculateCoverage/%s", cov))
	defer span.End()

	// 1. Get Base Rate
	base, version, err := e.provider.GetBaseRate(ctx, req.ProductCode, cov, req.RiskDimensions)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get base rate for coverage %s: %w", cov, err)
	}
	bd.AddTrace(fmt.Sprintf("BASE_RATE_%s", cov), base.String(), version, traceID, spanID)

	// 2. Fetch specific factors (e.g., Age) - Mocked to show pipeline structure
	ageFactor, fVersion, err := e.provider.GetFactor(ctx, req.ProductCode, cov, "AGE", req.RiskDimensions)
	if err != nil {
		// Default to 1.0 if not strictly required, or fail
		ageFactor = decimal.NewFromInt(1)
	} else {
		bd.AddTrace(fmt.Sprintf("FACTOR_AGE_%s", cov), ageFactor.String(), fVersion, traceID, spanID)
	}

	net := math.MultiplyFactors(base, ageFactor)
	return net, nil
}
