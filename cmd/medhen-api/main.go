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
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/config"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/eventbus"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/httpx"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
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

	registry := composeModules()
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
func composeModules() *app.Registry {
	return app.NewRegistry()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
