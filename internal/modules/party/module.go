// Package party is the composition seam for the party bounded context. It
// implements app.Module and exposes a Reader facade for other modules.
package party

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/adapters"
	partyapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/rest"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Module wires the party context into the monolith.
type Module struct {
	svc    *partyapp.Service
	logger *slog.Logger
}

// New builds the module from the shared database.
func New(db *database.DB) *Module {
	repo := adapters.NewPartyRepository(db)
	return &Module{svc: partyapp.NewService(db, repo)}
}

// Name identifies the module.
func (m *Module) Name() string { return "party" }

// Init captures platform dependencies.
func (m *Module) Init(k *app.Kernel) error {
	m.logger = k.Logger
	return nil
}

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "party", rest.New(m.svc, m.logger).Routes()
}

// Reader exposes party lookups to other modules (policy, claims).
func (m *Module) Reader() ports.Reader { return reader{svc: m.svc} }

type reader struct{ svc *partyapp.Service }

func (r reader) GetParty(ctx context.Context, tenantID, id string) (ports.PartyView, error) {
	p, err := r.svc.Get(ctx, tenantID, id)
	if err != nil {
		return ports.PartyView{}, err
	}
	return ports.PartyView{
		ID: p.ID, TenantID: p.TenantID, Type: string(p.Type), Status: string(p.Status),
		FullName: p.FullName, FullNameAmharic: p.FullNameAmharic, PhoneE164: p.PhoneE164,
	}, nil
}
