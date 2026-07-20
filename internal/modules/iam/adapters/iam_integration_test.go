package adapters_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/adapters"
	iamapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

func newService(t *testing.T) *iamapp.Service {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("medhen"), tcpostgres.WithUsername("postgres"), tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(30*time.Second)))
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	conn, _ := container.ConnectionString(ctx, "sslmode=disable")
	db, err := database.Connect(ctx, conn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(db.Close)
	if _, err := db.Pool().Exec(ctx, adapters.Schema); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return iamapp.NewService(adapters.NewUserRepository(db))
}

func TestRegisterAndResolveRoles(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	u, err := svc.Register(ctx, iamapp.RegisterInput{
		TenantID: "eic", Subject: "kc|abc", Email: "a@eic.et", FullName: "Adjuster A",
		Roles: []string{"claims", "staff"},
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if !u.HasRole("claims") {
		t.Fatal("expected claims role")
	}

	roles, err := svc.RolesForSubject(ctx, "eic", "kc|abc")
	if err != nil || len(roles) != 2 {
		t.Fatalf("roles = %v (%v), want 2", roles, err)
	}

	// Tenant isolation.
	if _, err := svc.RolesForSubject(ctx, "other", "kc|abc"); err != iamapp.ErrNotFound {
		t.Fatalf("cross-tenant lookup should be NotFound, got %v", err)
	}
}

func TestRegister_Validation(t *testing.T) {
	svc := newService(t)
	if _, err := svc.Register(context.Background(), iamapp.RegisterInput{TenantID: "eic", Roles: []string{"x"}}); err == nil {
		t.Fatal("expected error for missing subject")
	}
	if _, err := svc.Register(context.Background(), iamapp.RegisterInput{TenantID: "eic", Subject: "s"}); err == nil {
		t.Fatal("expected error for missing roles")
	}
}
