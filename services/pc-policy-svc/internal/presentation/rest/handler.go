package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/medhen/pc-policy-svc/internal/application/command"
)

type Handler struct {
	createQuoteHandler    *command.CreateQuoteHandler
	bindPolicyHandler     *command.BindPolicyHandler
	getTimelineHandler    *query.GetTimelineHandler
	searchPoliciesHandler *query.SearchPoliciesHandler
}

func NewHandler(
	cqh *command.CreateQuoteHandler,
	bph *command.BindPolicyHandler,
	gth *query.GetTimelineHandler,
	sph *query.SearchPoliciesHandler,
) *Handler {
	return &Handler{
		createQuoteHandler:    cqh,
		bindPolicyHandler:     bph,
		getTimelineHandler:    gth,
		searchPoliciesHandler: sph,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/pc-policy/v1", func(r chi.Router) {
		r.Post("/quotes", h.CreateQuote)
		r.Patch("/quotes/{id}", h.UpdateQuote)
		r.Post("/quotes/{id}/calculate", h.CalculateQuote)
		r.Post("/policies", h.BindPolicy)
		r.Get("/policies", h.SearchPolicies)
		r.Get("/policies/{id}/timeline", h.GetTimeline)
	})
}

func (h *Handler) CreateQuote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID    string          `json:"tenant_id"`
		ProductID   uuid.UUID       `json:"product_id"`
		PartyID     uuid.UUID       `json:"party_id"`
		RiskPayload json.RawMessage `json:"risk_payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := command.CreateQuoteCommand{
		TenantID:    req.TenantID,
		ProductID:   req.ProductID,
		PartyID:     req.PartyID,
		RiskPayload: req.RiskPayload,
	}

	quote, err := h.createQuoteHandler.Handle(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(quote)
}

func (h *Handler) UpdateQuote(w http.ResponseWriter, r *http.Request) {
	// Mock implementation for Quote Patching in a multi-step wizard
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "DRAFT", "message": "Quote incrementally updated successfully"}`))
}

func (h *Handler) CalculateQuote(w http.ResponseWriter, r *http.Request) {
	// Mock implementation for finalizing the draft and invoking rating
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "QUOTED", "premium": 500.00}`))
}

func (h *Handler) BindPolicy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QuoteID uuid.UUID `json:"quote_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idempKey := r.Header.Get("Idempotency-Key")
	if idempKey == "" {
		http.Error(w, "Idempotency-Key header is required", http.StatusBadRequest)
		return
	}

	cmd := command.BindPolicyCommand{
		QuoteID: req.QuoteID,
	}

	// In a real implementation we would check the Idempotency Store (Redis/Postgres) using the SDK here
	// before invoking the command to prevent double-binding.
	// if pc_idempotency.IsExecuted(ctx, idempKey) { return savedResponse }

	policy, err := h.bindPolicyHandler.Handle(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(policy)
}

func (h *Handler) SearchPolicies(w http.ResponseWriter, r *http.Request) {
	q := query.SearchPoliciesQuery{}
	
	if partyStr := r.URL.Query().Get("party_id"); partyStr != "" {
		partyID, err := uuid.Parse(partyStr)
		if err == nil {
			q.PartyID = &partyID
		}
	}
	if status := r.URL.Query().Get("status"); status != "" {
		q.Status = &status
	}

	results, err := h.searchPoliciesHandler.Handle(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *Handler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	policyID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid policy ID", http.StatusBadRequest)
		return
	}

	q := query.GetTimelineQuery{PolicyID: policyID}
	timeline, err := h.getTimelineHandler.Handle(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeline)
}
