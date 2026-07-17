package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// RatingRequest is the immutable command passed to the RatingEngine
type RatingRequest struct {
	RequestID         string            `validate:"required"`
	TenantID          string            `validate:"required"`
	ProductCode       string            `validate:"required"`
	AsOfDate          time.Time         `validate:"required"`
	RiskDimensions    map[string]string `validate:"omitempty"`
	SelectedCoverages []string          `validate:"required,min=1"`
}

// Validate ensures the request has the bare minimum payload before hitting the pipeline
func (r *RatingRequest) Validate() error {
	if r.RiskDimensions == nil {
		r.RiskDimensions = make(map[string]string)
	}
	return validate.Struct(r)
}
