// Package rest is the driving HTTP adapter for the billing module, including the
// Telebirr payment webhook.
package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/adapters"
	billingapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// Handler serves billing endpoints.
type Handler struct {
	svc           *billingapp.Service
	webhookSecret string
}

// New builds the handler.
func New(svc *billingapp.Service, webhookSecret string) *Handler {
	return &Handler{svc: svc, webhookSecret: webhookSecret}
}

// Routes returns the module's routes (mounted under /billing by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhooks/telebirr", h.telebirrWebhook)
	mux.HandleFunc("GET /invoices/{id}", h.getInvoice)
	return mux
}

type telebirrNotification struct {
	TenantID    string `json:"tenant_id"`
	InvoiceID   string `json:"invoice_id"`
	AmountMinor int64  `json:"amount_minor"`
	Reference   string `json:"reference"`
	Status      string `json:"status"`
}

func (h *Handler) telebirrWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot read body")
		return
	}
	// Verify the HMAC signature over the raw body before trusting anything.
	if !adapters.VerifyTelebirrSignature(h.webhookSecret, body, r.Header.Get("X-Telebirr-Signature")) {
		writeError(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	var n telebirrNotification
	if err := json.Unmarshal(body, &n); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if n.Status != "SUCCESS" {
		// Acknowledge non-success callbacks without applying a payment.
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	tenant := n.TenantID
	if t := auth.TenantOrHeader(r); t != "" {
		tenant = t
	}
	inv, err := h.svc.RecordPayment(r.Context(), tenant, n.InvoiceID,
		money.FromMinor(n.AmountMinor), "TELEBIRR", n.Reference)
	if errors.Is(err, billingapp.ErrNotFound) {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	inv, err := h.svc.GetInvoice(r.Context(), auth.TenantOrHeader(r), r.PathValue("id"))
	if errors.Is(err, billingapp.ErrNotFound) {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
