package aggregate

import (
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-underwriting-svc/internal/domain/valueobject"
)

type AssessmentStatus string

const (
	StatusSTPAccept AssessmentStatus = "STP_ACCEPT"
	StatusSTPDecline AssessmentStatus = "STP_DECLINE"
	StatusReferred   AssessmentStatus = "REFERRED"
	StatusInvalidated AssessmentStatus = "INVALIDATED"
)

// UnderwritingAssessment represents the point-in-time evaluation of a quote submission.
type UnderwritingAssessment struct {
	ID             uuid.UUID
	TenantID       string
	QuoteID        uuid.UUID
	ProductID      string
	Status         AssessmentStatus
	RiskScore      valueobject.RiskScore
	RulesTriggered []string
	CreatedAt      time.Time
}

// NewUnderwritingAssessment creates a new assessment aggregate.
func NewUnderwritingAssessment(tenantID string, quoteID uuid.UUID, productID string, score int, rules []string, status AssessmentStatus) (*UnderwritingAssessment, error) {
	return &UnderwritingAssessment{
		ID:             uuid.New(),
		TenantID:       tenantID,
		QuoteID:        quoteID,
		ProductID:      productID,
		Status:         status,
		RiskScore:      valueobject.RiskScore{Value: score},
		RulesTriggered: rules,
		CreatedAt:      time.Now().UTC(),
	}, nil
}

// Invalidate transitions the assessment to an INVALIDATED state if the quote is amended.
func (a *UnderwritingAssessment) Invalidate() {
	a.Status = StatusInvalidated
}
