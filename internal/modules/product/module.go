// Package product is the composition seam for the product-definition context. It
// implements app.Module, seeds the catalog on Init, and exposes a Catalog reader
// plus a RateProvider that satisfies the rating module's RateTableProvider port.
package product

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/adapters"
	productapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/rest"
	ratingports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Module wires the product context into the monolith.
type Module struct {
	svc      *productapp.Service
	provider *adapters.RateProvider
	logger   *slog.Logger
}

// New builds the module from the shared database.
func New(db *database.DB) *Module {
	repo := adapters.NewProductRepository(db)
	return &Module{
		svc:      productapp.NewService(repo),
		provider: adapters.NewRateProvider(repo),
	}
}

// Name identifies the module.
func (m *Module) Name() string { return "product" }

// Init captures the logger and seeds the Motor product (idempotent).
func (m *Module) Init(k *app.Kernel) error {
	m.logger = k.Logger
	return m.svc.Seed(context.Background(), adapters.MotorProduct())
}

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "product", rest.New(m.svc).Routes()
}

// RateProvider satisfies the rating module's RateTableProvider port (in-process
// cross-module wiring — replaces rating's static rate table).
func (m *Module) RateProvider() ratingports.RateTableProvider { return m.provider }
