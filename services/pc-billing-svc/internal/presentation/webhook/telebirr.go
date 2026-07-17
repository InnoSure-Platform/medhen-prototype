package webhook

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medhen/pc-billing-svc/internal/application/command"
	"github.com/shopspring/decimal"
)

type TelebirrCallbackPayload struct {
	TransactionID string `json:"transaction_id"`
	Method        string `json:"method"`
	Amount        string `json:"amount"` // string to preserve exact decimal
	Currency      string `json:"currency"`
	Status        string `json:"status"`
}

type TelebirrWebhookHandler struct {
	paymentCallbackHandler *command.ProcessPaymentCallbackHandler
}

func NewTelebirrWebhookHandler(handler *command.ProcessPaymentCallbackHandler) *TelebirrWebhookHandler {
	return &TelebirrWebhookHandler{
		paymentCallbackHandler: handler,
	}
}

func (h *TelebirrWebhookHandler) HandleCallback(c *gin.Context) {
	// 1. Verify HMAC signature (simplified for implementation example)
	signature := c.GetHeader("X-Telebirr-Signature")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature"})
		return
	}

	// 2. Parse payload
	var payload TelebirrCallbackPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}

	if payload.Status != "SUCCESS" {
		// Only processing success for now
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	amount, err := decimal.NewFromString(payload.Amount)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid decimal amount"})
		return
	}

	// 3. Dispatch to CQRS command
	cmd := command.ProcessPaymentCallbackCmd{
		TenantID:             "DEFAULT_TENANT", // Typically resolved from context/path
		GatewayTransactionID: payload.TransactionID,
		Method:               payload.Method,
		Amount:               amount,
		InvoiceID:            nil, // In a real scenario, this is passed via webhook metadata
	}

	err = h.paymentCallbackHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Acknowledge the gateway
	c.JSON(http.StatusOK, gin.H{"status": "acknowledged"})
}
