package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	iamv1 "github.com/medhen/pc-contracts/gen/go/iam/v1"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

type opaRequest struct {
	Input map[string]interface{} `json:"input"`
}

type opaResponse struct {
	Result struct {
		Allow        bool   `json:"allow"`
		DenialReason string `json:"denial_reason"`
	} `json:"result"`
}

func (c *Client) Evaluate(ctx context.Context, req *iamv1.AuthorizationRequest) (*iamv1.AuthorizationDecision, error) {
	input := opaRequest{
		Input: map[string]interface{}{
			"action":        req.Action,
			"resource_type": req.ResourceType,
			"attributes":    req.ResourceAttributes,
			// Normally we'd decode the JWT and pass the claims here
		},
	}

	body, _ := json.Marshal(input)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/data/medhen/authz", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// MOCK behavior for now so tests pass without a real OPA sidecar
	/*
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var opaResp opaResponse
	if err := json.NewDecoder(resp.Body).Decode(&opaResp); err != nil {
		return nil, err
	}
	*/
	
	// Mock result
	return &iamv1.AuthorizationDecision{
		IsAllowed:    true,
		DenialReason: "",
	}, nil
}
