package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"medhen.com/platform/pc-rating-calc-svc/internal/application"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/models"
)

// mockRateProvider implements RateTableProvider
type mockRateProvider struct{}

func (m *mockRateProvider) GetBaseRate(ctx context.Context, productCode string, coverageCode string, dims map[string]string) (decimal.Decimal, string, error) {
	return decimal.NewFromInt(1000), "v1", nil
}

func (m *mockRateProvider) GetFactor(ctx context.Context, productCode string, coverageCode string, factorType string, dims map[string]string) (decimal.Decimal, string, error) {
	return decimal.NewFromInt(1), "v1", nil
}

// mockAuditProducer implements AuditEventProducer
type mockAuditProducer struct {
	called bool
}

func (m *mockAuditProducer) PublishRatingEvent(ctx context.Context, breakdown *models.PremiumBreakdown, req models.RatingRequest) error {
	m.called = true
	return nil
}

func TestRatingApplicationService_CalculatePremium(t *testing.T) {
	provider := &mockRateProvider{}
	producer := &mockAuditProducer{}

	appService := application.NewRatingApplicationService(provider, producer)

	req := models.RatingRequest{
		RequestID:   "req-1",
		TenantID:    "t-1",
		ProductCode: "MOT-01",
		AsOfDate:    time.Now(),
		SelectedCoverages: []string{"COMP"},
	}

	breakdown, err := appService.CalculatePremium(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if breakdown == nil {
		t.Fatal("expected breakdown, got nil")
	}

	// Give the async goroutine a tiny moment to run before asserting
	time.Sleep(10 * time.Millisecond)

	if !producer.called {
		t.Error("expected audit event producer to be called")
	}
}
