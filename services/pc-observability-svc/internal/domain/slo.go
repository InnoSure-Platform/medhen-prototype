package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidSLOTarget = errors.New("target percentage must be between 0 and 100")
	ErrEmptySLIQuery    = errors.New("SLI query cannot be empty")
)

type SLOStatus string

const (
	SLOStatusSyncing  SLOStatus = "SYNCING"
	SLOStatusActive   SLOStatus = "ACTIVE"
	SLOStatusBreached SLOStatus = "BREACHED"
	SLOStatusArchived SLOStatus = "ARCHIVED"
)

// SLO represents a Service Level Objective aggregate root.
type SLO struct {
	ID               string
	TenantID         string
	Name             string
	Description      string
	TargetPercentage float64
	WindowDays       int
	SLIQuery         string
	Status           SLOStatus
	AlertPolicyID    string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewSLO creates a new SLO aggregate.
func NewSLO(id, tenantID, name, desc string, target float64, window int, query, alertPolicyID string) (*SLO, error) {
	if target <= 0 || target >= 100 {
		return nil, ErrInvalidSLOTarget
	}
	if query == "" {
		return nil, ErrEmptySLIQuery
	}

	return &SLO{
		ID:               id,
		TenantID:         tenantID,
		Name:             name,
		Description:      desc,
		TargetPercentage: target,
		WindowDays:       window,
		SLIQuery:         query,
		Status:           SLOStatusSyncing,
		AlertPolicyID:    alertPolicyID,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}, nil
}

// MarkActive marks the SLO rules as successfully pushed to Mimir.
func (s *SLO) MarkActive() {
	s.Status = SLOStatusActive
	s.UpdatedAt = time.Now().UTC()
}
