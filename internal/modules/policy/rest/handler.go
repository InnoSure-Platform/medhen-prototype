// Package rest is the driving HTTP adapter for the policy module.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	policyapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/auth"
)

// Handler serves quote/policy endpoints.
type Handler struct{ svc *policyapp.Service }

// New builds the handler.
func New(svc *policyapp.Service) *Handler { return &Handler{svc: svc} }

// Routes returns the module's routes (mounted under /policy by the registry).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /quotes", h.createQuote)
	mux.HandleFunc("GET /quotes", h.listQuotes)
	mux.HandleFunc("GET /quotes/{id}", h.getQuote)
	mux.HandleFunc("POST /quotes/{id}/bind", h.bind)
	mux.HandleFunc("GET /policies", h.listPolicies)
	mux.HandleFunc("GET /policies/{id}", h.getPolicy)
	mux.HandleFunc("POST /policies/{id}/endorse", h.endorse)
	mux.HandleFunc("POST /policies/{id}/cancel", h.cancel)
	mux.HandleFunc("POST /policies/{id}/renew", h.renew)
	return mux
}

func (h *Handler) tenant(r *http.Request, fallback string) string {
	if t := auth.TenantOrHeader(r); t != "" {
		return t
	}
	return fallback
}

func (h *Handler) createQuote(w http.ResponseWriter, r *http.Request) {
	var in policyapp.CreateQuoteInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.TenantID = h.tenant(r, in.TenantID)

	q, err := h.svc.CreateQuote(r.Context(), in)
	if errors.Is(err, policyapp.ErrPartyNotFound) {
		writeError(w, http.StatusUnprocessableEntity, "party not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, q)
}

func (h *Handler) getQuote(w http.ResponseWriter, r *http.Request) {
	q, err := h.svc.GetQuote(r.Context(), h.tenant(r, ""), r.PathValue("id"))
	if errors.Is(err, policyapp.ErrNotFound) {
		writeError(w, http.StatusNotFound, "quote not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, q)
}

func (h *Handler) bind(w http.ResponseWriter, r *http.Request) {
	policy, err := h.svc.BindQuote(r.Context(), h.tenant(r, ""), r.PathValue("id"))
	switch {
	case errors.Is(err, policyapp.ErrNotFound):
		writeError(w, http.StatusNotFound, "quote not found")
	case errors.Is(err, domain.ErrReferred):
		writeError(w, http.StatusConflict, "referred to underwriter")
	case errors.Is(err, domain.ErrDeclined):
		writeError(w, http.StatusUnprocessableEntity, "declined")
	case errors.Is(err, domain.ErrQuoteNotBindable):
		writeError(w, http.StatusConflict, "quote not bindable")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "bind failed")
	default:
		writeJSON(w, http.StatusCreated, policy)
	}
}

func (h *Handler) listPolicies(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.svc.ListPolicies(r.Context(), h.tenant(r, ""), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list failed")
		return
	}
	if items == nil {
		items = []*domain.Policy{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) getPolicy(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetPolicy(r.Context(), h.tenant(r, ""), r.PathValue("id"))
	if errors.Is(err, policyapp.ErrNotFound) {
		writeError(w, http.StatusNotFound, "policy not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) listQuotes(w http.ResponseWriter, r *http.Request) {
	limit, offset := page(r)
	items, err := h.svc.ListQuotes(r.Context(), h.tenant(r, ""), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list failed")
		return
	}
	if items == nil {
		items = []*domain.Quote{}
	}
	writeJSON(w, http.StatusOK, items)
}

type endorseRequest struct {
	PremiumDeltaMinor int64  `json:"premium_delta_minor"`
	Reason            string `json:"reason"`
}

func (h *Handler) endorse(w http.ResponseWriter, r *http.Request) {
	var req endorseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, err := h.svc.EndorsePolicy(r.Context(), h.tenant(r, ""), r.PathValue("id"), req.PremiumDeltaMinor, req.Reason)
	switch {
	case errors.Is(err, policyapp.ErrNotFound):
		writeError(w, http.StatusNotFound, "policy not found")
	case errors.Is(err, domain.ErrNotInForce):
		writeError(w, http.StatusConflict, "policy is not in force")
	case err != nil:
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		writeJSON(w, http.StatusOK, p)
	}
}

type cancelRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	var req cancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, refund, err := h.svc.CancelPolicy(r.Context(), h.tenant(r, ""), r.PathValue("id"), req.Reason)
	switch {
	case errors.Is(err, policyapp.ErrNotFound):
		writeError(w, http.StatusNotFound, "policy not found")
	case errors.Is(err, domain.ErrAlreadyCancelled):
		writeError(w, http.StatusConflict, "policy already cancelled")
	case err != nil:
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		writeJSON(w, http.StatusOK, map[string]any{"policy": p, "refund_minor": refund.Minor()})
	}
}

func (h *Handler) renew(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.RenewPolicy(r.Context(), h.tenant(r, ""), r.PathValue("id"))
	switch {
	case errors.Is(err, policyapp.ErrNotFound):
		writeError(w, http.StatusNotFound, "policy not found")
	case errors.Is(err, domain.ErrNotInForce):
		writeError(w, http.StatusConflict, "policy is not in force")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "renew failed")
	default:
		writeJSON(w, http.StatusCreated, p)
	}
}

// page parses limit/offset with a default (50) and cap (200).
func page(r *http.Request) (limit, offset int) {
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
