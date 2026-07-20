// Package integration is the composition seam for outbound external integrations
// (SMS, email, Telebirr). It is stateless and exposes sender ports for other
// modules (e.g. notification) to consume.
package integration

import (
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/integration/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/integration/ports"
)

// Module wires the integration ACL into the monolith.
type Module struct {
	sender *adapters.LogSender
}

// New builds the module. The concrete sender is chosen here (logging for the
// prototype; real gateways in production).
func New(logger *slog.Logger) *Module {
	return &Module{sender: adapters.NewLogSender(logger)}
}

// Name identifies the module.
func (m *Module) Name() string { return "integration" }

// Init has no startup work.
func (m *Module) Init(_ *app.Kernel) error { return nil }

// Mount exposes no HTTP routes.
func (m *Module) Mount() (string, http.Handler) { return "", nil }

// SmsSender exposes the outbound SMS capability.
func (m *Module) SmsSender() ports.SmsSender { return m.sender }

// EmailSender exposes the outbound email capability.
func (m *Module) EmailSender() ports.EmailSender { return m.sender }
