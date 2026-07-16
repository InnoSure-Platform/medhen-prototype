package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/medhen/pc-auth-sdk"
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

func (h *ProductHandler) RegisterRoutes(r chi.Router, idempMiddleware func(http.Handler) http.Handler) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/v1/products", func(r chi.Router) {
		if idempMiddleware != nil {
			r.With(idempMiddleware).Post("/", h.CreateProduct)
		} else {
			r.Post("/", h.CreateProduct)
		}
	})
}

type CreateProductRequest struct {
	Code             string `json:"code" validate:"required,min=3"`
	LOB              string `json:"lob" validate:"required"`
	Name             string `json:"name" validate:"required"`
	RequireFairValue bool   `json:"require_fair_value"`
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// Extract tenant from context securely injected by auth middleware
	tenantID, err := auth.GetTenantID(r.Context())
	if err != nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteProblem(w, http.StatusBadRequest, "Invalid JSON", "Could not parse request body")
		return
	}

	// Validate API input using go-playground/validator
	if validationErrs := ValidateStruct(req); validationErrs != nil {
		WriteValidationProblem(w, validationErrs)
		return
	}

	// 2. Map to command model
	cmd := command.CreateProductCommand{
		TenantID:         tenantID,
		Code:             req.Code,
		LOB:              req.LOB,
		Name:             req.Name,
		RequireFairValue: req.RequireFairValue,
	}

	// 3. Dispatch to Application Layer
	product, err := h.createCmd.Handle(r.Context(), cmd)
	if err != nil {
		// Log the internal error securely
		WriteProblem(w, http.StatusInternalServerError, "Internal Server Error", "An error occurred processing your request")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}
