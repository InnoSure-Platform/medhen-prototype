// Package rating is the composition seam for the rating bounded context. It
// implements app.Module and exposes a Calculator facade for other modules to
// consume in-process.
package rating

import (
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/rest"
)

// Module wires the rating engine into the monolith.
type Module struct {
	engine *domain.Engine
	logger *slog.Logger
}

// New constructs the rating module from its dependencies (a rate-table provider
// and tax policy). The engine is built eagerly so Calculator() is usable by
// other modules during composition, before Init runs.
func New(provider ports.RateTableProvider, tax domain.TaxPolicy) *Module {
	return &Module{engine: domain.NewEngine(provider, tax)}
}

// Name identifies the module.
func (m *Module) Name() string { return "rating" }

// Init captures platform dependencies.
func (m *Module) Init(k *app.Kernel) error {
	m.logger = k.Logger
	return nil
}

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "rating", rest.New(m.engine, m.logger).Routes()
}

// Calculator exposes the rating capability to other modules (e.g. policy),
// replacing the pre-refactor stubbed gRPC client.
func (m *Module) Calculator() ports.Calculator { return m.engine }
