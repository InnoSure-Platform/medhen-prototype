package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"medhen/pc-claims-svc/internal/application/commands"
	"medhen/pc-claims-svc/internal/domain"
	"medhen/pc-claims-svc/internal/infrastructure/ozone"
)

type ClaimHandler struct {
	submitFNOLHandler *commands.SubmitFNOLHandler
	ozoneClient       *ozone.OzoneClient
}

func NewClaimHandler(submitFNOL *commands.SubmitFNOLHandler, ozoneClient *ozone.OzoneClient) *ClaimHandler {
	return &ClaimHandler{
		submitFNOLHandler: submitFNOL,
		ozoneClient:       ozoneClient,
	}
}

func (h *ClaimHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.SubmitFNOL)
	r.Post("/{id}/reserves", h.AdjustReserve)
	r.Post("/{id}/evidence/presign", h.GeneratePresignURL)
}

type FNOLRequest struct {
	TenantID    string    `json:"tenant_id"`
	PolicyID    string    `json:"policy_id"`
	LossType    string    `json:"loss_type"`
	DateOfLoss  time.Time `json:"date_of_loss"`
}

// SubmitFNOL handles the incoming First Notice of Loss
func (h *ClaimHandler) SubmitFNOL(w http.ResponseWriter, r *http.Request) {
	var req FNOLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ideally this executes a Command via Application layer:
	// claimID, err := h.commandBus.Execute(ctx, application.SubmitFNOLCommand{...})
	
	// Mock response for structure
	claim := domain.NewClaim(req.TenantID, req.PolicyID, req.LossType, req.DateOfLoss, false)
	claim.ID = "c-mock-uuid-1234"
	claim.ClaimNumber = "CLM/MOC/2026/001"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"claim_id":     claim.ID,
		"claim_number": claim.ClaimNumber,
		"status":       string(claim.Status),
	})
}

// AdjustReserve handles reserve mutations
func (h *ClaimHandler) AdjustReserve(w http.ResponseWriter, r *http.Request) {
	// ... Decode payload, execute AdjustReserveCommand
	w.WriteHeader(http.StatusAccepted)
}

// GeneratePresignURL handles requests for secure upload links
func (h *ClaimHandler) GeneratePresignURL(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default-tenant"
	}

	var req struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url, err := h.ozoneClient.GeneratePresignedUploadURL(r.Context(), tenantID, claimID, req.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"upload_url": url,
		"expires_in": "900s",
	})
}
