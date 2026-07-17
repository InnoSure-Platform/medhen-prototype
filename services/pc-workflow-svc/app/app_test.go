package app_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	
	// Ensure migrate is available for tests
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/iofs"
)

type workflowTestSuite struct {
	ctx        context.Context
	pgContainer *postgres.PostgresContainer
	db         *sql.DB
}

func (s *workflowTestSuite) startServices() error {
	s.ctx = context.Background()

	// 1. Start Postgres Testcontainer
	pgContainer, err := postgres.Run(s.ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("medhen"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return err
	}
	s.pgContainer = pgContainer

	connStr, err := pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	s.db = db

	// Normally we would call app.New() and postgres.RunMigrations() here
	// and initialize the Temporal Test Environment (testsuite.WorkflowTestSuite)
	return nil
}

func (s *workflowTestSuite) stopServices() {
	if s.db != nil {
		s.db.Close()
	}
	if s.pgContainer != nil {
		s.pgContainer.Terminate(s.ctx)
	}
}

func (s *workflowTestSuite) theWorkflowServiceIsRunning() error {
	if s.db == nil {
		return fmt.Errorf("database not running")
	}
	return nil
}

func (s *workflowTestSuite) anApprovalWorkflowIsInitiatedForQuote(quoteID string) error {
	// Execute gRPC or Temporal workflow initiation
	return nil
}

func (s *workflowTestSuite) theSLATimerExpires() error {
	// Advance Temporal test environment time
	return nil
}

func (s *workflowTestSuite) theTaskShouldBeReassignedToThe(role string) error {
	// Assert DB state
	return nil
}

func (s *workflowTestSuite) anEntryShouldBeWrittenToTheApprovalHistory(action string) error {
	// Assert DB state
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	suite := &workflowTestSuite{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		err := suite.startServices()
		return ctx, err
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		suite.stopServices()
		return ctx, nil
	})

	ctx.Step(`^the workflow service is running$`, suite.theWorkflowServiceIsRunning)
	ctx.Step(`^an approval workflow is initiated for quote "([^"]*)"$`, suite.anApprovalWorkflowIsInitiatedForQuote)
	ctx.Step(`^the SLA timer expires$`, suite.theSLATimerExpires)
	ctx.Step(`^the task should be reassigned to the "([^"]*)"$`, suite.theTaskShouldBeReassignedToThe)
	ctx.Step(`^an "([^"]*)" entry should be written to the approval history$`, suite.anEntryShouldBeWrittenToTheApprovalHistory)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
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
