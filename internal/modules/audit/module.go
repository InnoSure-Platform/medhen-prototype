// Package audit is the composition seam for the audit bounded context. On Init it
// subscribes to ALL events and records an immutable trail entry for each — making
// "audit on every state change" real rather than aspirational.
package audit

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/audit/adapters"
	auditapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/audit/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/audit/rest"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/eventbus"
)

// Module wires the audit context into the monolith.
type Module struct {
	rec    *auditapp.Recorder
	logger *slog.Logger
}

// New builds the module from the shared database.
func New(db *database.DB) *Module {
	return &Module{rec: auditapp.NewRecorder(adapters.NewAuditRepository(db))}
}

// Name identifies the module.
func (m *Module) Name() string { return "audit" }

// Init subscribes to every event so all state changes are recorded. The handler
// runs inside the outbox-relay transaction, so the audit row commits atomically
// with the event being marked processed.
func (m *Module) Init(k *app.Kernel) error {
	m.logger = k.Logger
	k.Events.SubscribeAll(func(ctx context.Context, e eventbus.Event) error {
		payloaded, ok := e.(interface{ Payload() []byte })
		if !ok {
			return m.rec.Record(ctx, e.EventName(), []byte("{}"))
		}
		return m.rec.Record(ctx, e.EventName(), payloaded.Payload())
	})
	return nil
}

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "audit", rest.New(m.rec).Routes()
}
