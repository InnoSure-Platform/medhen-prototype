package main

import (
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
	"medhen.com/platform/pc-rating-calc-svc/internal/application"
	"medhen.com/platform/pc-rating-calc-svc/internal/infrastructure/cache"
	"medhen.com/platform/pc-rating-calc-svc/internal/infrastructure/messaging"
	grpchandler "medhen.com/platform/pc-rating-calc-svc/internal/presentation/grpc"
)

func main() {
	slog.Info("Initializing pc-rating-calc-svc (Tier-0 Engine)")

	// 1. Init Infra: Redis (Rate Table Cache)
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	rateProvider := cache.NewRedisRateProvider(rdb)

	// 2. Init Infra: Kafka (Audit Publisher)
	// We handle the error gracefully in production, here we use a mock-ready structure
	kafkaProducer, err := messaging.NewKafkaAuditProducer("localhost:9092")
	if err != nil {
		slog.Warn("Failed to initialize Kafka producer, audit telemetry will be disabled", "err", err)
	}

	// 3. Init Application & Domain Orchestrator
	appService := application.NewRatingApplicationService(rateProvider, kafkaProducer)

	// 4. Init Presentation / Ports
	_ = grpchandler.NewRatingHandler(appService)

	slog.Info("Service components wired successfully. Ready to bind ports.")
	
	// Normally here we would start the grpc.Server and listen on a port
	// grpcServer := grpc.NewServer()
	// pb.RegisterRatingServiceServer(grpcServer, handler)
	// _ = grpcServer.Serve(listener)
	
	os.Exit(0)
}
