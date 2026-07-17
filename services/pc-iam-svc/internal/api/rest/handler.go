package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/medhen/pc-iam-svc/internal/application"
)

type Handler struct {
	tenantUseCase *application.TenantUseCase
	userUseCase   *application.UserUseCase
}

func NewHandler(t *application.TenantUseCase, u *application.UserUseCase) *Handler {
	return &Handler{tenantUseCase: t, userUseCase: u}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Route("/api/pc-iam-svc/v1", func(r chi.Router) {
		r.Post("/tenants", h.ProvisionTenant)
		r.Post("/tenants/{tenantID}/users", h.ProvisionUser)
	})
}

type provisionTenantReq struct {
	Name string `json:"name"`
}

func (h *Handler) ProvisionTenant(w http.ResponseWriter, r *http.Request) {
	var req provisionTenantReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	tenant, err := h.tenantUseCase.ProvisionTenant(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tenant)
}

type provisionUserReq struct {
	PartyID         string `json:"party_id"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	InitialPassword string `json:"initial_password"`
}

func (h *Handler) ProvisionUser(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	var req provisionUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.PartyID == "" || req.Username == "" || req.Email == "" || req.InitialPassword == "" {
		http.Error(w, "party_id, username, email, initial_password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userUseCase.ProvisionUser(r.Context(), tenantID, req.PartyID, req.Username, req.Email, req.InitialPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
