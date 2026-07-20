// Package ports is the published contract of the policy module, consumed by
// billing and claims.
package ports

import (
	"context"
	"time"
)

// PolicyView is the public read model of a policy.
type PolicyView struct {
	ID            string    `json:"id"`
	PolicyNumber  string    `json:"policy_number"`
	TenantID      string    `json:"tenant_id"`
	PartyID       string    `json:"party_id"`
	ProductCode   string    `json:"product_code"`
	Status        string    `json:"status"`
	GrossMinor    int64     `json:"gross_minor"`
	EffectiveFrom time.Time `json:"effective_from"`
	EffectiveTo   time.Time `json:"effective_to"`
}

// Reader is the policy module's public capability.
type Reader interface {
	GetPolicy(ctx context.Context, tenantID, id string) (PolicyView, error)
}
