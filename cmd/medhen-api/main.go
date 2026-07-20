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
	"strings"
	"syscall"
	"time"

	"github.com/shopspring/decimal"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/audit"
	auditadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/audit/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing"
	billingadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims"
	claimsadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/document"
	documentadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam"
	iamadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/integration"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification"
	notificationadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party"
	partyadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/adapters"
	partydomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy"
	policyadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/adapters"
	policydomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product"
	productadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating"
	ratingadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/adapters"
	ratingdomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/domain"
	ratingports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/reporting"
	reportingadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/reporting/adapters"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting"
	uwdomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/config"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/eventbus"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/httpx"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/migrate"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
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
		// M10: never run without TLS to Postgres outside local dev.
		if cfg.Env != "dev" && strings.Contains(cfg.DatabaseURL, "sslmode=disable") {
			logger.Warn("DATABASE_URL uses sslmode=disable in a non-dev environment — enable TLS (sslmode=require)")
		}
		db, err := database.Connect(relayCtx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer db.Close()
		if err := runMigrations(relayCtx, db, logger); err != nil {
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

	// Start module background loops (e.g. notification dispatcher). They stop when
	// relayCtx is cancelled at shutdown.
	registry.StartBackground(relayCtx, kernel)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", healthHandler)
	registry.MountAll(mux)

	// Edge authentication + server-side RBAC. Public paths bypass auth (health +
	// the HMAC-authenticated Telebirr webhook). When Keycloak is not configured
	// the middleware is a pass-through (dev), and handlers read X-Tenant-ID.
	public := []string{"/healthz", "/readyz", "/billing/webhooks/telebirr"}
	rbac := []auth.AccessRule{
		{Prefix: "/iam/", AnyOf: []string{"admin"}},
		{Prefix: "/audit/", AnyOf: []string{"staff", "admin"}},
		{Prefix: "/claims/", AnyOf: []string{"claims", "staff", "admin"}},
		{Prefix: "/billing/", AnyOf: []string{"finance", "staff", "admin"}},
		{Prefix: "/policy/", AnyOf: []string{"agent", "staff", "admin"}},
		{Prefix: "/party/", AnyOf: []string{"agent", "staff", "admin"}},
		// /product, /reporting, /document require only a valid token.
	}

	handler := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(logger),
		auth.EdgeMiddleware(validator, public, rbac),
	)

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

	// underwriting — stateless STP decision engine.
	uwMod := underwriting.New(uwdomain.Rules{
		ReferAbove:     money.FromInt(100000),
		MaxPriorClaims: 1,
	})

	// integration — stateless outbound ACL (SMS/email/Telebirr); consumed by
	// notification.
	integrationMod := integration.New(k.Logger)

	// rating's pricing source: the product catalog when a DB is available,
	// otherwise a static Motor table so the stateless path still works.
	var rateProvider ratingports.RateTableProvider = ratingadapters.NewMotorRateTable()
	modules := []app.Module{uwMod, integrationMod}

	if k.DB != nil {
		// audit — registered first so its SubscribeAll captures every event into
		// the immutable trail.
		modules = append(modules, audit.New(k.DB))

		// product — DB-backed catalog; supplies rating's RateTableProvider.
		productMod := product.New(k.DB)
		rateProvider = productMod.RateProvider()
		modules = append(modules, productMod)
	}

	// rating — consumes the rate provider (cross-module port when DB is present).
	ratingMod := rating.New(rateProvider, taxPolicy)
	modules = append(modules, ratingMod)

	if k.DB != nil {
		// party — DB-backed, emits events via the outbox.
		partyMod := party.New(k.DB)
		// Demo subscribers: audit trail (the audit module will own these).
		k.Events.Subscribe(partydomain.TopicPartyRegistered, logEvent(k))
		k.Events.Subscribe(policydomain.TopicPolicyIssued, logEvent(k))
		modules = append(modules, partyMod)

		// policy — the keystone: wires rating.Calculator + party.Reader +
		// underwriting.Decider and issues atomically via the UoW + outbox.
		policyMod := policy.New(k.DB, ratingMod.Calculator(), partyMod.Reader(), uwMod.Decider())
		modules = append(modules, policyMod)

		// billing — subscribes to policy.issued to raise the first invoice, and
		// applies Telebirr payments (HMAC-verified).
		billingMod := billing.New(k.DB, k.Config.TelebirrWebhookSecret)
		modules = append(modules, billingMod)

		// claims — FNOL (validates cover via policy.Reader) → fast-track settle.
		claimsMod := claims.New(k.DB, policyMod.Reader(), money.FromInt(50000))
		modules = append(modules, claimsMod)

		// document — generates the Certificate of Insurance on policy.issued.
		modules = append(modules, document.New(k.DB, partyMod.Reader()))

		// notification — queues SMS on policy.issued/claims.settled (recipient via
		// party.Reader) and dispatches them via integration in a background loop.
		modules = append(modules, notification.New(k.DB, partyMod.Reader(), integrationMod.SmsSender(), k.Logger))

		// reporting — projects policy.issued/claims.settled into real KPIs.
		modules = append(modules, reporting.New(k.DB))

		// iam — application user/role management (auth kernel lives in platform).
		modules = append(modules, iam.New(k.DB))
	}

	return app.NewRegistry(modules...)
}

func logEvent(k *app.Kernel) eventbus.Handler {
	return func(_ context.Context, e eventbus.Event) error {
		k.Logger.Info("event", "topic", e.EventName())
		return nil
	}
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

// runMigrations builds the ordered migration set from each module's own DDL (the
// single source of truth) plus a final security migration (least-privilege role
// + row-level security), and applies it transactionally. Idempotent: already-
// applied migrations are skipped, so this runs safely at every boot.
func runMigrations(ctx context.Context, db *database.DB, logger *slog.Logger) error {
	migrations := []migrate.Migration{
		{Version: 1, Name: "platform_outbox", SQL: outbox.Schema},
		{Version: 2, Name: "audit", SQL: auditadapters.Schema},
		{Version: 3, Name: "party", SQL: partyadapters.Schema},
		{Version: 4, Name: "product", SQL: productadapters.Schema},
		{Version: 5, Name: "policy", SQL: policyadapters.Schema},
		{Version: 6, Name: "billing", SQL: billingadapters.Schema},
		{Version: 7, Name: "claims", SQL: claimsadapters.Schema},
		{Version: 8, Name: "document", SQL: documentadapters.Schema},
		{Version: 9, Name: "notification", SQL: notificationadapters.Schema},
		{Version: 10, Name: "reporting", SQL: reportingadapters.Schema},
		{Version: 11, Name: "iam", SQL: iamadapters.Schema},
		{Version: 100, Name: "security_roles_rls", SQL: securityMigration},
	}
	n, err := migrate.Apply(ctx, db.Pool(), migrations)
	if err != nil {
		return err
	}
	logger.Info("schema migrations applied", "count", n)
	return nil
}

// securityMigration provisions a least-privilege application role (no GRANT ALL,
// no DDL) and enables row-level security on every tenant-scoped table. The table
// owner/superuser bypasses RLS (used for migrations/admin); the app role does
// not, so connecting the pool as medhen_app enforces tenant isolation at the DB.
// Enforcement is active per transaction via database.WithinTenantTx, which sets
// app.current_tenant.
const securityMigration = `
DO $$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'medhen_app') THEN
    CREATE ROLE medhen_app LOGIN PASSWORD 'medhen_app';
  END IF;
END $$;

GRANT USAGE ON SCHEMA public TO medhen_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO medhen_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO medhen_app;

DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY[
    'parties','quotes','policies','invoices','payments','claims',
    'documents','notifications','audit_log','reporting_kpis','iam_users'
  ] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
    EXECUTE format($f$CREATE POLICY tenant_isolation ON %I
        USING (tenant_id = current_setting('app.current_tenant', true))
        WITH CHECK (tenant_id = current_setting('app.current_tenant', true))$f$, t);
  END LOOP;
END $$;
`

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
