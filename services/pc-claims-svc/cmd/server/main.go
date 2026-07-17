package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"medhen/pc-claims-svc/internal/application/commands"
	"medhen/pc-claims-svc/internal/infrastructure/ozone"
	"medhen/pc-claims-svc/internal/infrastructure/postgres"
	"medhen/pc-claims-svc/internal/presentation/rest"
)

func main() {
	log.Println("Starting pc-claims-svc...")

	// 1. Load Configuration
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://medhen:secret@localhost:5432/pc_claims_db?sslmode=disable"
	}

	// 2. Initialize Database Connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to database.")

	// 3. Initialize Domain, Application, and Infrastructure layers
	ozoneEndpoint := os.Getenv("OZONE_OM_URL")
	if ozoneEndpoint == "" {
		ozoneEndpoint = "http://localhost:9874"
	}
	ozoneClient, err := ozone.NewOzoneClient(ctx, ozoneEndpoint)
	if err != nil {
		log.Fatalf("Failed to initialize Ozone client: %v", err)
	}

	// Mocks for repo and services
	claimRepo := postgres.NewClaimRepository(pool)
	// policyService := grpc.NewPolicyServiceClient(...)
	// fraudService := grpc.NewFraudServiceClient(...)

	submitFNOL := commands.NewSubmitFNOLHandler(claimRepo, nil, nil) // Mocked upstream

	// 4. Setup HTTP Router (Presentation)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Mount domain routers
	claimHandler := rest.NewClaimHandler(submitFNOL, ozoneClient)
	r.Mount("/api/pc-claims/v1/claims", func() http.Handler {
		cr := chi.NewRouter()
		claimHandler.RegisterRoutes(cr)
		return cr
	}())

	// 5. Start Server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutDown()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
