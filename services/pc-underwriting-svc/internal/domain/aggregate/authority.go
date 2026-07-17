package aggregate

import (
	"github.com/google/uuid"
)

// AuthorityLevel configuration governing who can approve what.
type AuthorityLevel struct {
	ID         uuid.UUID
	TenantID   string
	LevelCode  string // e.g. L1, L2, COMMITTEE
	ProductLOB string
	MaxPremium float64
	MaxTSI     float64
}

// CanApprove checks if this authority level is sufficient for the quote financials.
func (a *AuthorityLevel) CanApprove(premium, tsi float64) bool {
	return premium <= a.MaxPremium && tsi <= a.MaxTSI
}

// NextLevel returns a mock of the next authority level logic (e.g. L1 -> L2).
// In a real system, this would be backed by a directed graph or integer rank of levels.
func (a *AuthorityLevel) NextLevel() string {
	switch a.LevelCode {
	case "L1":
		return "L2"
	case "L2":
		return "L3"
	case "L3":
		return "COMMITTEE"
	default:
		return "COMMITTEE"
	}
}
