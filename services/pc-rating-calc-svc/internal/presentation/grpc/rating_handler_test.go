package grpc_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"medhen.com/platform/pc-rating-calc-svc/internal/application"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/models"
	grpchandler "medhen.com/platform/pc-rating-calc-svc/internal/presentation/grpc"
)

type mockRateProvider struct{}

func (m *mockRateProvider) GetBaseRate(ctx context.Context, productCode string, coverageCode string, dims map[string]string) (decimal.Decimal, string, error) {
	return decimal.NewFromInt(100), "v1", nil
}
func (m *mockRateProvider) GetFactor(ctx context.Context, productCode string, coverageCode string, factorType string, dims map[string]string) (decimal.Decimal, string, error) {
	return decimal.NewFromInt(1), "v1", nil
}

type mockAuditProducer struct{}

func (m *mockAuditProducer) PublishRatingEvent(ctx context.Context, breakdown *models.PremiumBreakdown, req models.RatingRequest) error {
	return nil
}

func TestRatingHandler_CalculatePremium(t *testing.T) {
	appService := application.NewRatingApplicationService(&mockRateProvider{}, &mockAuditProducer{})
	handler := grpchandler.NewRatingHandler(appService)

	breakdown, err := handler.CalculatePremium(
		context.Background(),
		"tenant-1",
		"MOT-01",
		[]string{"COMP"},
		map[string]string{"age": "30"},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if breakdown == nil {
		t.Fatal("expected breakdown, got nil")
	}

	expectedGross, _ := decimal.NewFromString("115.00") // 100 * 1.0 * 1.15
	if !breakdown.GrossPremium.Equal(expectedGross) {
		t.Errorf("expected %s gross premium, got %s", expectedGross.String(), breakdown.GrossPremium.String())
	}
}
