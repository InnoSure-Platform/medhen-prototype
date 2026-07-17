package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/medhen/pc-workflow-svc/config"
	"github.com/medhen/pc-workflow-svc/internal/infrastructure/kafka"
	"github.com/medhen/pc-workflow-svc/internal/infrastructure/postgres"
	"github.com/medhen/pc-workflow-svc/internal/infrastructure/telemetry"
)

// App represents the application container holding dependencies.
type App struct {
	Config          *config.Config
	Logger          *slog.Logger
	DB              *sql.DB
	OutboxRelay     *kafka.OutboxRelay
	ShutdownTracer  func(context.Context) error
}

// New initializes the application container with dependencies.
func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	// Initialize Postgres connection
	dsn := cfg.Database.URL
	if dsn == "" {
		// Fallback for tests
		dsn = "postgres://postgres:postgres@localhost:5432/medhen?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	// 1. Run Database Migrations
	if err := postgres.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// 2. Initialize OpenTelemetry
	shutdownTracer, err := telemetry.InitTracerProvider(context.Background(), "pc-workflow-svc")
	if err != nil {
		logger.Warn("Failed to initialize telemetry", "error", err)
		// We don't block boot for telemetry
		shutdownTracer = func(context.Context) error { return nil }
	}

	// 3. Initialize Kafka Outbox Relay
	brokers := []string{"kafka.infrastructure.svc.cluster.local:9092"} // Hardcoded for demo
	relay := kafka.NewOutboxRelay(db, brokers, logger)

	return &App{
		Config:         cfg,
		Logger:         logger,
		DB:             db,
		OutboxRelay:    relay,
		ShutdownTracer: shutdownTracer,
	}, nil
}

// Start starts the application services (REST, gRPC servers).
func (a *App) Start(ctx context.Context) error {
	a.Logger.Info("Starting Workflow Service", "port", a.Config.Port, "grpc_port", a.Config.GrpcPort)
	
	// Start Kafka Outbox Relay in background
	go a.OutboxRelay.Start(ctx, 2*time.Second)
	
	// Start gRPC server
	// Start REST server
	
	<-ctx.Done()
	a.Logger.Info("Shutting down Workflow Service")
	return nil
}

// Close cleans up resources.
func (a *App) Close() {
	a.Logger.Info("Closing resources")
	if a.ShutdownTracer != nil {
		_ = a.ShutdownTracer(context.Background())
	}
	if a.DB != nil {
		a.DB.Close()
	}
}
