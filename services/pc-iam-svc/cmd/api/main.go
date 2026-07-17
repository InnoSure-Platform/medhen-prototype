package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	iamv1 "github.com/medhen/pc-contracts/gen/go/iam/v1"
	grpcapi "github.com/medhen/pc-iam-svc/internal/api/grpc"
	rest "github.com/medhen/pc-iam-svc/internal/api/rest"
	"github.com/medhen/pc-iam-svc/internal/application"
	"github.com/medhen/pc-iam-svc/internal/infrastructure/db"
	"github.com/medhen/pc-iam-svc/internal/infrastructure/keycloak"
	"github.com/medhen/pc-iam-svc/internal/infrastructure/opa"

	auth "github.com/medhen/pc-auth-sdk"
	idempotency "github.com/medhen/pc-idempotency-mgmt-sdk"
	telemetry "github.com/medhen/pc-telemetry-sdk"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize Telemetry
	telCfg := telemetry.Config{
		ServiceName: "pc-iam-svc",
		Version:     "v1.0.0",
		Endpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}
	shutdownTelemetry, err := telemetry.Init(ctx, telCfg)
	if err != nil {
		slog.Error("Failed to initialize telemetry", "error", err)
	} else {
		defer func() {
			if err := shutdownTelemetry(context.Background()); err != nil {
				slog.Error("Failed to shutdown telemetry", "error", err)
			}
		}()
	}

	slog.Info("Starting pc-iam-svc")

	// 2. Initialize Infrastructure
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://medhen:medhen@localhost:5432/medhen?sslmode=disable"
	}
	repo, err := db.NewPostgresRepository(ctx, dbURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	keycloakClient := keycloak.NewClient(
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("KEYCLOAK_USER"),
		os.Getenv("KEYCLOAK_PASSWORD"),
	)
	
	opaClient := opa.NewClient(os.Getenv("OPA_URL"))

	// Initialize Application Use Cases
	tenantUseCase := application.NewTenantUseCase(repo, keycloakClient)
	userUseCase := application.NewUserUseCase(repo, keycloakClient)
	authzUseCase := application.NewAuthzUseCase(repo, opaClient)

	// Initialize REST Server (Control Plane)
	r := chi.NewRouter()
	
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(5 * time.Second))
	r.Use(telemetry.Middleware("pc-iam-svc"))

	r.Get("/health/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	r.Get("/health/readiness", func(w http.ResponseWriter, r *http.Request) {
		if err := repo.Ping(r.Context()); err != nil {
			http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	// Auth SDK Integration
	authCfg := auth.JWTConfig{SecretKey: "dev-secret-key"}
	authMiddleware := auth.Middleware(authCfg)

	// Idempotency SDK Integration
	idempCfg := idempotency.Config{RedisURL: "redis://localhost:6379"}
	idempMgr, err := idempotency.NewManager(idempCfg)
	var idempMiddleware func(http.Handler) http.Handler
	if err != nil {
		slog.Warn("Failed to connect to Redis for idempotency. Continuing without idempotency enforcement", "error", err)
		idempMiddleware = func(next http.Handler) http.Handler { return next }
	} else {
		idempMiddleware = idempMgr.Middleware
	}

	restHandler := rest.NewHandler(tenantUseCase, userUseCase)
	scimHandler := rest.NewSCIMHandler()

	r.Route("/api", func(r chi.Router) {
		r.Use(authMiddleware)
		// For provisioning routes, wrap with Idempotency
		r.With(idempMiddleware).Group(func(r chi.Router) {
			restHandler.RegisterRoutes(r)
			scimHandler.RegisterRoutes(r)
		})
	})

	// Add TLS for HTTP
	certFile := "../../certs/server.crt"
	keyFile := "../../certs/server.key"
	
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		slog.Info("Starting REST Server on :8080 (HTTPS)")
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			slog.Error("REST server failed", "error", err)
		}
	}()

	// Initialize gRPC Server (Data Plane)
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		slog.Error("Failed to load TLS keys for gRPC", "error", err)
		os.Exit(1)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	iamv1.RegisterPolicyEnforcementServiceServer(grpcServer, grpcapi.NewHandler(authzUseCase))

	go func() {
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			slog.Error("failed to listen on :9090", "error", err)
		}
		slog.Info("Starting gRPC Server on :9090 (TLS)")
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	slog.Info("Shutting down servers...")
	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutDown()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}
	grpcServer.GracefulStop()
	
	slog.Info("Servers stopped.")
}
