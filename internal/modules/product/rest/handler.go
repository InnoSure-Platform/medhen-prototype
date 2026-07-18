// Package rest is the driving HTTP adapter for the product module.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/app"
)

// Handler serves product catalog endpoints.
type Handler struct{ svc *app.Service }

// New builds the handler.
func New(svc *app.Service) *Handler { return &Handler{svc: svc} }

// Routes returns the module's routes (mounted under /product by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /products", h.list)
	mux.HandleFunc("GET /products/{code}", h.get)
	return mux
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ps, err := h.svc.ListProducts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list failed"})
		return
	}
	writeJSON(w, http.StatusOK, ps)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetProduct(r.Context(), r.PathValue("code"))
	if errors.Is(err, app.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup failed"})
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
