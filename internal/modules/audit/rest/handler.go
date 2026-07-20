// Package rest is the driving HTTP adapter for the audit module.
package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	auditapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/audit/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
)

// Handler serves the audit trail query endpoint.
type Handler struct{ rec *auditapp.Recorder }

// New builds the handler.
func New(rec *auditapp.Recorder) *Handler { return &Handler{rec: rec} }

// Routes returns the module's routes (mounted under /audit by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /logs", h.list)
	return mux
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	tenant := auth.TenantOrHeader(r)
	topic := r.URL.Query().Get("topic")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	records, err := h.rec.List(r.Context(), tenant, topic, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "query failed"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(records)
}
