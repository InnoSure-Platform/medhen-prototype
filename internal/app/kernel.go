// Package app is the composition kernel of the modular monolith. It defines the
// Module contract that every bounded context implements and the Kernel of
// shared platform dependencies injected into modules at startup.
//
// Modules are sealed: a module may depend on another module only through its
// published port interface (passed in via wiring in cmd/medhen-api), never by
// importing another module's internal packages.
package app

import (
	"log/slog"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/config"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/eventbus"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
)

// Kernel holds the shared platform dependencies available to every module.
// It grows as platform capabilities land in Phase 2.
type Kernel struct {
	Config config.Config
	Logger *slog.Logger

	// Events is the in-process domain-event bus (outbox-backed).
	Events *eventbus.Bus
	// Auth validates access tokens; nil when Keycloak is not configured (dev).
	Auth *auth.Validator
	// Sequencer issues monotonic business numbers (e.g. policy numbers).
	Sequencer ids.Sequencer

	// Phase 2 additions still pending:
	//   DB      *database.Pool     (pgx pool + UnitOfWork)
	//   Outbox  *outbox.Writer     (transactional outbox)
	//   Idem    *idempotency.Store (Valkey SETNX)
}
