package reporting

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-shared-go/httpx"
)

// MountReporting attaches read-optimized reporting routes.
func MountReporting(r chi.Router, repo store.Repository) {
	r.Route("/reporting", func(r chi.Router) {
		r.Get("/dashboards/executive-kpi", func(w http.ResponseWriter, req *http.Request) {
			// Stub: Executive KPIs built from CQRS projections
			httpx.WriteJSON(w, 200, map[string]any{
				"gwp_etb": 15000000,
				"policies_in_force": 450,
				"loss_ratio_percent": 65.5,
			})
		})
		r.Get("/reports/nbe-quarterly", func(w http.ResponseWriter, req *http.Request) {
			// Stub: NBE Regulatory returns
			httpx.WriteJSON(w, 200, map[string]any{
				"status": "ready",
				"report_data": map[string]string{},
			})
		})
	})
}
