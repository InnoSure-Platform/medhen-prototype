// Package iam is the composition seam for the IAM context.
package iam

import (
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/adapters"
	iamapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/rest"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Module wires the IAM context into the monolith.
type Module struct {
	svc    *iamapp.Service
	logger *slog.Logger
}

// New builds the module from the shared database.
func New(db *database.DB) *Module {
	return &Module{svc: iamapp.NewService(adapters.NewUserRepository(db))}
}

// Name identifies the module.
func (m *Module) Name() string { return "iam" }

// Init captures platform dependencies.
func (m *Module) Init(k *app.Kernel) error { m.logger = k.Logger; return nil }

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "iam", rest.New(m.svc).Routes()
}

// Reader exposes role resolution to other modules.
func (m *Module) Reader() ports.Reader { return m.svc }
