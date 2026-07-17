package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medhen/pc-policy-svc/config"
	"github.com/medhen/pc-policy-svc/internal/application/command"
	"github.com/medhen/pc-policy-svc/internal/application/query"
	"github.com/medhen/pc-policy-svc/internal/infrastructure/grpc_client"
	"github.com/medhen/pc-policy-svc/internal/infrastructure/repository"
	"github.com/medhen/pc-policy-svc/internal/presentation/rest"
	
	// Simulated import for pc-telemetry-sdk
	// "github.com/medhen/pc-telemetry-sdk/tracing"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize OpenTelemetry Tracing
	// tracerProvider, err := tracing.InitProvider(ctx, cfg.App.Name, cfg.App.Env)
	// if err != nil { log.Fatalf("failed to initialize tracing: %v", err) }
	// defer tracerProvider.Shutdown(ctx)

	// Initialize Database Pool
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName, cfg.Postgres.SSLMode)
	
	var pool *pgxpool.Pool
	// Fallback to simple stubbed repositories if no DB config is provided (for local testing without dependencies)
	if cfg.Postgres.Host != "" {
		pool, err = pgxpool.New(ctx, connString)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v", err)
		}
		defer pool.Close()
	}

	// Setup Repositories
	// If pool is nil (e.g. running locally without setting up DB config yet), these will panic on use, but it's enough for scaffolding.
	quoteRepo := repository.NewPostgresQuoteRepository(pool)
	policyRepo := repository.NewPostgresPolicyRepository(pool)

	// Setup Clients
	uwClient, err := grpc_client.NewUWClient("localhost:50053") // Assumed UW address
	if err != nil {
		log.Printf("Warning: Failed to create UW client: %v", err)
	}

	// Setup Command Handlers
	createQuoteHandler := command.NewCreateQuoteHandler(quoteRepo, uwClient)
	bindPolicyHandler := command.NewBindPolicyHandler(quoteRepo, policyRepo)

	// Setup Query Handlers
	getTimelineHandler := query.NewGetTimelineHandler(pool)
	searchPoliciesHandler := query.NewSearchPoliciesHandler(pool)

	// Setup Presentation (HTTP REST)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	handler := rest.NewHandler(createQuoteHandler, bindPolicyHandler, getTimelineHandler, searchPoliciesHandler)
	handler.RegisterRoutes(r)

	// Start HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.HTTPPort),
		Handler: r,
	}

	go func() {
		log.Printf("Starting HTTP server on port %d", cfg.App.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
