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

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"

	"github.com/medhen/pc-reporting-svc/config"
	"github.com/medhen/pc-reporting-svc/internal/application/extractor"
	"github.com/medhen/pc-reporting-svc/internal/application/job"
	"github.com/medhen/pc-reporting-svc/internal/application/query"
	clickhouse_repo "github.com/medhen/pc-reporting-svc/internal/infrastructure/clickhouse"
	reportinggrpc "github.com/medhen/pc-reporting-svc/internal/presentation/grpc"
	"github.com/medhen/pc-reporting-svc/internal/presentation/rest"

	"github.com/medhen/pc-auth-sdk"
	reportingpb "github.com/medhen/pc-contracts/gen/go/reporting/v1"
	idempotency "github.com/medhen/pc-idempotency-mgmt-sdk"
	"github.com/medhen/pc-telemetry-sdk"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize Telemetry
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

	// 2. Initialize ClickHouse (Tier-0 OLAP)
	chConn, err := clickhouse.Open(&clickhouse.Options{
		Addr: cfg.ClickHouse.Addrs,
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouse.Database,
			Username: cfg.ClickHouse.Username,
			Password: cfg.ClickHouse.Password,
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: cfg.ServiceName, Version: cfg.Version},
			},
		},
		MaxOpenConns: 50,
		MaxIdleConns: 10,
	})
	if err != nil {
		slog.Error("Failed to initialize ClickHouse connection", "error", err)
		os.Exit(1)
	}
	defer chConn.Close()

	if err := chConn.Ping(ctx); err != nil {
		slog.Error("ClickHouse ping failed", "error", err)
		os.Exit(1)
	}

	// 3. Initialize Domain, Extractors, Workers
	sealer := extractor.NewCryptoSealer("nbe-secure-hmac-key")
	nbeExtractor := extractor.NewNBEMotorExtractor(chConn, sealer)
	workerPool := job.NewReportWorkerPool(nbeExtractor, 5)

	queryRepo := clickhouse_repo.NewQueryRepository(chConn)
	kpiHandler := query.NewKPIHandler(queryRepo)
	
	// 4. Setup gRPC Server
	grpcServer := grpc.NewServer()
	reportingpb.RegisterReportingQueryServiceServer(grpcServer, reportinggrpc.NewServer(kpiHandler))

	go func() {
		lis, err := net.Listen("tcp", ":9091")
		if err != nil {
			slog.Error("Failed to listen for gRPC", "error", err)
			os.Exit(1)
		}
		slog.Info("gRPC server listening", "addr", ":9091")
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
		}
	}()

	// 5. Setup REST Router
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(10 * time.Second))

	// Auth Middleware
	authCfg := auth.JWTConfig{SecretKey: "dev-secret-key"}
	r.Use(auth.Middleware(authCfg))

	// Idempotency Middleware
	idempCfg := idempotency.Config{RedisURL: "redis://localhost:6379"}
	idempMgr, err := idempotency.NewManager(idempCfg)
	var idempMiddleware func(http.Handler) http.Handler
	if err != nil {
		slog.Warn("Failed to connect to Redis for idempotency. Bypassing.", "error", err)
		idempMiddleware = func(next http.Handler) http.Handler { return next }
	} else {
		idempMiddleware = idempMgr.Middleware
	}

	restHandler := rest.NewKPIRESTHandler(kpiHandler, workerPool)
	restHandler.RegisterRoutes(r, idempMiddleware)

	r.Get("/health/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.Get("/health/readiness", func(w http.ResponseWriter, r *http.Request) {
		if err := chConn.Ping(r.Context()); err != nil {
			http.Error(w, "ClickHouse unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	// 6. Start HTTP Server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		slog.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
		}
	}()

	// 7. Graceful Shutdown
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
