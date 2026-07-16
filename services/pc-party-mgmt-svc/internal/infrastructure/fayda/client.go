package fayda

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	cb         *gobreaker.CircuitBreaker
}

// NewClient initializes the Fayda HTTP client with a built-in Circuit Breaker
func NewClient(baseURL string) *Client {
	// Configure Circuit Breaker
	var cbSettings gobreaker.Settings
	cbSettings.Name = "Fayda-ACL-HTTP"
	cbSettings.MaxRequests = 5                 // Max requests allowed to half-open state
	cbSettings.Interval = 10 * time.Second     // Cyclic period of the closed state to clear counts
	cbSettings.Timeout = 5 * time.Second       // Time after which state switches from open to half-open
	cbSettings.ReadyToTrip = func(counts gobreaker.Counts) bool {
		// Trip if error rate is > 50% over at least 10 requests
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 10 && failureRatio >= 0.5
	}
	cbSettings.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
		slog.Warn("Circuit Breaker state changed", "name", name, "from", from.String(), "to", to.String())
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 3 * time.Second, // Strict inner timeout for the HTTP calls
		},
		baseURL: baseURL,
		cb:      gobreaker.NewCircuitBreaker(cbSettings),
	}
}

func (c *Client) VerifyIdentity(ctx context.Context, nationalID string) (bool, error) {
	// Wrap the HTTP call in the Circuit Breaker Execute block
	result, err := c.cb.Execute(func() (interface{}, error) {
		url := fmt.Sprintf("%s/api/v1/verify/%s", c.baseURL, nationalID)
		
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return false, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return false, err // Returning error trips the breaker (if threshold is met)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			// A 5xx error from Fayda should count as a failure for the circuit breaker
			return false, fmt.Errorf("fayda internal server error: %d", resp.StatusCode)
		}

		if resp.StatusCode == 200 {
			var payload struct {
				Verified bool `json:"verified"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				return false, err
			}
			return payload.Verified, nil
		}

		return false, nil
	})

	if err != nil {
		// If the circuit breaker is open or the HTTP call failed, degrade gracefully.
		// Instead of crashing or blocking registration, we return (false, nil) indicating 
		// "Verification could not be definitively approved right now (fallback to PENDING)".
		slog.Error("Fayda verification degraded to fallback", "nationalID", nationalID, "error", err)
		return false, nil
	}

	return result.(bool), nil
}
