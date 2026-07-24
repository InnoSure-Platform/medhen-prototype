// Package rest is the driving HTTP adapter for the party module.
package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
)

// Handler serves party endpoints.
type Handler struct {
	svc    *app.Service
	logger *slog.Logger
}

// New builds the handler.
func New(svc *app.Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// Routes returns the module's routes (mounted under /party by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /parties", h.register)
	mux.HandleFunc("GET /parties", h.list)
	mux.HandleFunc("GET /parties/{id}", h.get)
	return mux
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in app.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if t := auth.TenantOrHeader(r); t != "" {
		in.TenantID = t
	}

	id, err := h.svc.Register(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	limit, offset := 50, 0
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 {
		limit = v
	}
	if limit > 200 {
		limit = 200
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v >= 0 {
		offset = v
	}
	items, err := h.svc.List(r.Context(), auth.TenantOrHeader(r), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list failed")
		return
	}
	if items == nil {
		items = []*domain.Party{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.TenantOrHeader(r)
	p, err := h.svc.Get(r.Context(), tenantID, r.PathValue("id"))
	if errors.Is(err, app.ErrNotFound) {
		writeError(w, http.StatusNotFound, "party not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
