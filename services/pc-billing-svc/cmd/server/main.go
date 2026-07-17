package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting pc-billing-svc (Tier-0)...")

	// In an enterprise scenario, dependencies are wired up here
	// uow := postgres.NewUnitOfWork(dbPool)
	// paymentRepo := postgres.NewPaymentRepository()
	// invoiceRepo := postgres.NewInvoiceRepository()
	// accountRepo := postgres.NewBillingAccountRepository()
	// ledgerRepo := postgres.NewLedgerRepository()

	// paymentHandler := command.NewProcessPaymentCallbackHandler(uow, paymentRepo, invoiceRepo, accountRepo, ledgerRepo)
	// webhookHandler := webhook.NewTelebirrWebhookHandler(paymentHandler)

	r := gin.Default()
	
	v1 := r.Group("/api/pc-billing/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "up", "tier": 0})
		})
		
		// v1.POST("/webhooks/telebirr/callback", webhookHandler.HandleCallback)
	}

	log.Println("pc-billing-svc listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
