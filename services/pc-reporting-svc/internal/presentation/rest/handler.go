package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/medhen/pc-reporting-svc/internal/application/job"
	"github.com/medhen/pc-reporting-svc/internal/application/query"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("medhen.com/pc-reporting-svc/rest")

type KPIRESTHandler struct {
	kpiQuery   *query.KPIHandler
	workerPool *job.ReportWorkerPool
}

func NewKPIRESTHandler(kpiQuery *query.KPIHandler, workerPool *job.ReportWorkerPool) *KPIRESTHandler {
	return &KPIRESTHandler{
		kpiQuery:   kpiQuery,
		workerPool: workerPool,
	}
}

func (h *KPIRESTHandler) RegisterRoutes(r chi.Router, idempMiddleware func(http.Handler) http.Handler) {
	r.Route("/api/pc-reporting/v1", func(r chi.Router) {
		r.Get("/kpis/production", h.GetProductionKPIs)
		
		r.With(idempMiddleware).Post("/reports/jobs", h.SubmitReportJob)
	})
}

func (h *KPIRESTHandler) GetProductionKPIs(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetProductionKPIs")
	defer span.End()

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID", http.StatusBadRequest)
		return
	}

	lob := r.URL.Query().Get("lob")

	summary, err := h.kpiQuery.Handle(ctx, tenantID, lob)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *KPIRESTHandler) SubmitReportJob(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "SubmitReportJob")
	defer span.End()

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID", http.StatusBadRequest)
		return
	}

	type JobRequest struct {
		Year    int `json:"year"`
		Quarter int `json:"quarter"`
	}

	var req JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	jobID := uuid.NewString()

	h.workerPool.Submit(job.ReportJob{
		JobID:    jobID,
		TenantID: tenantID,
		Year:     req.Year,
		Quarter:  req.Quarter,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"job_id": jobID, "status": "accepted"})
}
