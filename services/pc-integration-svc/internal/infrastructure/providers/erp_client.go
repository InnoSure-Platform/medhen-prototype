package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

type ERPClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
}

func NewERPClient(baseURL, apiKey string) *ERPClient {
	cbSettings := gobreaker.Settings{
		Name:        "ERPSync",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.5
		},
	}

	return &ERPClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cb: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

func (c *ERPClient) SyncJournalEntry(ctx context.Context, referenceID, transactionType string, amount float64, currency string) error {
	_, err := c.cb.Execute(func() (interface{}, error) {
		payload := map[string]interface{}{
			"referenceId":     referenceID,
			"transactionType": transactionType,
			"amount":          amount,
			"currency":        currency,
			"timestamp":       time.Now().Format(time.RFC3339),
		}

		bodyBytes, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/erp/journal/sync", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("erp gateway responded with status %d", resp.StatusCode)
		}

		return nil, nil
	})

	return err
}
