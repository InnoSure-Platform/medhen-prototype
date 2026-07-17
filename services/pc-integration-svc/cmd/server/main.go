package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-idempotency-mgmt-sdk"
	pb "github.com/medhen/pc-contracts/gen/go/integration/v1"
	api_grpc "github.com/medhen/pc-integration-svc/internal/api/grpc"
	"github.com/medhen/pc-integration-svc/internal/api/rest"
	"github.com/medhen/pc-integration-svc/internal/application/command"
	"github.com/medhen/pc-integration-svc/internal/application/ports"
	"github.com/medhen/pc-integration-svc/internal/infrastructure/kafka"
	"github.com/medhen/pc-integration-svc/internal/infrastructure/postgres"
	"github.com/medhen/pc-integration-svc/internal/infrastructure/providers"
	vault "github.com/hashicorp/vault/api"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- OpenBao (Vault) Setup ---
	vaultConfig := vault.DefaultConfig()
	vaultConfig.Address = getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200")
	vaultClient, err := vault.NewClient(vaultConfig)
	if err != nil {
		logger.Fatal("failed to initialize OpenBao client", zap.Error(err))
	}
	vaultClient.SetToken(getEnvOrDefault("VAULT_TOKEN", "mock-token")) // Fallback for local dev

	telebirrSecret := "mock-telebirr-secret"
	faydaAPIKey := "mock-fayda-key"
	smsAPIKey := "mock-sms-key"
	erpAPIKey := "mock-erp-key"

	// Attempt to read secrets, gracefully fallback if Vault is not fully setup locally
	secret, err := vaultClient.Logical().Read("secret/data/medhen/integration")
	if err == nil && secret != nil && secret.Data != nil {
		if data, ok := secret.Data["data"].(map[string]interface{}); ok {
			if v, ok := data["telebirr_secret"].(string); ok { telebirrSecret = v }
			if v, ok := data["fayda_api_key"].(string); ok { faydaAPIKey = v }
			if v, ok := data["sms_api_key"].(string); ok { smsAPIKey = v }
			if v, ok := data["erp_api_key"].(string); ok { erpAPIKey = v }
		}
	} else {
		logger.Warn("Failed to read from OpenBao; using local mock secrets")
	}

	// --- PostgreSQL Setup ---
	dbUrl := getEnvOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/integration?sslmode=disable")
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		logger.Fatal("failed to connect to db", zap.Error(err))
	}
	defer pool.Close()

	// --- Kafka Setup ---
	kClient, err := kgo.NewClient(kgo.SeedBrokers("localhost:9092"))
	if err != nil {
		logger.Fatal("failed to connect to kafka", zap.Error(err))
	}
	defer kClient.Close()

	// --- Redis Idempotency Manager Setup ---
	redisUrl := getEnvOrDefault("REDIS_URL", "redis://localhost:6379/0")
	idempManager, err := idempotency.NewManager(idempotency.Config{RedisURL: redisUrl})
	if err != nil {
		logger.Fatal("failed to setup idempotency manager", zap.Error(err))
	}

	// --- Adapters Setup ---
	txnRepo := postgres.NewTransactionRepository(pool)
	webhookRepo := postgres.NewWebhookRepository(pool)
	outboxPub := kafka.NewOutboxPublisher(kClient)

	telebirrClient := providers.NewTelebirrClient(getEnvOrDefault("TELEBIRR_URL", "http://localhost:8081"))
	faydaClient := providers.NewFaydaClient(getEnvOrDefault("FAYDA_URL", "http://localhost:8081"), faydaAPIKey)
	smsClient := providers.NewSMSClient(getEnvOrDefault("SMS_URL", "http://localhost:8081"), smsAPIKey)
	_ = providers.NewERPClient(getEnvOrDefault("ERP_URL", "http://localhost:8081"), erpAPIKey)

	providerMap := map[string]ports.PaymentProvider{
		"TELEBIRR": telebirrClient,
	}

	// --- Handlers Setup ---
	ingestCmd := command.NewIngestWebhookHandler(webhookRepo, txnRepo, outboxPub)
	initiatePaymentCmd := command.NewInitiatePaymentHandler(txnRepo, providerMap)

	// --- REST Webhook Server ---
	r := chi.NewRouter()
	webhookHandler := rest.NewWebhookHandler(logger, ingestCmd, telebirrSecret)
	webhookHandler.RegisterRoutes(r, idempManager)

	httpSrv := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		logger.Info("starting REST webhook server on :8080")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen error", zap.Error(err))
		}
	}()

	// --- gRPC Server ---
	grpcSrv := grpc.NewServer()
	integrationServer := api_grpc.NewServer(logger, initiatePaymentCmd, faydaClient, smsClient)
	pb.RegisterIntegrationServiceServer(grpcSrv, integrationServer)
	reflection.Register(grpcSrv)

	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			logger.Fatal("failed to listen for grpc", zap.Error(err))
		}
		logger.Info("starting gRPC server on :50051")
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatal("grpc server error", zap.Error(err))
		}
	}()

	// --- Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down servers...")
	ctxShut, cancelShut := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShut()
	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctxShut); err != nil {
		logger.Fatal("rest server shutdown failed", zap.Error(err))
	}
	logger.Info("server exited cleanly")
}

func getEnvOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
