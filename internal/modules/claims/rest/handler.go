// Package rest is the driving HTTP adapter for the claims module.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	claimsapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// Handler serves claims endpoints.
type Handler struct{ svc *claimsapp.Service }

// New builds the handler.
func New(svc *claimsapp.Service) *Handler { return &Handler{svc: svc} }

// Routes returns the module's routes (mounted under /claims by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /claims", h.fnol)
	mux.HandleFunc("GET /claims", h.list)
	mux.HandleFunc("GET /claims/{id}", h.get)
	mux.HandleFunc("POST /claims/{id}/settle", h.settle)
	return mux
}

func (h *Handler) fnol(w http.ResponseWriter, r *http.Request) {
	var in claimsapp.FNOLInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if t := auth.TenantOrHeader(r); t != "" {
		in.TenantID = t
	}

	c, err := h.svc.FileFNOL(r.Context(), in)
	switch {
	case errors.Is(err, claimsapp.ErrPolicyNotFound):
		writeError(w, http.StatusUnprocessableEntity, "policy not found")
	case errors.Is(err, claimsapp.ErrPolicyNotActive):
		writeError(w, http.StatusUnprocessableEntity, "policy not in force")
	case err != nil:
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		writeJSON(w, http.StatusCreated, c)
	}
}

type settleRequest struct {
	AmountMinor int64 `json:"amount_minor"`
}

func (h *Handler) settle(w http.ResponseWriter, r *http.Request) {
	var req settleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c, err := h.svc.FastTrackSettle(r.Context(), auth.TenantOrHeader(r), r.PathValue("id"), money.FromMinor(req.AmountMinor))
	switch {
	case errors.Is(err, claimsapp.ErrNotFound):
		writeError(w, http.StatusNotFound, "claim not found")
	case errors.Is(err, domain.ErrAuthorityExceeded):
		writeError(w, http.StatusConflict, "exceeds fast-track authority (referred)")
	case errors.Is(err, domain.ErrAlreadySettled):
		writeError(w, http.StatusConflict, "claim already settled")
	case err != nil:
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		writeJSON(w, http.StatusOK, c)
	}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.svc.ListClaims(r.Context(), auth.TenantOrHeader(r), r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list failed")
		return
	}
	if items == nil {
		items = []*domain.Claim{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.GetClaim(r.Context(), auth.TenantOrHeader(r), r.PathValue("id"))
	if errors.Is(err, claimsapp.ErrNotFound) {
		writeError(w, http.StatusNotFound, "claim not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// pageParams parses limit/offset with a sane default (50) and cap (200).
func pageParams(r *http.Request) (limit, offset int) {
	limit, offset = 50, 0
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 {
		limit = v
	}
	if limit > 200 {
		limit = 200
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v >= 0 {
		offset = v
	}
	return limit, offset
}
