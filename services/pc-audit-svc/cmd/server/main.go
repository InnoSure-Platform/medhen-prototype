package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"github.com/medhen/pc-audit-svc/internal/application/commands"
	"github.com/medhen/pc-audit-svc/internal/application/services"
	"github.com/medhen/pc-audit-svc/internal/infrastructure/database"
	"github.com/medhen/pc-audit-svc/internal/infrastructure/kms"
	"github.com/medhen/pc-audit-svc/internal/infrastructure/messaging"
	"github.com/medhen/pc-audit-svc/internal/infrastructure/notary"
	auditgrpc "github.com/medhen/pc-audit-svc/internal/presentation/grpc"
	"github.com/medhen/pc-audit-svc/internal/presentation/rest"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Starting pc-audit-svc (Tier-0 Immutable Ledger)")

	// 1. Initialize Infrastructure
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/medhen_audit?sslmode=disable"
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	hotLedger := database.NewPostgresLedger(pool)
	kmsClient := kms.NewKMSClient("http://kms-svc:8080")
	anomalyPublisher := messaging.NewKafkaAnomalyPublisher("platform.security.anomalies.v1")
	notaryClient := notary.NewNotaryClient("https://nbe.gov.et/api/v1/vault")
	_ = notaryClient // In a real app, this would be wired to a Cron Scheduler

	// 2. Initialize Application Core (Domain Services & CQRS Handlers)
	merkleManager := services.NewDefaultMerkleManager()
	appendHandler := commands.NewAppendRecordHandler(hotLedger, kmsClient, merkleManager, anomalyPublisher)

	// 3. Initialize Kafka CDC Consumer
	kafkaBrokers := []string{"localhost:9092"}
	cdcConsumer, err := messaging.NewKafkaConsumer(kafkaBrokers, appendHandler)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}
	cdcConsumer.Start(ctx)
	defer cdcConsumer.Close()

	// 4. Initialize REST Presentation (Search & Export)
	r := chi.NewRouter()
	restHandler := rest.NewAuditHandler()
	restHandler.RegisterRoutes(r)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 5. Initialize gRPC Presentation (Sync Ingestion)
	grpcServer := grpc.NewServer()
	grpcHandler := auditgrpc.NewAuditIngestionServer(appendHandler)
	grpcHandler.Register(grpcServer)

	// Start HTTP Server
	go func() {
		fmt.Println("REST server listening on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC Server
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port: %v", err)
		}
		fmt.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down pc-audit-svc...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}

	fmt.Println("Service shutdown completed gracefully.")
}
