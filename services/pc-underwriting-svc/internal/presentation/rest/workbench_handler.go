package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/medhen/pc-idempotency-mgmt-sdk/idempotency"
	"github.com/medhen/pc-underwriting-svc/internal/application/command"
	"github.com/medhen/pc-underwriting-svc/internal/application/query"
	"github.com/medhen/pc-underwriting-svc/internal/domain/valueobject"
	customMiddleware "github.com/medhen/pc-underwriting-svc/internal/presentation/rest/middleware"
)

type WorkbenchHandler struct {
	submitDecisionHandler *command.SubmitDecisionHandler
	assignReferralHandler *command.AssignReferralHandler
	queries               query.ReferralQueryService
	sseHandler            *SSEHandler
}

func NewWorkbenchHandler(sdh *command.SubmitDecisionHandler, arh *command.AssignReferralHandler, q query.ReferralQueryService, sse *SSEHandler) *WorkbenchHandler {
	return &WorkbenchHandler{
		submitDecisionHandler: sdh,
		assignReferralHandler: arh,
		queries:               q,
		sseHandler:            sse,
	}
}

func (h *WorkbenchHandler) RegisterRoutes(r chi.Router) {
	// Setup mock idempotency store (would be injected in real app)
	idempStore := idempotency.NewMockStore()

	// Register SSE routes
	if h.sseHandler != nil {
		h.sseHandler.RegisterRoutes(r)
	}

	r.Route("/api/pc-underwriting/v1/referrals", func(r chi.Router) {
		r.Get("/", h.ListReferrals)
		r.Get("/{id}", h.GetReferral)
		r.Post("/{id}/assign", h.AssignReferral)
		
		r.With(customMiddleware.IdempotencyMiddleware(idempStore)).
			Post("/{id}/decision", h.SubmitDecision)
	})
}

func (h *WorkbenchHandler) ListReferrals(w http.ResponseWriter, r *http.Request) {
	// Extract advanced filter parameters
	tenantID := r.Header.Get("X-Tenant-ID")
	status := r.URL.Query().Get("status")
	lob := r.URL.Query().Get("lob")
	underwriter := r.URL.Query().Get("underwriter")

	referrals, err := h.queries.ListReferrals(r.Context(), tenantID, status, lob, underwriter, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(referrals)
}

func (h *WorkbenchHandler) GetReferral(w http.ResponseWriter, r *http.Request) {
	// Mock implementation
	w.WriteHeader(http.StatusOK)
}

func (h *WorkbenchHandler) AssignReferral(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	cmd := command.AssignReferralCommand{
		TenantID:      r.Header.Get("X-Tenant-ID"),
		ReferralID:    id,
		UnderwriterID: r.Header.Get("X-User-ID"), // Ideally from auth context
	}

	if err := h.assignReferralHandler.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WorkbenchHandler) SubmitDecision(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	var payload struct {
		Decision                 valueobject.DecisionType `json:"decision"`
		Conditions               []valueobject.Condition  `json:"conditions"`
		DisclosuresAcknowledged  []string                 `json:"disclosures_acknowledged"`
		FacultativeCleared       bool                     `json:"facultative_cleared"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	cmd := command.SubmitDecisionCommand{
		TenantID:           r.Header.Get("X-Tenant-ID"),
		ReferralID:         id,
		UnderwriterID:      r.Header.Get("X-User-ID"),
		ActorLevelCode:     r.Header.Get("X-User-Authority-Level"),
		ProductLOB:         "motor", // Ideally derived from context
		Premium:            5000.0,  // Mocked for example
		TSI:                2500000.0, // Mocked
		Decision:           payload.Decision,
		Conditions:         payload.Conditions,
		Disclosures:        payload.DisclosuresAcknowledged,
		FacultativeCleared: payload.FacultativeCleared,
	}

	if err := h.submitDecisionHandler.Handle(r.Context(), cmd); err != nil {
		// Proper mapping of domain errors to HTTP 403, 409, 400 goes here
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}
