package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-product-defn-svc/app"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContext struct {
	dbPool   *pgxpool.Pool
	handler  http.Handler
	tenantID string
	response *httptest.ResponseRecorder
}

func (tc *TestContext) iAmAnAuthenticatedProductManager() error {
	tc.tenantID = "TENANT-123"
	return nil
}

func (tc *TestContext) iSubmitACreateProductCommandFor(code string) error {
	reqBody := map[string]interface{}{
		"code":               code,
		"lob":                "MOTOR",
		"name":               "Integration Test Product",
		"require_fair_value": false,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer mock-valid-token")
	req.Header.Set("Content-Type", "application/json")

	tc.response = httptest.NewRecorder()
	tc.handler.ServeHTTP(tc.response, req)

	if tc.response.Code != http.StatusCreated {
		return fmt.Errorf("expected 201 Created, got %d. Body: %s", tc.response.Code, tc.response.Body.String())
	}
	return nil
}

func (tc *TestContext) theProductIsPersistedInTheState(state string) error {
	var rowStatus string
	query := `SELECT status FROM products LIMIT 1`
	err := tc.dbPool.QueryRow(context.Background(), query).Scan(&rowStatus)
	if err != nil {
		return err
	}
	if rowStatus != state {
		return fmt.Errorf("expected product to be in state %s, got %s", state, rowStatus)
	}
	return nil
}

func (tc *TestContext) aProductDraftCreatedEventIsPublishedTo(topic string) error {
	var outboxTopic string
	query := `SELECT topic FROM outbox LIMIT 1`
	err := tc.dbPool.QueryRow(context.Background(), query).Scan(&outboxTopic)
	if err != nil {
		return err
	}
	if outboxTopic != topic {
		return fmt.Errorf("expected event in outbox for topic %s, got %s", topic, outboxTopic)
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext, tc *TestContext) {
	ctx.Step(`^an authenticated Product Manager$`, tc.iAmAnAuthenticatedProductManager)
	ctx.Step(`^they submit a CreateProduct command for "([^"]*)"$`, tc.iSubmitACreateProductCommandFor)
	ctx.Step(`^the product is persisted in the "([^"]*)" state$`, tc.theProductIsPersistedInTheState)
	ctx.Step(`^a ProductDraftCreated event is published to "([^"]*)"$`, tc.aProductDraftCreatedEventIsPublishedTo)

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Clean up database after each scenario
		_, _ = tc.dbPool.Exec(context.Background(), "TRUNCATE products CASCADE; TRUNCATE outbox CASCADE;")
		return ctx, nil
	})
}

func TestFeatures(t *testing.T) {
	ctx := context.Background()

	// 1. Spin up Postgres using Testcontainers
	postgresContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("medhen_product"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer postgresContainer.Terminate(ctx)

	connString, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	dbPool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer dbPool.Close()

	// 2. Run DDL Migrations
	migrationSQL, err := os.ReadFile("../../services/pc-product-defn-svc/internal/infrastructure/postgres/migrations/000001_init_product_schema.up.sql")
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}
	if _, err := dbPool.Exec(ctx, string(migrationSQL)); err != nil {
		t.Fatalf("failed to execute migrations: %v", err)
	}

	// 3. Setup the Application
	handler := app.NewTestHandler(dbPool)

	tc := &TestContext{
		dbPool:  dbPool,
		handler: handler,
	}

	// 4. Run Godog Test Suite
	suite := godog.TestSuite{
		ScenarioInitializer: func(s *godog.ScenarioContext) {
			InitializeScenario(s, tc)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
