// Package underwriting is the composition seam for the underwriting context.
package underwriting

import (
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting/ports"
)

// Module wires the underwriting engine into the monolith.
type Module struct {
	engine *domain.Engine
}

// New builds the module from its rules.
func New(rules domain.Rules) *Module {
	return &Module{engine: domain.NewEngine(rules)}
}

// Name identifies the module.
func (m *Module) Name() string { return "underwriting" }

// Init has no startup work.
func (m *Module) Init(_ *app.Kernel) error { return nil }

// Mount returns no HTTP routes yet (the workbench UI migrates later).
func (m *Module) Mount() (string, http.Handler) { return "", nil }

// Decider exposes the STP decision capability to the policy module.
func (m *Module) Decider() ports.Decider { return m.engine }
