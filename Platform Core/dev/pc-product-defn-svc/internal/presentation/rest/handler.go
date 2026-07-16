package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/medhen/pc-product-defn-svc/internal/application/command"
)

type ProductHandler struct {
	createCmd *command.CreateProductHandler
}

func NewProductHandler(createCmd *command.CreateProductHandler) *ProductHandler {
	return &ProductHandler{
		createCmd: createCmd,
	}
}

func (h *ProductHandler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/v1/products", func(r chi.Router) {
		r.Post("/", h.CreateProduct)
	})
}

type CreateProductRequest struct {
	Code             string `json:"code"`
	LOB              string `json:"lob"`
	Name             string `json:"name"`
	RequireFairValue bool   `json:"require_fair_value"`
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// Extract tenant from headers (mocked for now)
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "missing X-Tenant-ID", http.StatusBadRequest)
		return
	}

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cmd := command.CreateProductCommand{
		TenantID:         tenantID,
		Code:             req.Code,
		LOB:              req.LOB,
		Name:             req.Name,
		RequireFairValue: req.RequireFairValue,
	}

	product, err := h.createCmd.Handle(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}
