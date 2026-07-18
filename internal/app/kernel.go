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

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/config"
)

// Kernel holds the shared platform dependencies available to every module.
// It grows as platform capabilities land in Phase 2 (database/UoW, event bus,
// auth validator, outbox, idempotency, telemetry).
type Kernel struct {
	Config config.Config
	Logger *slog.Logger

	// Phase 2 additions (placeholders documented for the migration):
	//   DB      *database.Pool
	//   Events  *eventbus.Bus
	//   Auth    *auth.Validator
	//   Outbox  *outbox.Writer
}
