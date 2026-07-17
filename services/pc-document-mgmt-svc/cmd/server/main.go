package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Starting pc-document-mgmt-svc...")
	log.Println("Tier-0 Service: Document Management Engine")
	log.Println("Initializing dependencies...")

	// 1. Load config (stubbed)
	// 2. Initialize DB Connection (stubbed)
	// 3. Initialize Ozone Client
	// 4. Initialize Chromium Renderer Pool
	// 5. Initialize Kafka Publisher
	// 6. Wire Domain Repositories (stubbed)
	// 7. Wire Application Use Cases (stubbed)
	
	// Example wiring for REST upload
	// repo := infrastructure.NewPostgresDocumentRepository(db)
	// storage := infrastructure.NewOzoneS3Client(...)
	// scanner := infrastructure.NewICAPMalwareScanner(...)
	// pub := infrastructure.NewKafkaOutboxPublisher(...)
	// uploadUC := application.NewUploadDocumentUseCase(repo, storage, scanner, pub)
	// restHandler := presentation.NewDocumentRestHandler(uploadUC)
	// http.HandleFunc("/api/pc-document-mgmt/v1/documents/upload", restHandler.UploadDocument)

	// 8. Start gRPC & REST Servers
	go func() {
		log.Println("Starting REST Edge Server on :8080...")
		// if err := http.ListenAndServe(":8080", nil); err != nil {
		// 	log.Fatalf("REST server failed: %v", err)
		// }
	}()

	log.Println("Service pc-document-mgmt-svc is running. Waiting for termination signal...")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down pc-document-mgmt-svc gracefully...")
}
