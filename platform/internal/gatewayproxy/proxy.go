// Package gatewayproxy implements the pc-gateway BFF that routes to pc-*-svc processes.
package gatewayproxy

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/InnoSure-Platform/pc-platform/internal/runtime"
)

// Enabled when MEDHEN_MESH=1 and downstream URLs are configured.
func Enabled() bool {
	return os.Getenv("MEDHEN_MESH") == "1"
}

// Mount registers reverse-proxy routes to microservices.
func Mount(r chi.Router) {
	r.Handle("/api/v1/parties", proxy("PARTY_URL", "http://localhost:8101"))
	r.Handle("/api/v1/parties/*", proxy("PARTY_URL", "http://localhost:8101"))
	r.Handle("/api/v1/products/*", proxy("POLICY_URL", "http://localhost:8103"))
	r.Handle("/api/v1/quotes", proxy("POLICY_URL", "http://localhost:8103"))
	r.Handle("/api/v1/quotes/*", proxy("POLICY_URL", "http://localhost:8103"))
	r.Handle("/api/v1/policies/*", proxy("POLICY_URL", "http://localhost:8103"))
	r.Handle("/api/v1/demo/kpis", proxy("POLICY_URL", "http://localhost:8103"))
	r.Handle("/api/v1/billing/*", proxy("BILLING_URL", "http://localhost:8107"))
	r.Handle("/api/v1/claims", proxy("CLAIMS_URL", "http://localhost:8106"))
	r.Handle("/api/v1/claims/*", proxy("CLAIMS_URL", "http://localhost:8106"))
	r.Handle("/api/v1/audit", proxy("AUDIT_URL", "http://localhost:8117"))
	r.Handle("/api/v1/fincrime/*", proxy("FINCRIME_URL", "http://localhost:8114"))
	r.Get("/api/v1/health", runtime.Health("pc-gateway"))
}

func proxy(envKey, fallback string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := os.Getenv(envKey)
		if base == "" {
			base = fallback
		}
		base = strings.TrimRight(base, "/")
		url := base + r.URL.Path
		if r.URL.RawQuery != "" {
			url += "?" + r.URL.RawQuery
		}
		var body io.Reader
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			body = bytes.NewReader(b)
		}
		req, err := http.NewRequestWithContext(r.Context(), r.Method, url, body)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		for k, vals := range r.Header {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		defer res.Body.Close()
		for k, vals := range res.Header {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(res.StatusCode)
		_, _ = io.Copy(w, res.Body)
	})
}
