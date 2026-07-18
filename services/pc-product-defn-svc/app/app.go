package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-product-defn-svc/internal/application/command"
	"github.com/medhen/pc-product-defn-svc/internal/infrastructure/kafka"
	"github.com/medhen/pc-product-defn-svc/internal/infrastructure/postgres"
	"github.com/medhen/pc-product-defn-svc/internal/presentation/rest"
)

// NewTestHandler returns an http.Handler wired with all real internal components
// for integration testing. The caller supplies the authentication middleware so
// tests can inject a validator backed by a locally generated signing key.
func NewTestHandler(dbPool *pgxpool.Pool, authMW func(http.Handler) http.Handler) http.Handler {
	repo := postgres.NewProductRepository(dbPool)
	outboxPub := kafka.NewOutboxPublisher()
	createCmd := command.NewCreateProductHandler(dbPool, repo, outboxPub)

	r := chi.NewRouter()

	if authMW != nil {
		r.Use(authMW)
	}

	handler := rest.NewProductHandler(createCmd)
	handler.RegisterRoutes(r, nil)

	return r
}
