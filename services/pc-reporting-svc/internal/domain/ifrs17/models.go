package ifrs17

import "time"

type ContractGroup struct {
	GroupID      string
	TenantID     string
	Inception    time.Time
	CoverageTerm int // months
}

type CashFlow struct {
	Period int
	Amount float64
	Type   CashFlowType
}

type CashFlowType string

const (
	Premium     CashFlowType = "PREMIUM"
	Claim       CashFlowType = "CLAIM"
	Acquisition CashFlowType = "ACQUISITION"
)

type MeasurementResult struct {
	PresentValueCashFlows float64
	RiskAdjustment        float64
	ContractualServiceMargin float64
}
