package product

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrParentCoverageMissing = errors.New("parent coverage dependency is missing")
)

// Coverage represents an insurance coverage attached to a Product.
type Coverage struct {
	ID                 uuid.UUID
	ProductID          uuid.UUID
	Code               string
	Name               string
	IsMandatory        bool
	MinLimit           *float64
	MaxLimit           *float64
	DeductibleConfig   map[string]interface{}
	ParentCoverageCode *string // Dependency mapping
}

// NewCoverage creates a new Coverage instance.
func NewCoverage(productID uuid.UUID, code, name string, isMandatory bool) *Coverage {
	return &Coverage{
		ID:               uuid.New(),
		ProductID:        productID,
		Code:             code,
		Name:             name,
		IsMandatory:      isMandatory,
		DeductibleConfig: make(map[string]interface{}),
	}
}

// SetLimits defines the optional min and max limits.
func (c *Coverage) SetLimits(min, max float64) {
	c.MinLimit = &min
	c.MaxLimit = &max
}

// SetParent dependency mapping.
func (c *Coverage) SetParent(parentCode string) {
	c.ParentCoverageCode = &parentCode
}
