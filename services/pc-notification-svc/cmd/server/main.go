package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"pc-notification-svc/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.LoadConfig()

	slog.Info("Starting pc-notification-svc", "port", cfg.GRPCPort)

	// Context for graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: Initialize Repositories (Postgres)
	// TODO: Initialize Kafka Consumers/Outbox
	// TODO: Initialize gRPC server

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down pc-notification-svc...")
	// TODO: Shutdown gRPC, flush outbox, close DB pool
}
