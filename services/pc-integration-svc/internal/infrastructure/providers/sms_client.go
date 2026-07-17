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

type SMSClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
}

func NewSMSClient(baseURL, apiKey string) *SMSClient {
	cbSettings := gobreaker.Settings{
		Name:        "SMSGateway",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.5
		},
	}

	return &SMSClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cb: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

func (c *SMSClient) SendSMS(ctx context.Context, phoneNumber, message string) (bool, string, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		payload := map[string]string{
			"to":      phoneNumber,
			"message": message,
		}

		bodyBytes, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/sms/send", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("sms gateway responded with status %d", resp.StatusCode)
		}

		var respData struct {
			MessageID string `json:"messageId"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
			return nil, err
		}

		return respData.MessageID, nil
	})

	if err != nil {
		return false, "", err
	}

	return true, result.(string), nil
}
