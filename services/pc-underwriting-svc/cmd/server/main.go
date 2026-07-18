package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/medhen/pc-underwriting-svc/config"
	"github.com/medhen/pc-underwriting-svc/internal/application/command"
	"github.com/medhen/pc-underwriting-svc/internal/application/port"
	"github.com/medhen/pc-underwriting-svc/internal/application/query"
	"github.com/medhen/pc-underwriting-svc/internal/application/worker"
	telemetry "github.com/medhen/pc-telemetry-sdk"
	"github.com/medhen/pc-underwriting-svc/internal/infrastructure/grpcclient"
	"github.com/medhen/pc-underwriting-svc/internal/infrastructure/postgres"
	"github.com/medhen/pc-underwriting-svc/internal/presentation/grpc"
	"github.com/medhen/pc-underwriting-svc/internal/presentation/rest"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize OpenTelemetry
	shutdown, err := telemetry.Init(context.Background(), telemetry.Config{
		ServiceName: "pc-underwriting-svc",
		Version:     "v1.0.0",
		Endpoint:    "localhost:4317",
	})
	if err == nil {
		defer shutdown(context.Background())
	}

	// 1. Setup Infrastructure
	// dbPool, err := pgxpool.New(context.Background(), cfg.DBURL)
	// var dbPool interface{} // Mocked for compilation

	// Repositories
	// These casts and nil checks are purely structural mocks to represent the DI graph.
	// We pass nil for dbPool simply to show the Hexagonal wiring.
	assessmentRepo := postgres.NewAssessmentRepo(nil)
	referralRepo := postgres.NewReferralRepo(nil)
	outboxRepo := postgres.NewOutboxRepo(nil)
	productSvcClient := grpcclient.NewProductClient()
	
	// Note: We need a UoW implementation, here we mock it structurally
	var uow port.UnitOfWork

	// Mock Authority Repo
	var authorityRepo port.AuthorityRepository
	var enrichmentProvider port.EnrichmentProvider

	// 2. Setup Application Handlers
	assessRiskHandler := command.NewAssessRiskHandler(uow, assessmentRepo, referralRepo, outboxRepo, productSvcClient, enrichmentProvider)
	submitDecisionHandler := command.NewSubmitDecisionHandler(uow, referralRepo, authorityRepo, outboxRepo)
	assignReferralHandler := command.NewAssignReferralHandler(uow, referralRepo)

	// 3. Setup Presentation Servers
	_ = grpc.NewAssessmentServer(assessRiskHandler)

	sseHandler := rest.NewSSEHandler()
	referralQueries := query.NewReferralQueryService(nil) // nil for mock db pool

	r := chi.NewRouter()
	workbenchHandler := rest.NewWorkbenchHandler(submitDecisionHandler, assignReferralHandler, referralQueries, sseHandler)
	workbenchHandler.RegisterRoutes(r)

	// 4. Start Workers
	slaWorker := worker.NewSLAWorker(uow, referralRepo, outboxRepo)
	go slaWorker.Run(context.Background(), 60*time.Second)

	// 5. Start Servers
	log.Printf("Starting pc-underwriting-svc REST on port %s", cfg.RESTPort)
	log.Printf("Starting pc-underwriting-svc gRPC on port %s", cfg.GRPCPort)

	// In a real app we'd use error groups and graceful shutdown.
	err = http.ListenAndServe(":"+cfg.RESTPort, r)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
