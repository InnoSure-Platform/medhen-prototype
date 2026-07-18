// Command medhen-api is the modular-monolith entrypoint: a single process that
// wires the platform kernel and all bounded-context modules behind one HTTP
// edge. Modules are registered in composeModules; they migrate in during
// Phase 3 of docs/refactor/modular-monolith-plan.md.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party"
	partyadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/adapters"
	partydomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product"
	productadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating"
	ratingadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/adapters"
	ratingdomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/domain"
	ratingports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/config"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/eventbus"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/httpx"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
	"github.com/shopspring/decimal"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := run(logger); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()

	// Auth is enabled when Keycloak is configured; otherwise the process runs
	// with only public routes (dev). It never falls back to an insecure mode.
	var validator *auth.Validator
	if authCfg := auth.ConfigFromEnv(); authCfg.IssuerURL != "" {
		v, err := auth.NewValidator(authCfg)
		if err != nil {
			return err
		}
		validator = v
		logger.Info("authentication enabled", "issuer", authCfg.IssuerURL)
	} else {
		logger.Warn("authentication DISABLED: set KEYCLOAK_URL and KEYCLOAK_REALM to enable")
	}

	kernel := &app.Kernel{
		Config:    cfg,
		Logger:    logger,
		Events:    eventbus.New(logger),
		Auth:      validator,
		Sequencer: ids.NewInMemorySequencer(),
	}

	// Database + outbox relay are enabled when DATABASE_URL is set. Stateless
	// modules (rating) run either way.
	relayCtx, stopRelay := context.WithCancel(context.Background())
	defer stopRelay()
	if cfg.DatabaseURL != "" {
		db, err := database.Connect(relayCtx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer db.Close()
		if err := applySchemas(relayCtx, db); err != nil {
			return err
		}
		kernel.DB = db

		// Bridge the transactional outbox to the in-process event bus.
		relay := outbox.NewRelay(db, busPublisher(kernel.Events), 100, logger)
		go relay.Run(relayCtx, cfg.OutboxPollInterval)
		logger.Info("database + outbox relay enabled")
	} else {
		logger.Warn("DATABASE_URL not set: DB-backed modules disabled")
	}

	registry := composeModules(kernel)
	if err := registry.InitAll(kernel); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", healthHandler)
	registry.MountAll(mux)

	handler := httpx.Chain(mux, httpx.RequestID, httpx.Recover(logger))

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("medhen-api listening",
			"addr", cfg.HTTPAddr, "env", cfg.Env, "modules", registry.Names())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// composeModules is the single place modules are registered, in dependency
// order. It is empty until Phase 3 migrates the bounded contexts in, e.g.:
//
//	party := partymod.New(...)
//	rating := ratingmod.New(...)
//	policy := policymod.New(rating.Calculator(), underwriting.Decider(), ...)
//	return app.NewRegistry(iam, party, product, rating, underwriting, policy, ...)
func composeModules(k *app.Kernel) *app.Registry {
	taxPolicy := ratingdomain.TaxPolicy{
		VATRate:   decimal.NewFromFloat(0.15), // Ethiopia standard VAT
		StampDuty: money.FromInt(35),          // fixed stamp duty (placeholder → product config)
	}

	// rating's pricing source: the product catalog when a DB is available,
	// otherwise a static Motor table so the stateless path still works.
	var rateProvider ratingports.RateTableProvider = ratingadapters.NewMotorRateTable()
	var modules []app.Module

	if k.DB != nil {
		// product — DB-backed catalog; supplies rating's RateTableProvider.
		productMod := product.New(k.DB)
		rateProvider = productMod.RateProvider()
		modules = append(modules, productMod)
	}

	// rating — consumes the rate provider (cross-module port when DB is present).
	modules = append(modules, rating.New(rateProvider, taxPolicy))

	if k.DB != nil {
		// party — DB-backed, emits events via the outbox.
		partyMod := party.New(k.DB)
		// Demo subscriber: audit trail for party registrations (the audit module
		// will own this once migrated).
		k.Events.Subscribe(partydomain.TopicPartyRegistered, func(_ context.Context, e eventbus.Event) error {
			k.Logger.Info("event", "topic", e.EventName())
			return nil
		})
		modules = append(modules, partyMod)
	}

	return app.NewRegistry(modules...)
}

// busPublisher adapts the outbox relay to the in-process event bus. Each outbox
// message is delivered as a generic event keyed by its topic; subscribers decode
// the payload. This is the same seam that would target Kafka after extraction.
func busPublisher(bus *eventbus.Bus) outbox.Publisher {
	return outbox.PublisherFunc(func(ctx context.Context, m outbox.Message) error {
		return bus.Publish(ctx, relayedEvent{topic: m.Topic, payload: m.Payload})
	})
}

type relayedEvent struct {
	topic   string
	payload []byte
}

func (e relayedEvent) EventName() string { return e.topic }
func (e relayedEvent) Payload() []byte   { return e.payload }

// applySchemas creates the platform + module tables. This is a stopgap until the
// migration tool lands in Phase 5.
func applySchemas(ctx context.Context, db *database.DB) error {
	for _, ddl := range []string{outbox.Schema, partyadapters.Schema, productadapters.Schema} {
		if _, err := db.Pool().Exec(ctx, ddl); err != nil {
			return err
		}
	}
	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
