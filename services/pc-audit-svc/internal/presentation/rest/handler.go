package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type AuditHandler struct {
	// queries *queries.AuditQueries // Stub for read model
}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{}
}

func (h *AuditHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/pc-audit/v1", func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		r.Get("/ledger/search", h.searchLedger)
		r.Get("/ledger/time-travel", h.timeTravel)
		r.Post("/exports", h.requestExport)
	})
}

func (h *AuditHandler) searchLedger(w http.ResponseWriter, r *http.Request) {
	// 1. Validate RBAC 'audit.read'
	// 2. Parse query params (actor_id, entity_id, date_range)
	// 3. Query HotLedger and/or ColdLake
	// 4. Perform dynamic PII masking if missing 'audit.pii.view'

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": []interface{}{},
		"meta": map[string]interface{}{"total": 0},
	})
}

func (h *AuditHandler) timeTravel(w http.ResponseWriter, r *http.Request) {
	// 1. Extract 'as_of' timestamp
	// 2. Query Iceberg ColdLakeRepository
	w.WriteHeader(http.StatusOK)
}

func (h *AuditHandler) requestExport(w http.ResponseWriter, r *http.Request) {
	// 1. Create ExportJob aggregate
	// 2. Queue for async processing
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"export_id": "job-123"})
}
