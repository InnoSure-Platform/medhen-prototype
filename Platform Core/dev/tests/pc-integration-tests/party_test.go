package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-party-mgmt-svc/app"
)

func TestPartyRegistrationIntegration(t *testing.T) {
	// 1. Only run if we explicitly have infrastructure running, otherwise skip
	// to keep CI fast if it's purely a unit-test stage.
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test: set INTEGRATION_TEST=1 to run")
	}

	ctx := context.Background()

	// 2. Connect to local Postgres
	dbPool, err := pgxpool.New(ctx, "postgres://user:pass@localhost:5432/pc_party_db")
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer dbPool.Close()

	// 3. Connect to local Elasticsearch
	esClient, err := es.NewDefaultClient()
	if err != nil {
		t.Fatalf("failed to connect to es: %v", err)
	}

	// 4. Create HTTP test server with NewTestHandler
	handler := app.NewTestHandler(dbPool, esClient)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// 5. Send POST /api/pc-party-mgmt/v1/parties/individual
	reqBody := map[string]interface{}{
		"tenant_id":               "tenant-123",
		"first_name":              "Abebe",
		"last_name":               "Kebede",
		"date_of_birth":           time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC),
		"gender":                  "M",
		"national_id_type":        "fayda",
		"national_id_number":      "7788990011",
		"tin":                     "0011223344",
		"override_duplicate_flag": false,
	}

	b, _ := json.Marshal(reqBody)
	resp, err := http.Post(ts.URL+"/api/pc-party-mgmt/v1/parties/individual", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 6. Assert HTTP 201 Created
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d", resp.StatusCode)
	}

	t.Log("Party Registration Integration Test Passed")
}
