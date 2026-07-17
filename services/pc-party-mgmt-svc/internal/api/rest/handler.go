package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/medhen/pc-party-mgmt-svc/internal/application/command"
	"github.com/medhen/pc-party-mgmt-svc/internal/application/query"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type PartyHandler struct {
	registerCmd      *command.RegisterPartyHandler
	addAddrCmd       *command.AddAddressHandler
	updateConsentCmd *command.UpdateConsentHandler
	anonymizeCmd     *command.AnonymizePartyHandler
	query360         *query.Customer360QueryService
}

func NewPartyHandler(
	registerCmd *command.RegisterPartyHandler, 
	addAddrCmd *command.AddAddressHandler,
	updateConsentCmd *command.UpdateConsentHandler,
	anonymizeCmd *command.AnonymizePartyHandler,
	query360 *query.Customer360QueryService,
) *PartyHandler {
	return &PartyHandler{
		registerCmd:      registerCmd,
		addAddrCmd:       addAddrCmd,
		updateConsentCmd: updateConsentCmd,
		anonymizeCmd:     anonymizeCmd,
		query360:         query360,
	}
}

type RegisterIndividualRequest struct {
	TenantID              string    `json:"tenant_id"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	DOB                   time.Time `json:"date_of_birth"`
	Gender                string    `json:"gender"`
	NationalIDType        string    `json:"national_id_type"`
	NationalIDNumber      string    `json:"national_id_number"`
	TIN                   string    `json:"tin"`
	OverrideDuplicateFlag bool      `json:"override_duplicate_flag"`
}

func (h *PartyHandler) RegisterIndividual(w http.ResponseWriter, r *http.Request) {
	var req RegisterIndividualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := command.RegisterIndividualCommand{
		TenantID:              req.TenantID,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		DOB:                   req.DOB,
		Gender:                req.Gender,
		NationalIDType:        req.NationalIDType,
		NationalIDNumber:      req.NationalIDNumber,
		TIN:                   req.TIN,
		OverrideDuplicateFlag: req.OverrideDuplicateFlag,
	}

	party, err := h.registerCmd.HandleIndividual(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(party)
}

type AddAddressRequest struct {
	Type        string `json:"type"`
	IsPrimary   bool   `json:"is_primary"`
	Region      string `json:"region"`
	Zone        string `json:"zone"`
	Woreda      string `json:"woreda"`
	Kebele      string `json:"kebele"`
	HouseNumber string `json:"house_number"`
}

func (h *PartyHandler) AddAddress(w http.ResponseWriter, r *http.Request) {
	partyIDStr := chi.URLParam(r, "id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		http.Error(w, "invalid party ID", http.StatusBadRequest)
		return
	}

	var req AddAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := command.AddAddressCommand{
		PartyID:     partyID,
		Type:        req.Type,
		IsPrimary:   req.IsPrimary,
		Region:      req.Region,
		Zone:        req.Zone,
		Woreda:      req.Woreda,
		Kebele:      req.Kebele,
		HouseNumber: req.HouseNumber,
	}

	if err := h.addAddrCmd.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *PartyHandler) GetCustomer360(w http.ResponseWriter, r *http.Request) {
	partyIDStr := chi.URLParam(r, "id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		http.Error(w, "invalid party ID", http.StatusBadRequest)
		return
	}
	
	tenantID := r.Header.Get("X-Tenant-ID") // Example header usage

	view, err := h.query360.GetCustomer360(r.Context(), tenantID, partyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(view)
}

type UpdateConsentRequest struct {
	ConsentType string `json:"consent_type"`
	Status      string `json:"status"`
}

func (h *PartyHandler) UpdateConsent(w http.ResponseWriter, r *http.Request) {
	partyIDStr := chi.URLParam(r, "id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		http.Error(w, "invalid party ID", http.StatusBadRequest)
		return
	}

	var req UpdateConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := command.UpdateConsentCommand{
		PartyID:     partyID,
		ConsentType: req.ConsentType,
		Status:      domain.ConsentStatus(req.Status),
	}

	if err := h.updateConsentCmd.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type ErasureRequest struct {
	Reason string `json:"reason"`
}

func (h *PartyHandler) RequestErasure(w http.ResponseWriter, r *http.Request) {
	partyIDStr := chi.URLParam(r, "id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		http.Error(w, "invalid party ID", http.StatusBadRequest)
		return
	}

	var req ErasureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := command.AnonymizePartyCommand{
		PartyID: partyID,
		Reason:  req.Reason,
	}

	if err := h.anonymizeCmd.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PartyHandler) RegisterRoutes(r chi.Router) {
	r.Post("/parties/individuals", h.RegisterIndividual)
	r.Post("/parties/{id}/addresses", h.AddAddress)
	r.Get("/parties/{id}/360", h.GetCustomer360)
	r.Put("/parties/{id}/consents", h.UpdateConsent)
	r.Post("/parties/{id}/erasure-request", h.RequestErasure)
}
