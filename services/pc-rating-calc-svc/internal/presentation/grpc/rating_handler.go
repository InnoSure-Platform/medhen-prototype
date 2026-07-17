package grpc

import (
	"context"
	"time"

	"medhen.com/platform/pc-rating-calc-svc/internal/application"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/models"
)

// RatingHandler represents the gRPC controller
type RatingHandler struct {
	appService *application.RatingApplicationService
	// pb.UnimplementedRatingServiceServer
}

func NewRatingHandler(app *application.RatingApplicationService) *RatingHandler {
	return &RatingHandler{appService: app}
}

// CalculatePremium implements the gRPC contract
// This would take a pb.CalculateRequest and return a pb.CalculateResponse
// For this implementation, we simulate the domain translation
func (h *RatingHandler) CalculatePremium(ctx context.Context, tenantID, productCode string, coverages []string, dims map[string]string) (*models.PremiumBreakdown, error) {
	req := models.RatingRequest{
		RequestID:         "req-123", // normally from context
		TenantID:          tenantID,
		ProductCode:       productCode,
		AsOfDate:          time.Now(),
		SelectedCoverages: coverages,
		RiskDimensions:    dims,
	}

	return h.appService.CalculatePremium(ctx, req)
}

// CalculateProRata implements the gRPC contract for pro-rata midterm changes
func (h *RatingHandler) CalculateProRata(ctx context.Context, tenantID, productCode string, coverages []string, dims map[string]string, daysActive, totalDays int64) (*models.PremiumBreakdown, error) {
	req := models.RatingRequest{
		RequestID:         "req-124", // normally from context
		TenantID:          tenantID,
		ProductCode:       productCode,
		AsOfDate:          time.Now(),
		SelectedCoverages: coverages,
		RiskDimensions:    dims,
	}

	return h.appService.CalculateProRata(ctx, req, daysActive, totalDays)
}
