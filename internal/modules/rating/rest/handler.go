// Package rest is the driving HTTP adapter for the rating module.
package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
)

// Handler serves rating endpoints backed by a Calculator.
type Handler struct {
	calc   ports.Calculator
	logger *slog.Logger
}

// New builds the handler.
func New(calc ports.Calculator, logger *slog.Logger) *Handler {
	return &Handler{calc: calc, logger: logger}
}

// Routes returns the module's HTTP routes (mounted under /rating by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /quote", h.quote)
	return mux
}

func (h *Handler) quote(w http.ResponseWriter, r *http.Request) {
	var req ports.PremiumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// Prefer the authenticated tenant when present; fall back to the body only
	// for unauthenticated/dev calls.
	if tid, err := auth.GetTenantID(r.Context()); err == nil {
		req.TenantID = tid
	}

	bd, err := h.calc.Calculate(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, bd)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
