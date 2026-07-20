// Package policy is the composition seam for the policy bounded context — the
// keystone that wires rating, party and underwriting together and issues policies
// atomically.
package policy

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	partyports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/adapters"
	policyapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/rest"
	ratingports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	uwports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Module wires the policy context into the monolith.
type Module struct {
	svc    *policyapp.Service
	logger *slog.Logger
}

// New builds the module from the database and the ports of its collaborators.
func New(db *database.DB, rating ratingports.Calculator, party partyports.Reader, uw uwports.Decider) *Module {
	svc := policyapp.NewService(policyapp.Deps{
		DB:           db,
		Quotes:       adapters.NewQuoteRepository(db),
		Policies:     adapters.NewPolicyRepository(db),
		Rating:       rating,
		Party:        party,
		Underwriting: uw,
		Insurer:      "EIC",
	})
	return &Module{svc: svc}
}

// Name identifies the module.
func (m *Module) Name() string { return "policy" }

// Init captures platform dependencies.
func (m *Module) Init(k *app.Kernel) error { m.logger = k.Logger; return nil }

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "policy", rest.New(m.svc).Routes()
}

// Reader exposes policy lookups to billing and claims.
func (m *Module) Reader() ports.Reader { return reader{svc: m.svc} }

type reader struct{ svc *policyapp.Service }

func (r reader) GetPolicy(ctx context.Context, tenantID, id string) (ports.PolicyView, error) {
	p, err := r.svc.GetPolicy(ctx, tenantID, id)
	if err != nil {
		return ports.PolicyView{}, err
	}
	return ports.PolicyView{
		ID: p.ID, PolicyNumber: p.PolicyNumber, TenantID: p.TenantID, PartyID: p.PartyID,
		ProductCode: p.ProductCode, Status: string(p.Status), GrossMinor: p.GrossPremium.Minor(),
		EffectiveFrom: p.EffectiveFrom, EffectiveTo: p.EffectiveTo,
	}, nil
}
