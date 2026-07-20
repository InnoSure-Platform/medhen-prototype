// Package rest is the driving HTTP adapter for the IAM module.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	iamapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
)

// Handler serves IAM endpoints.
type Handler struct{ svc *iamapp.Service }

// New builds the handler.
func New(svc *iamapp.Service) *Handler { return &Handler{svc: svc} }

// Routes returns the module's routes (mounted under /iam by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", h.register)
	mux.HandleFunc("GET /users/{id}", h.get)
	return mux
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in iamapp.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if t := auth.TenantOrHeader(r); t != "" {
		in.TenantID = t
	}
	u, err := h.svc.Register(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	u, err := h.svc.Get(r.Context(), auth.TenantOrHeader(r), r.PathValue("id"))
	if errors.Is(err, iamapp.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
