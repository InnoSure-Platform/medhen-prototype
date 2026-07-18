// Package ports is the published contract of the party module. Other modules
// (e.g. policy, claims) depend only on this package — never on party's domain,
// app, or adapters — and receive a Reader via the composition root.
package ports

import "context"

// PartyView is the public read model of a party for cross-module consumers.
type PartyView struct {
	ID              string `json:"id"`
	TenantID        string `json:"tenant_id"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	FullName        string `json:"full_name"`
	FullNameAmharic string `json:"full_name_amharic"`
	PhoneE164       string `json:"phone_e164"`
}

// Reader is the party module's public capability: look up a party by id within
// a tenant. Consumed in-process by other modules.
type Reader interface {
	GetParty(ctx context.Context, tenantID, id string) (PartyView, error)
}
