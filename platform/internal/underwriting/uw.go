package underwriting

import pcerr "github.com/InnoSure-Platform/pc-shared-go/errors"

// Decision is an STP underwriting outcome.
type Decision struct {
	Outcome string // ACCEPT | DECLINE | REFER
	Reason  string
}

// EvaluateSTP applies Phase 0 straight-through rules for standard motor risks.
func EvaluateSTP(year int, sumInsuredMinor int64, coverType string) Decision {
	if year < 2010 {
		return Decision{Outcome: "DECLINE", Reason: "vehicle older than underwriting threshold"}
	}
	if sumInsuredMinor > 500_000_000 { // 5,000,000.00 ETB
		return Decision{Outcome: "REFER", Reason: "sum insured exceeds STP authority"}
	}
	if coverType != "comprehensive" && coverType != "third_party" {
		return Decision{Outcome: "DECLINE", Reason: "unsupported cover type"}
	}
	return Decision{Outcome: "ACCEPT", Reason: "standard risk — STP"}
}

func RequireAccept(d Decision) error {
	if d.Outcome != "ACCEPT" {
		return pcerr.E(pcerr.CodeUWDeclined, d.Reason)
	}
	return nil
}
