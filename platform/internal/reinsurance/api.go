package reinsurance

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-shared-go/httpx"
)

// MountReinsurance attaches reinsurance routes.
func MountReinsurance(r chi.Router, repo store.Repository) {
	r.Route("/reinsurance", func(r chi.Router) {
		r.Post("/treaties", func(w http.ResponseWriter, req *http.Request) {
			// Stub: Configure Treaty (Quota Share, Surplus)
			httpx.WriteJSON(w, 201, map[string]string{"status": "treaty_created", "id": "TR-001"})
		})
		r.Get("/cessions", func(w http.ResponseWriter, req *http.Request) {
			// Stub: Cession register / Bordereaux
			httpx.WriteJSON(w, 200, map[string]any{"cessions": []string{}})
		})
	})
}
