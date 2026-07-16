package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-product-defn-svc/internal/application/command"
	"github.com/medhen/pc-product-defn-svc/internal/infrastructure/kafka"
	"github.com/medhen/pc-product-defn-svc/internal/infrastructure/postgres"
	"github.com/medhen/pc-product-defn-svc/internal/presentation/rest"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Postgres Pool
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/medhen_product?sslmode=disable"
	}
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("Failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Initialize Repositories and Infrastructure
	productRepo := postgres.NewProductRepository(dbPool)
	outboxPub := kafka.NewOutboxPublisher()

	// Initialize Application Layer (Command Handlers)
	createProductCmd := command.NewCreateProductHandler(dbPool, productRepo, outboxPub)
	// transitionProductCmd := command.NewTransitionProductHandler(dbPool, productRepo, outboxPub)

	// Initialize Presentation Layer (REST API)
	r := chi.NewRouter()
	productHandler := rest.NewProductHandler(createProductCmd)
	productHandler.RegisterRoutes(r)

	// Start HTTP Server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		slog.Info("Starting HTTP server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down gracefully...")
	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutDown()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}

	slog.Info("Server exited")
}
