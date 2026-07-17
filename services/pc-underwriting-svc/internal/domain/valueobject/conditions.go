package valueobject

type ConditionType string
type ValueType string

const (
	ConditionPremiumLoading   ConditionType = "PREMIUM_LOADING"
	ConditionMandatoryExclude ConditionType = "MANDATORY_EXCLUSION"
	ConditionSpecialDeduct    ConditionType = "SPECIAL_DEDUCTIBLE"

	ValueTypePercentage ValueType = "PERCENTAGE"
	ValueTypeFlat       ValueType = "FLAT_AMOUNT"
)

// Condition represents a loading or exclusion applied to a referral decision.
type Condition struct {
	Type           ConditionType `json:"type"`
	ValueType      ValueType     `json:"value_type,omitempty"`
	Value          float64       `json:"value,omitempty"`
	TargetCoverage string        `json:"target_coverage,omitempty"`
	Text           string        `json:"text,omitempty"`
}

// RiskScore represents the computed score from the rules engine.
type RiskScore struct {
	Value int `json:"value"` // 0 - 100
}

// DecisionType represents the underwriter's decision outcome.
type DecisionType string

const (
	DecisionApprove                 DecisionType = "APPROVE"
	DecisionApproveWithConditions   DecisionType = "APPROVE_WITH_CONDITIONS"
	DecisionDecline                 DecisionType = "DECLINE"
	DecisionReferHigher             DecisionType = "REFER_HIGHER"
)
