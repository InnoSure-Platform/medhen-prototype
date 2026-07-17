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

type FaydaClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
}

func NewFaydaClient(baseURL, apiKey string) *FaydaClient {
	cbSettings := gobreaker.Settings{
		Name:        "Fayda",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.5
		},
	}

	return &FaydaClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cb: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

func (c *FaydaClient) VerifyIdentity(ctx context.Context, faydaID, firstName, lastName, dob string) (bool, string, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		payload := map[string]string{
			"faydaId":   faydaID,
			"firstName": firstName,
			"lastName":  lastName,
			"dob":       dob,
		}

		bodyBytes, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/fayda/verify", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// A real implementation would use avast/retry-go here for transient network errors before returning
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("fayda responded with status %d", resp.StatusCode)
		}

		var respData struct {
			IsVerified bool   `json:"isVerified"`
			Reason     string `json:"reason"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
			return nil, err
		}

		return respData, nil
	})

	if err != nil {
		return false, "", err
	}

	type respType struct {
		IsVerified bool   `json:"isVerified"`
		Reason     string `json:"reason"`
	}
	res := result.(respType)
	return res.IsVerified, res.Reason, nil
}
