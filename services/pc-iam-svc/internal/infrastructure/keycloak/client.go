package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL  string
	user     string
	password string
	client   *http.Client
}

func NewClient(baseURL, user, password string) *Client {
	return &Client{
		baseURL:  baseURL,
		user:     user,
		password: password,
		client:   &http.Client{},
	}
}

// In a real implementation, we would acquire an admin token first.
// Here we mock the HTTP calls for the MVP structure.

func (c *Client) CreateRealm(ctx context.Context, realmName string) error {
	// e.g., POST /admin/realms
	payload := map[string]interface{}{
		"id":      realmName,
		"realm":   realmName,
		"enabled": true,
	}
	_, err := c.doRequest(ctx, "POST", "/admin/realms", payload)
	return err
}

func (c *Client) CreateUser(ctx context.Context, realmName, username, email, initialPassword string) (string, error) {
	// e.g., POST /admin/realms/{realm}/users
	payload := map[string]interface{}{
		"username": username,
		"email":    email,
		"enabled":  true,
		"credentials": []map[string]interface{}{
			{
				"type":      "password",
				"value":     initialPassword,
				"temporary": true,
			},
		},
	}
	
	path := fmt.Sprintf("/admin/realms/%s/users", realmName)
	_, err := c.doRequest(ctx, "POST", path, payload)
	if err != nil {
		return "", err
	}
	
	// Mock returning a generated UUID from Keycloak
	return "kc-user-uuid-mock", nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, payload interface{}) (*http.Response, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", "Bearer " + c.getAdminToken()) // Mocked
	
	// Mock successful response to allow the service to compile and run
	// return c.client.Do(req)
	return &http.Response{StatusCode: 201}, nil
}
