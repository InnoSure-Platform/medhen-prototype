package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// This mock server simulates Telebirr, CBE, Fayda, SMS, and ERP endpoints
// to be used in local dev or during chaos engineering tests.

func main() {
	r := chi.NewRouter()

	// Telebirr Mock Endpoint
	r.Post("/v1/checkout", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Chaos-Fail") == "true" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		response := map[string]interface{}{
			"code": "0",
			"msg":  "success",
			"data": map[string]string{
				"toPayUrl": "https://mock.telebirr.com/checkout/123",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Fayda Mock Endpoint
	r.Post("/v1/fayda/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Chaos-Fail") == "true" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		response := map[string]interface{}{
			"isVerified": true,
			"reason":     "",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// SMS Mock Endpoint
	r.Post("/v1/sms/send", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"messageId": "SMS-MOCK-789",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// ERP Journal Sync Mock Endpoint
	r.Post("/v1/erp/journal/sync", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("Starting mock provider server on :8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
