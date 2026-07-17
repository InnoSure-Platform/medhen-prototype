package preference

import "github.com/google/uuid"

type RoutingPreference struct {
	PartyID        uuid.UUID
	OptedOutSMS    bool
	OptedOutEmail  bool
	OptedOutInApp  bool
	MarketingOptIn bool
}

func (p *RoutingPreference) IsOptedOut(channel string, category string) bool {
	if category == "STATUTORY" {
		return false // Statutory overrides opt-outs
	}
	if category == "MARKETING" && !p.MarketingOptIn {
		return true
	}
	switch channel {
	case "SMS":
		return p.OptedOutSMS
	case "EMAIL":
		return p.OptedOutEmail
	case "IN_APP":
		return p.OptedOutInApp
	}
	return false
}
