package ratetable

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RateTable is an Aggregate Root representing a versioned multi-dimensional pricing table.
type RateTable struct {
	ID            uuid.UUID
	TenantID      string
	Name          string
	Version       int
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	Dimensions    []string // e.g., ["Age", "VehicleMake"]
	Rows          []RateRow
	
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// RateRow defines a single pricing factor for a set of dimension bounds.
type RateRow struct {
	ID              int64
	RateTableID     uuid.UUID
	DimensionBounds map[string]interface{}
	Factor          float64
}

// Repository defines the persistence interface for the RateTable.
type Repository interface {
	Save(ctx context.Context, rt *RateTable) error
	GetByID(ctx context.Context, tenantID string, id uuid.UUID) (*RateTable, error)
}
