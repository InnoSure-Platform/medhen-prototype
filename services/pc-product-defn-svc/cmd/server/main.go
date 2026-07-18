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
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"github.com/medhen/pc-product-defn-svc/config"
	"github.com/medhen/pc-product-defn-svc/internal/application/command"
	"github.com/medhen/pc-product-defn-svc/internal/infrastructure/kafka"
	"github.com/medhen/pc-product-defn-svc/internal/infrastructure/postgres"
	"github.com/medhen/pc-product-defn-svc/internal/presentation/rest"
	
	"github.com/medhen/pc-telemetry-sdk"
	"github.com/medhen/pc-auth-sdk"
	idempotency "github.com/medhen/pc-idempotency-mgmt-sdk"

	"net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	productpb "github.com/medhen/pc-contracts/gen/go/product/v1"
	productgrpc "github.com/medhen/pc-product-defn-svc/internal/presentation/grpc"
	"github.com/medhen/pc-product-defn-svc/internal/application/query"
)

func main() {
	// 1. Load Configuration (Viper)
	cfg, err := config.Load()
	if err != nil {
		// Can't use structured logger yet if it fails before init
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Initialize Telemetry (OpenTelemetry + slog)
	telCfg := telemetry.Config{
		ServiceName: cfg.ServiceName,
		Version:     cfg.Version,
		Endpoint:    cfg.Telemetry.Endpoint,
	}
	shutdownTelemetry, err := telemetry.Init(ctx, telCfg)
	if err != nil {
		slog.Error("Failed to initialize telemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownTelemetry(context.Background()); err != nil {
			slog.Error("Failed to shutdown telemetry", "error", err)
		}
	}()

	slog.Info("Starting service", "name", cfg.ServiceName, "version", cfg.Version)

	// 3. Initialize Postgres Pool with Tier-0 advanced config
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		slog.Error("Failed to parse database URL", "error", err)
		os.Exit(1)
	}
	poolCfg.MaxConns = 50
	poolCfg.MinConns = 10
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	dbPool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		slog.Error("Failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Verify DB connectivity
	if err := dbPool.Ping(ctx); err != nil {
		slog.Error("Postgres ping failed", "error", err)
		os.Exit(1)
	}

	// 4. Initialize Repositories and Infrastructure
	productRepo := postgres.NewProductRepository(dbPool)
	outboxPub := kafka.NewOutboxPublisher()

	// Initialize Kafka Relay
	// In production, broker list should come from cfg.Kafka.Brokers
	relay, err := kafka.NewOutboxRelay(dbPool, []string{"localhost:9092"})
	if err != nil {
		slog.Error("Failed to initialize Kafka relay", "error", err)
		os.Exit(1)
	}
	go relay.Start(ctx)

	// 5. Initialize Application Layer (Command Handlers)
	createProductCmd := command.NewCreateProductHandler(dbPool, productRepo, outboxPub)
	
	// 6. Initialize Presentation Layer (REST API)
	r := chi.NewRouter()
	
	// Enterprise Middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger) // Replaced with otelchi/otelslog in full implementation
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(5 * time.Second))

	// Health Probes
	r.Get("/health/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.Get("/health/readiness", func(w http.ResponseWriter, r *http.Request) {
		if err := dbPool.Ping(r.Context()); err != nil {
			http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	// Auth Middleware
	// In Tier-0, this secret comes from Vault or we fetch JWKS dynamically from Keycloak.
	authCfg := auth.ConfigFromEnv()
	r.Use(auth.Middleware(authCfg))

	// Idempotency SDK Integration
	idempCfg := idempotency.Config{RedisURL: "redis://localhost:6379"}
	idempMgr, err := idempotency.NewManager(idempCfg)
	var idempMiddleware func(http.Handler) http.Handler
	if err != nil {
		slog.Warn("Failed to connect to Redis for idempotency", "error", err)
		idempMiddleware = func(next http.Handler) http.Handler { return next }
	} else {
		idempMiddleware = idempMgr.Middleware
	}

	productHandler := rest.NewProductHandler(createProductCmd)
	productHandler.RegisterRoutes(r, idempMiddleware)

	// Initialize Query Handlers
	getProductQuery := query.NewGetProductHandler(productRepo)

	// Add TLS Certs
	certFile := "../../certs/server.crt"
	keyFile := "../../certs/server.key"

	// Start gRPC Server
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		slog.Error("Failed to load TLS keys for gRPC", "error", err)
		os.Exit(1)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	productpb.RegisterProductQueryServiceServer(grpcServer, productgrpc.NewServer(getProductQuery))
	
	go func() {
		lis, err := net.Listen("tcp", ":9091")
		if err != nil {
			slog.Error("Failed to listen for gRPC", "error", err)
			os.Exit(1)
		}
		slog.Info("gRPC server listening (TLS)", "addr", ":9091")
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
		}
	}()

	// 7. Start HTTP Server
	srv := &http.Server{
		Addr:    ":8080", // Ideally use cfg.Server.Port
		Handler: r,
	}

	go func() {
		slog.Info("HTTP server listening (HTTPS)", "addr", srv.Addr)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
		}
	}()

	// 8. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down gracefully...")
	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutDown()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}
	grpcServer.GracefulStop()
	slog.Info("Server exited")
}


