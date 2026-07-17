package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	adapterhttp "medhen.com/pc-observability-svc/internal/adapters/http"
	"medhen.com/pc-observability-svc/internal/app"
	"medhen.com/pc-observability-svc/internal/domain"

	telemetry "github.com/medhen/pc-telemetry-sdk"
)

// Config holds the environment configuration.
type Config struct {
	Port         int           `envconfig:"PORT" default:"8080"`
	LogLevel     string        `envconfig:"LOG_LEVEL" default:"info"`
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"10s"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
}

// Mock repository for scaffolding
type mockSLORepo struct{}

func (m *mockSLORepo) Save(ctx context.Context, slo *domain.SLO) error { return nil }
func (m *mockSLORepo) FindByID(ctx context.Context, id string) (*domain.SLO, error) {
	return nil, nil
}

// Mock mimir client for scaffolding
type mockMimirClient struct{}

func (m *mockMimirClient) PushRules(ctx context.Context, slo *domain.SLO) error { return nil }

func main() {
	// 1. Load Configuration
	var cfg Config
	if err := envconfig.Process("obs", &cfg); err != nil {
		fmt.Printf("failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Initialize Telemetry (OpenTelemetry + slog)
	telCfg := telemetry.Config{
		ServiceName: "pc-observability-svc",
		Version:     "v1.0.0",
		Endpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
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

	slog.Info("Starting pc-observability-svc (Control Plane)", "port", cfg.Port)

	// 3. Wire up the application
	sloRepo := &mockSLORepo{}
	mimirClient := &mockMimirClient{}
	sloService := app.NewSLOService(sloRepo, mimirClient)

	// 4. Setup HTTP Server
	router := adapterhttp.NewRouter(sloService)
	mux := router.SetupRoutes()

	certFile := "../../certs/server.crt"
	keyFile := "../../certs/server.key"

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// 5. Graceful shutdown
	go func() {
		slog.Info(fmt.Sprintf("Listening on :%d (HTTPS)", cfg.Port))
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down gracefully...")

	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutDown()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting")
}
