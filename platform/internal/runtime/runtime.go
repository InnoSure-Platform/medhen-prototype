// Package runtime bootstraps Medhen microservices (Postgres, Kafka, HTTP).
package runtime

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-shared-go/auth"
	"github.com/InnoSure-Platform/pc-shared-go/httpx"
	"github.com/InnoSure-Platform/pc-shared-go/kafka"
	mw "github.com/InnoSure-Platform/pc-shared-go/middleware"
)

// Config for a Medhen service process.
type Config struct {
	Name string
	Addr string
}

func Env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func OpenStore(ctx context.Context) store.Repository {
	repo, err := store.OpenRepository(ctx)
	if err != nil {
		slog.Error("repository", "err", err)
		os.Exit(1)
	}
	return repo
}

func OpenKafka(ctx context.Context, repo store.Repository) *kafka.Publisher {
	pub := kafka.NewPublisherFromEnv()
	if pub == nil {
		return nil
	}
	pg, ok := repo.(*store.PostgresRepository)
	if !ok {
		return pub
	}
	go kafka.RelayOutbox(ctx, pub,
		func(ctx context.Context, limit int) ([]kafka.OutboxRow, error) {
			rows, err := pg.FetchOutbox(ctx, limit)
			if err != nil {
				return nil, err
			}
			out := make([]kafka.OutboxRow, len(rows))
			for i, r := range rows {
				out[i] = kafka.OutboxRow{ID: r.ID, AggregateType: r.AggregateType, AggregateID: r.AggregateID, EventType: r.EventType, Payload: r.Payload}
			}
			return out, nil
		},
		func(ctx context.Context, id string) error { return pg.MarkOutboxPublished(ctx, id) },
		500*time.Millisecond,
	)
	return pub
}

// BaseRouter returns chi with standard middleware + optional Keycloak JWT.
func BaseRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(httpx.CORS)
	r.Use(httpx.RequestID)
	r.Use(mw.Recover)
	r.Use(mw.Logging)
	r.Use(auth.Middleware(auth.NewValidatorFromEnv()))
	r.Use(mw.DemoAuth) // fills X-User-ID when JWT absent (demo)
	return r
}

func Listen(name, addr string, h http.Handler) {
	slog.Info("service listening", "name", name, "addr", addr)
	srv := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("listen failed", "name", name, "err", err)
		os.Exit(1)
	}
}

func Health(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, 200, map[string]string{"status": "ok", "service": name})
	}
}

// EmitOutbox writes audit + outbox when repo supports it.
func EmitOutbox(ctx context.Context, repo store.Repository, aggregateType, aggregateID, eventType string, payload any) {
	_ = repo.InsertOutbox(ctx, aggregateType, aggregateID, eventType, payload)
}

func Audit(ctx context.Context, repo store.Repository, entityType, entityID, action, actor, detail string) {
	_ = repo.AppendAudit(ctx, store.NewAuditEntry(entityType, entityID, action, actor, detail))
}
