package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/medhen/pc-idempotency-mgmt-sdk"
	"github.com/medhen/pc-integration-svc/internal/application/command"
	"go.uber.org/zap"
)

type WebhookHandler struct {
	logger        *zap.Logger
	ingestCommand *command.IngestWebhookHandler
	telebirrSecret string
}

func NewWebhookHandler(logger *zap.Logger, ingestCommand *command.IngestWebhookHandler, telebirrSecret string) *WebhookHandler {
	return &WebhookHandler{
		logger:        logger,
		ingestCommand: ingestCommand,
		telebirrSecret: telebirrSecret,
	}
}

func (h *WebhookHandler) RegisterRoutes(r chi.Router, idempManager *idempotency.Manager) {
	// Group webhooks
	r.Route("/v1/webhooks", func(r chi.Router) {
		// Mount the SDK's Idempotency Middleware.
		r.Use(idempManager.Middleware)

		r.Post("/telebirr/callback", h.handleTelebirrCallback)
		r.Post("/cbe/notification", h.handleCBENotification)
	})
}

// verifySignature checks the HMAC-SHA256 signature of the payload
func (h *WebhookHandler) verifySignature(payload []byte, signature string) bool {
	// In a real implementation, this would compute the HMAC-SHA256 of the payload using h.telebirrSecret
	// and compare it in constant time with the provided signature.
	// For MVP, we'll just check if a signature is provided (or simulate).
	if signature == "" && h.telebirrSecret != "" {
		h.logger.Warn("Missing signature on webhook")
		return false
	}
	return true
}

// handleTelebirrCallback processes async payment results from Telebirr.
func (h *WebhookHandler) handleTelebirrCallback(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	signature := r.Header.Get("X-Telebirr-Signature")
	if !h.verifySignature(bodyBytes, signature) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// Validate signature here (skipped for brevity)
	// Example parsing
	var payload struct {
		OutTradeNo  string `json:"outTradeNo"` // Internal Reference
		TradeNo     string `json:"tradeNo"`    // Provider Transaction ID
		TradeStatus int    `json:"tradeStatus"` // 1 = Success, 2 = Failed
	}

	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	internalRef, err := uuid.Parse(payload.OutTradeNo)
	if err != nil {
		http.Error(w, "invalid internal reference", http.StatusBadRequest)
		return
	}

	cmd := command.IngestWebhookCmd{
		Provider:              "TELEBIRR",
		ProviderTransactionID: payload.TradeNo,
		RawPayload:            bodyBytes,
		StatusIsSuccess:       payload.TradeStatus == 1,
		InternalReferenceID:   internalRef,
		Amount:                0, // Ideally extracted from payload
		Currency:              "ETB",
	}

	if err := h.ingestCommand.Handle(r.Context(), cmd); err != nil {
		h.logger.Error("failed to ingest webhook", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Vendor expects 200 OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"code": 0, "msg": "success"}`))
}

func (h *WebhookHandler) handleCBENotification(w http.ResponseWriter, r *http.Request) {
	// Implementation follows similar pattern to Telebirr but with CBE's payload scheme (possibly XML)
	w.WriteHeader(http.StatusOK)
}
