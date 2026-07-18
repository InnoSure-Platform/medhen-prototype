package adapters

import (
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/shopspring/decimal"
)

// MotorProduct returns the seed definition for the EIC Motor product: own-damage
// and third-party-liability coverages with an AGE factor on own damage.
func MotorProduct() *domain.Product {
	return &domain.Product{
		Code: "MOT", LOB: "MOTOR", Name: "Motor Insurance",
		NameAmharic: "የተሽከርካሪ መድን", Status: domain.StatusActive, RateVersion: "MOTOR-2026.1",
		Coverages: []domain.Coverage{
			{Code: "OD", Name: "Own Damage", NameAmharic: "የራስ ጉዳት", BaseRate: money.FromInt(1200)},
			{Code: "TPL", Name: "Third-Party Liability", NameAmharic: "የሶስተኛ ወገን ኃላፊነት", BaseRate: money.FromInt(800)},
		},
		Factors: []domain.Factor{
			{CoverageCode: "OD", FactorType: "AGE", Dimension: "young", Value: decimal.NewFromFloat(1.25)},
			{CoverageCode: "OD", FactorType: "AGE", Dimension: "adult", Value: decimal.NewFromFloat(1.00)},
			{CoverageCode: "OD", FactorType: "AGE", Dimension: "senior", Value: decimal.NewFromFloat(1.10)},
		},
	}
}
