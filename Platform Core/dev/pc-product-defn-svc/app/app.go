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

// NewTestHandler returns an http.Handler wired with all real internal components for integration testing.
func NewTestHandler(dbPool *pgxpool.Pool) http.Handler {
	repo := postgres.NewProductRepository(dbPool)
	outboxPub := kafka.NewOutboxPublisher()
	createCmd := command.NewCreateProductHandler(dbPool, repo, outboxPub)
	
	r := chi.NewRouter()
	handler := rest.NewProductHandler(createCmd)
	handler.RegisterRoutes(r)
	
	return r
}
