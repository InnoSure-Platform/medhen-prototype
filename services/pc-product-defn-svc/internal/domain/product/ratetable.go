package product

import (
	"time"

	"github.com/google/uuid"
)

// RateTable represents a multi-dimensional pricing table.
type RateTable struct {
	ID            uuid.UUID
	TenantID      string
	Name          string
	Version       int
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	Dimensions    map[string]interface{}
	Rows          []*RateRow
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// RateRow represents a single entry in a RateTable.
type RateRow struct {
	ID             int64
	DimensionBounds map[string]interface{}
	Factor         float64
}

// NewRateTable creates a new RateTable.
func NewRateTable(tenantID, name string, dimensions map[string]interface{}) *RateTable {
	now := time.Now().UTC()
	return &RateTable{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Name:          name,
		Version:       1,
		EffectiveFrom: now,
		Dimensions:    dimensions,
		Rows:          make([]*RateRow, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// AddRow adds a row to the rate table.
func (rt *RateTable) AddRow(bounds map[string]interface{}, factor float64) {
	rt.Rows = append(rt.Rows, &RateRow{
		DimensionBounds: bounds,
		Factor:         factor,
	})
	rt.UpdatedAt = time.Now().UTC()
}
