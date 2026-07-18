package ifrs17

import (
	"math"
)

type Calculator struct {
	DiscountRate float64
	RiskMargin   float64
}

func NewCalculator(discountRate float64, riskMargin float64) *Calculator {
	return &Calculator{
		DiscountRate: discountRate,
		RiskMargin:   riskMargin,
	}
}

// CalculateMeasurement measures a contract group under the General Measurement Model (GMM)
func (c *Calculator) CalculateMeasurement(group ContractGroup, flows []CashFlow) *MeasurementResult {
	var pvInflows float64
	var pvOutflows float64

	for _, flow := range flows {
		// Discount factor: 1 / (1 + r)^n
		discountFactor := 1.0 / math.Pow(1.0+c.DiscountRate, float64(flow.Period))
		discountedAmount := flow.Amount * discountFactor

		if flow.Type == Premium {
			pvInflows += discountedAmount
		} else {
			pvOutflows += discountedAmount
		}
	}

	netPresentValue := pvOutflows - pvInflows // IFRS 17 treats outflows as positive liability
	riskAdj := math.Abs(netPresentValue) * c.RiskMargin

	// CSM is established to prevent day-1 gains. If NPV + RA < 0, CSM = -(NPV + RA)
	csm := 0.0
	if netPresentValue+riskAdj < 0 {
		csm = -(netPresentValue + riskAdj)
	}

	return &MeasurementResult{
		PresentValueCashFlows:    netPresentValue,
		RiskAdjustment:           riskAdj,
		ContractualServiceMargin: csm,
	}
}
