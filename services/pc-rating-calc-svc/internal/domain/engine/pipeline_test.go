package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/engine"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/models"
)

// mockRateProvider implements engine.RateTableProvider for testing
type mockRateProvider struct{}

func (m *mockRateProvider) GetBaseRate(ctx context.Context, productCode string, coverageCode string, dims map[string]string) (decimal.Decimal, string, error) {
	if coverageCode == "COMP" {
		return decimal.NewFromInt(1000), "v1", nil
	}
	return decimal.NewFromInt(500), "v1", nil
}

func (m *mockRateProvider) GetFactor(ctx context.Context, productCode string, coverageCode string, factorType string, dims map[string]string) (decimal.Decimal, string, error) {
	if factorType == "AGE" {
		age := dims["age"]
		if age == "20" {
			factor, _ := decimal.NewFromString("1.5")
			return factor, "v1", nil
		}
	}
	return decimal.NewFromInt(1), "v1", nil
}

func TestCalculatePremium_Success(t *testing.T) {
	provider := &mockRateProvider{}
	ratingEngine := engine.NewRatingEngine(provider)

	req := models.RatingRequest{
		RequestID:   "req-test-1",
		TenantID:    "t-1",
		ProductCode: "MOT-01",
		AsOfDate:    time.Now(),
		RiskDimensions: map[string]string{
			"age": "20",
		},
		SelectedCoverages: []string{"COMP"},
	}

	breakdown, err := ratingEngine.CalculatePremium(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base = 1000, Age Factor = 1.5 -> Net = 1500
	expectedNet, _ := decimal.NewFromString("1500")
	if !breakdown.NetPremium.Equal(expectedNet) {
		t.Errorf("expected net premium %s, got %s", expectedNet.String(), breakdown.NetPremium.String())
	}

	// Taxes = 15% of 1500 = 225
	expectedTaxes, _ := decimal.NewFromString("225")
	if !breakdown.TotalTaxes.Equal(expectedTaxes) {
		t.Errorf("expected taxes %s, got %s", expectedTaxes.String(), breakdown.TotalTaxes.String())
	}

	// Gross = 1500 + 225 = 1725
	expectedGross, _ := decimal.NewFromString("1725")
	if !breakdown.GrossPremium.Equal(expectedGross) {
		t.Errorf("expected gross premium %s, got %s", expectedGross.String(), breakdown.GrossPremium.String())
	}
}

func TestCalculatePremium_ValidationFailed(t *testing.T) {
	provider := &mockRateProvider{}
	ratingEngine := engine.NewRatingEngine(provider)

	req := models.RatingRequest{
		// Missing TenantID and ProductCode
	}

	_, err := ratingEngine.CalculatePremium(context.Background(), req)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func FuzzCalculatePremium(f *testing.F) {
	// Seed with normal parameters
	f.Add("MOT-01", "COMP", "30", int64(30), int64(365))
	
	f.Fuzz(func(t *testing.T, product, coverage, age string, daysActive, totalDays int64) {
		provider := &mockRateProvider{}
		engine := engine.NewRatingEngine(provider)
		
		req := models.RatingRequest{
			RequestID:         "fuzz-req",
			TenantID:          "fuzz-tenant",
			ProductCode:       product,
			AsOfDate:          time.Now(),
			SelectedCoverages: []string{coverage},
			RiskDimensions: map[string]string{
				"age": age,
			},
		}

		// Ensure nothing panics under garbage string dimensions or random ProRata bounds
		_, _ = engine.CalculatePremium(context.Background(), req)

		if totalDays != 0 {
			_, _ = engine.CalculateProRata(context.Background(), req, daysActive, totalDays)
		}
	})
}
