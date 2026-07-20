// Package rest is the driving HTTP adapter for the document module.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	docapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
)

// Handler serves document endpoints.
type Handler struct{ svc *docapp.Service }

// New builds the handler.
func New(svc *docapp.Service) *Handler { return &Handler{svc: svc} }

// Routes returns the module's routes (mounted under /document by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /documents/{id}", h.get)
	return mux
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.Get(r.Context(), auth.TenantOrHeader(r), r.PathValue("id"))
	if errors.Is(err, docapp.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "document not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "lookup failed"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(d)
}
