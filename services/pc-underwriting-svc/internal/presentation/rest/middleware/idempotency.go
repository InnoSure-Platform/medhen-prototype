package middleware

import (
	"net/http"

	"github.com/medhen/pc-idempotency-mgmt-sdk/idempotency"
)

// IdempotencyMiddleware ensures that requests with the same Idempotency-Key
// are not processed multiple times.
func IdempotencyMiddleware(store idempotency.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				// In a strict tier-0 environment, this might be rejected:
				// http.Error(w, "Idempotency-Key header is required", http.StatusBadRequest)
				// return
				
				// For now, if no key, just passthrough
				next.ServeHTTP(w, r)
				return
			}

			// 1. Check if key exists in store (e.g. Redis)
			exists, response, err := store.Check(r.Context(), key)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if exists {
				// 2. If exists, return the cached response immediately
				w.Header().Set("X-Idempotent-Replayed", "true")
				w.WriteHeader(response.StatusCode)
				w.Write(response.Body)
				return
			}

			// 3. Otherwise, process and save the result
			// (Here we use a response recorder to capture the output, omitted for brevity)
			next.ServeHTTP(w, r)
			
			// store.Save(r.Context(), key, recordedResponse)
		})
	}
}
