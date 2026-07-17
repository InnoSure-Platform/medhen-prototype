package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hamba/avro/v2"
)

// ApicurioClient implements SchemaRegistryClient for Apicurio Registry.
type ApicurioClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewApicurioClient initializes a client for the Apicurio Schema Registry.
func NewApicurioClient(baseURL string) *ApicurioClient {
	return &ApicurioClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetSchema fetches and parses an Avro schema by its global ID from Apicurio.
func (c *ApicurioClient) GetSchema(id int) (avro.Schema, error) {
	// Apicurio v2 REST API to fetch schema content by global ID
	url := fmt.Sprintf("%s/apis/registry/v2/ids/globalIds/%d", c.baseURL, id)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var schemaData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&schemaData); err != nil {
		return nil, fmt.Errorf("failed to decode schema JSON: %w", err)
	}

	rawSchema, err := json.Marshal(schemaData)
	if err != nil {
		return nil, fmt.Errorf("failed to re-marshal schema: %w", err)
	}

	parsedSchema, err := avro.ParseBytes(rawSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse avro schema: %w", err)
	}

	return parsedSchema, nil
}
