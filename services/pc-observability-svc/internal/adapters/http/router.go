package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
	
	"medhen.com/pc-observability-svc/internal/app"
	auth "github.com/medhen/pc-auth-sdk"
	idempotency "github.com/medhen/pc-idempotency-mgmt-sdk"
)

// Router sets up the REST API endpoints.
type Router struct {
	sloService *app.SLOService
}

// NewRouter creates a new HTTP router adapter.
func NewRouter(sloService *app.SLOService) *Router {
	return &Router{
		sloService: sloService,
	}
}

// SetupRoutes registers the HTTP handlers.
func (r *Router) SetupRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// Add OTel middleware for the chi router
	mux.Use(otelchi.Middleware("pc-observability-svc"))

	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(5 * time.Second))

	mux.Get("/health/liveness", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.Get("/health/readiness", func(w http.ResponseWriter, req *http.Request) {
		// Mock DB ping
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	// Auth SDK Integration
	authCfg := auth.ConfigFromEnv()
	authMiddleware := auth.Middleware(authCfg)

	// Idempotency SDK Integration
	idempCfg := idempotency.Config{RedisURL: "redis://localhost:6379"}
	idempMgr, err := idempotency.NewManager(idempCfg)
	var idempMiddleware func(http.Handler) http.Handler
	if err != nil {
		idempMiddleware = func(next http.Handler) http.Handler { return next }
	} else {
		idempMiddleware = idempMgr.Middleware
	}

	mux.Route("/api/pc-observability/v1", func(api chi.Router) {
		api.Use(authMiddleware)
		api.With(idempMiddleware).Post("/tenants", r.handleCreateTenant)
		api.With(idempMiddleware).Post("/slos", r.handleCreateSLO)
	})

	return mux
}

func (r *Router) handleCreateTenant(w http.ResponseWriter, req *http.Request) {
	// Stub implementation
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "tenant provisioning not implemented yet"})
}

type CreateSLORequest struct {
	TenantID         string  `json:"tenant_id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	TargetPercentage float64 `json:"target_percentage"`
	WindowDays       int     `json:"window_days"`
	SLIQuery         string  `json:"sli_query"`
	AlertPolicyID    string  `json:"alert_policy_id"`
}

func (r *Router) handleCreateSLO(w http.ResponseWriter, req *http.Request) {
	var payload CreateSLORequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slo, err := r.sloService.CreateSLO(
		req.Context(),
		payload.TenantID,
		payload.Name,
		payload.Description,
		payload.TargetPercentage,
		payload.WindowDays,
		payload.SLIQuery,
		payload.AlertPolicyID,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(slo)
}
