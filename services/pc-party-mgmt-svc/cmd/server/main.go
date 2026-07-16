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

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	googlegrpc "google.golang.org/grpc"
	
	partypb "github.com/medhen/pc-contracts/gen/go/party/v1"

	"github.com/medhen/pc-party-mgmt-svc/internal/application/command"
	"github.com/medhen/pc-party-mgmt-svc/internal/application/query"
	"github.com/medhen/pc-party-mgmt-svc/internal/application/saga"
	"github.com/medhen/pc-party-mgmt-svc/internal/infrastructure/elasticsearch"
	"github.com/medhen/pc-party-mgmt-svc/internal/infrastructure/fayda"
	"github.com/medhen/pc-party-mgmt-svc/internal/infrastructure/kafka"
	"github.com/medhen/pc-party-mgmt-svc/internal/infrastructure/postgres"
	partygrpc "github.com/medhen/pc-party-mgmt-svc/internal/presentation/grpc"
	"github.com/medhen/pc-party-mgmt-svc/internal/presentation/rest"

	auth "github.com/medhen/pc-auth-sdk"
	idempotency "github.com/medhen/pc-idempotency-mgmt-sdk"
	telemetry "github.com/medhen/pc-telemetry-sdk"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize Telemetry
	telCfg := telemetry.Config{
		ServiceName: "pc-party-mgmt-svc",
		Version:     "v1.0.0",
		Endpoint:    "localhost:4317",
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

	slog.Info("Starting pc-party-mgmt-svc")

	// 2. Initialize Database & Infrastructure
	dbPool, err := pgxpool.New(ctx, "postgres://user:pass@localhost:5432/pc_party_db")
	if err != nil {
		slog.Error("Failed to parse database URL", "error", err)
	}
	defer dbPool.Close()

	esClient, err := es.NewDefaultClient()
	if err != nil {
		slog.Error("Failed to initialize elasticsearch client", "error", err)
	}

	uow := postgres.NewUnitOfWork(dbPool)
	searchRepo := elasticsearch.NewSearchRepository(esClient)
	faydaClient := fayda.NewClient("http://pc-iam-svc.medhen.svc.cluster.local:8080") // Example internal DNS

	// 3. Initialize CQRS Handlers
	registerPartyCmd := command.NewRegisterPartyHandler(uow, searchRepo, faydaClient)
	addAddrCmd := command.NewAddAddressHandler(uow)
	query360 := query.NewCustomer360QueryService(dbPool)

	// 4. Initialize API Router
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(5 * time.Second))

	// Auth SDK Integration
	authCfg := auth.JWTConfig{SecretKey: "dev-secret-key"}
	r.Use(auth.Middleware(authCfg))

	r.Get("/health/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Idempotency SDK Integration
	idempCfg := idempotency.Config{RedisURL: "redis://localhost:6379"}
	idempMgr, err := idempotency.NewManager(idempCfg)
	var idempMiddleware func(http.Handler) http.Handler
	if err != nil {
		slog.Warn("Failed to connect to Redis for idempotency. Continuing without idempotency enforcement", "error", err)
		// Dummy middleware if redis is down
		idempMiddleware = func(next http.Handler) http.Handler { return next }
	} else {
		idempMiddleware = idempMgr.Middleware
	}

	partyHandler := rest.NewPartyHandler(registerPartyCmd, addAddrCmd, query360)
	// We map routes manually here to wrap POST requests in idempotency middleware
	r.Route("/api/pc-party-mgmt/v1", func(r chi.Router) {
		r.With(idempMiddleware).Post("/parties/individual", partyHandler.RegisterIndividual)
		r.With(idempMiddleware).Post("/parties/{id}/addresses", partyHandler.AddAddress)
		r.Get("/parties/{id}/360", partyHandler.GetCustomer360)
	})

	// 5. Start Outbox Worker
	kafkaBrokers := []string{"localhost:9092"} // Hardcoded for prototype
	kafkaPublisher := kafka.NewOutboxPublisher(kafkaBrokers)
	defer kafkaPublisher.Close()

	outboxWorker := kafka.NewOutboxWorker(dbPool, kafkaPublisher, 100*time.Millisecond)
	go outboxWorker.Start(ctx)

	// 6. Start KYC Retry Saga
	kycSaga := saga.NewKYCRetrySaga(dbPool, uow, faydaClient)
	go kycSaga.Start(ctx)

	// 7. Start gRPC Server
	grpcListener, err := net.Listen("tcp", ":9090")
	if err != nil {
		slog.Error("Failed to listen for gRPC", "error", err)
	} else {
		grpcServer := googlegrpc.NewServer()
		partyGrpcServer := partygrpc.NewPartyServer(dbPool)
		partypb.RegisterPartyResolutionServiceServer(grpcServer, partyGrpcServer)

		go func() {
			slog.Info("gRPC server listening on :9090")
			if err := grpcServer.Serve(grpcListener); err != nil {
				slog.Error("gRPC server failed", "error", err)
			}
		}()
	}

	// 8. Start HTTP Server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		slog.Info("HTTP server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down gracefully...")
	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutDown()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}
	slog.Info("Server exited")
}
