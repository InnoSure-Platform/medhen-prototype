// Package rest is the driving HTTP adapter for the reporting module.
package rest

import (
	"encoding/json"
	"net/http"

	reportapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/reporting/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
)

// Handler serves reporting endpoints.
type Handler struct{ svc *reportapp.Service }

// New builds the handler.
func New(svc *reportapp.Service) *Handler { return &Handler{svc: svc} }

// Routes returns the module's routes (mounted under /reporting by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /kpis", h.kpis)
	return mux
}

func (h *Handler) kpis(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.KPIs(r.Context(), auth.TenantOrHeader(r))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "kpi query failed"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(view)
}
