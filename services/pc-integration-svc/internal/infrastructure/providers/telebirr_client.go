package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-integration-svc/internal/application/ports"
	"github.com/medhen/pc-integration-svc/internal/domain"
	"github.com/sony/gobreaker"
)

type TelebirrClient struct {
	baseURL    string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
}

func NewTelebirrClient(baseURL string) ports.PaymentProvider {
	cbSettings := gobreaker.Settings{
		Name:        "Telebirr",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.5 // 50% failure trips it
		},
	}

	return &TelebirrClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cb: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

func (c *TelebirrClient) InitiatePayment(ctx context.Context, txn *domain.IntegrationTransaction) (string, error) {
	// Execute the HTTP call through the circuit breaker
	result, err := c.cb.Execute(func() (interface{}, error) {
		payload := map[string]interface{}{
			"nonce":           uuid.New().String(),
			"outTradeNo":      txn.InternalReferenceID.String(),
			"subject":         "Insurance Premium",
			"totalAmount":     fmt.Sprintf("%.2f", txn.Amount),
			"shortCode":       "123456",
			"returnUrl":       "https://medhen.com/callback",
		}

		bodyBytes, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/checkout", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		// In a real implementation, we would sign this payload here using OpenBao keys

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("telebirr responded with status %d", resp.StatusCode)
		}

		var respData struct {
			Data struct {
				ToPayUrl string `json:"toPayUrl"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
			return nil, err
		}

		return respData.Data.ToPayUrl, nil
	})

	if err != nil {
		return "", err
	}

	return result.(string), nil
}
