package models_test

import (
	"testing"
	"time"

	"medhen.com/platform/pc-rating-calc-svc/internal/domain/models"
)

func TestRatingRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     models.RatingRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			req: models.RatingRequest{
				RequestID:         "R1",
				TenantID:          "T1",
				ProductCode:       "P1",
				AsOfDate:          time.Now(),
				SelectedCoverages: []string{"C1"},
			},
			wantErr: false,
		},
		{
			name: "Missing TenantID",
			req: models.RatingRequest{
				RequestID:         "R1",
				ProductCode:       "P1",
				AsOfDate:          time.Now(),
				SelectedCoverages: []string{"C1"},
			},
			wantErr: true,
		},
		{
			name: "Missing ProductCode",
			req: models.RatingRequest{
				TenantID:          "T1",
				SelectedCoverages: []string{"C1"},
			},
			wantErr: true,
		},
		{
			name: "Missing Coverages",
			req: models.RatingRequest{
				TenantID:    "T1",
				ProductCode: "P1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
