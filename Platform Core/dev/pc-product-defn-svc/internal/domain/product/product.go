package product

import (
	"time"

	"github.com/google/uuid"
)

// Product is the Aggregate Root.
type Product struct {
	ID               uuid.UUID
	TenantID         string
	Code             string
	LOB              string
	Name             string
	Status           Status
	Version          int
	EffectiveFrom    time.Time
	EffectiveTo      *time.Time
	RequireFairValue bool
	SchemaPayload    map[string]interface{}
	Coverages        []*Coverage
	
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewProduct creates a new Product in the DRAFT state.
func NewProduct(tenantID, code, lob, name string, requireFairValue bool) *Product {
	now := time.Now().UTC()
	return &Product{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Code:             code,
		LOB:              lob,
		Name:             name,
		Status:           StatusDraft,
		Version:          1,
		EffectiveFrom:    now,
		RequireFairValue: requireFairValue,
		SchemaPayload:    make(map[string]interface{}),
		Coverages:        make([]*Coverage, 0),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// AddCoverage adds a coverage to the product.
func (p *Product) AddCoverage(cov *Coverage) error {
	if p.Status != StatusDraft {
		return ErrProductNotInDraft
	}

	// Enforce coverage dependencies
	if cov.ParentCoverageCode != nil {
		foundParent := false
		for _, existing := range p.Coverages {
			if existing.Code == *cov.ParentCoverageCode {
				foundParent = true
				break
			}
		}
		if !foundParent {
			return ErrParentCoverageMissing
		}
	}

	p.Coverages = append(p.Coverages, cov)
	p.incrementVersion()
	return nil
}

// TransitionTo attempts to transition the product to a new state.
func (p *Product) TransitionTo(next Status, hasFairValueAssessment bool) error {
	if !p.Status.CanTransitionTo(next) {
		return ErrInvalidStateTransition
	}

	// UK Consumer Duty block
	if p.Status == StatusDraft && next == StatusReview {
		if p.RequireFairValue && !hasFairValueAssessment {
			return ErrMissingFairValue
		}
	}

	p.Status = next
	p.incrementVersion()
	return nil
}

func (p *Product) incrementVersion() {
	p.Version++
	p.UpdatedAt = time.Now().UTC()
}
