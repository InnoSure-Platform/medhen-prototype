package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/medhen/pc-iam-svc/internal/domain"
)

// SCIMHandler handles standard SCIM 2.0 protocol endpoints
type SCIMHandler struct {
	// scimRepo domain.SCIMRepository
}

func NewSCIMHandler() *SCIMHandler {
	return &SCIMHandler{}
}

func (h *SCIMHandler) RegisterRoutes(router chi.Router) {
	// SCIM 2.0 typically routes under /scim/v2
	router.Route("/scim/v2/{tenantID}", func(r chi.Router) {
		r.Post("/Users", h.CreateUser)
		r.Get("/Users/{id}", h.GetUser)
		r.Put("/Users/{id}", h.UpdateUser)
		r.Delete("/Users/{id}", h.DeleteUser)
	})
}

func (h *SCIMHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// tenantID := chi.URLParam(r, "tenantID")
	var req domain.SCIMUser
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Mocking SCIM Provisioning (would typically delegate to Keycloak & DB)
	req.ID = "mock-scim-uuid"
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func (h *SCIMHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *SCIMHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": id, "status": "updated"})
}

func (h *SCIMHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
