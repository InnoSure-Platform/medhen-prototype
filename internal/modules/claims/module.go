// Package claims is the composition seam for the claims bounded context. It
// consumes the policy module's Reader to validate cover at FNOL time.
package claims

import (
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/adapters"
	claimsapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/rest"
	policyports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// Module wires the claims context into the monolith.
type Module struct {
	svc    *claimsapp.Service
	logger *slog.Logger
}

// New builds the module from the database, the policy Reader, and the fast-track
// authority limit.
func New(db *database.DB, policyReader policyports.Reader, fastTrackLimit money.Amount) *Module {
	svc := claimsapp.NewService(claimsapp.Deps{
		DB:             db,
		Claims:         adapters.NewClaimRepository(db),
		Policy:         policyReader,
		FastTrackLimit: fastTrackLimit,
	})
	return &Module{svc: svc}
}

// Name identifies the module.
func (m *Module) Name() string { return "claims" }

// Init captures platform dependencies.
func (m *Module) Init(k *app.Kernel) error { m.logger = k.Logger; return nil }

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "claims", rest.New(m.svc).Routes()
}
