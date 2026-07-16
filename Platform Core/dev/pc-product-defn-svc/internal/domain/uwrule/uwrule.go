package uwrule

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Action defines the underwriting action.
type Action string

const (
	ActionAutoAccept Action = "AUTO_ACCEPT"
	ActionReferral   Action = "REFERRAL"
	ActionDecline    Action = "DECLINE"
)

// UWRuleSet is an Aggregate Root representing the underwriting rules for a Product.
type UWRuleSet struct {
	ID        uuid.UUID
	TenantID  string
	ProductID uuid.UUID
	Version   int
	Rules     []Rule
	
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Rule defines a single underwriting condition.
type Rule struct {
	ID        uuid.UUID
	Priority  int
	Condition map[string]interface{} // JSON Logic expression
	Action    Action
	Message   string
}

// Repository defines the persistence interface for the UWRuleSet.
type Repository interface {
	Save(ctx context.Context, set *UWRuleSet) error
	GetByProductID(ctx context.Context, tenantID string, productID uuid.UUID) (*UWRuleSet, error)
}
