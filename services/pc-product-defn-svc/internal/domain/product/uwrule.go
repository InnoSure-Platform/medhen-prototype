package product

import (
	"time"

	"github.com/google/uuid"
)

// UnderwritingAction defines the result of a rule.
type UnderwritingAction string

const (
	ActionAutoAccept UnderwritingAction = "AUTO_ACCEPT"
	ActionRefer      UnderwritingAction = "REFER"
	ActionDecline    UnderwritingAction = "DECLINE"
)

// UnderwritingRuleSet groups rules for a product version.
type UnderwritingRuleSet struct {
	ID        uuid.UUID
	TenantID  string
	ProductID uuid.UUID
	Version   int
	Rules     []*UnderwritingRule
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UnderwritingRule represents a single evaluation rule.
type UnderwritingRule struct {
	ID        uuid.UUID
	Priority  int
	Condition map[string]interface{}
	Action    UnderwritingAction
	Message   string
}

// NewUnderwritingRuleSet creates a new rule set.
func NewUnderwritingRuleSet(tenantID string, productID uuid.UUID) *UnderwritingRuleSet {
	now := time.Now().UTC()
	return &UnderwritingRuleSet{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ProductID: productID,
		Version:   1,
		Rules:     make([]*UnderwritingRule, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddRule adds a rule to the set.
func (urs *UnderwritingRuleSet) AddRule(priority int, condition map[string]interface{}, action UnderwritingAction, message string) {
	urs.Rules = append(urs.Rules, &UnderwritingRule{
		ID:        uuid.New(),
		Priority:  priority,
		Condition: condition,
		Action:    action,
		Message:   message,
	})
	urs.UpdatedAt = time.Now().UTC()
}
