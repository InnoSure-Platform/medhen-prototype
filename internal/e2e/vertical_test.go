// Package e2e holds cross-module integration tests that wire several bounded
// contexts together (party + product + rating + underwriting + policy). They live
// outside any single module so they may legitimately import multiple modules.
package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	partyadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/adapters"
	partyapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/app"
	partydomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/domain"
	partyports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/ports"
	policyadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/adapters"
	policyapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/app"
	policydomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/domain"
	productadapters "github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/adapters"
	ratingdomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/domain"
	uwdomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
)

// partyReader adapts the party service to policy's consumed party port.
type partyReader struct{ svc *partyapp.Service }

func (r partyReader) GetParty(ctx context.Context, tenantID, id string) (partyports.PartyView, error) {
	p, err := r.svc.Get(ctx, tenantID, id)
	if err != nil {
		return partyports.PartyView{}, err
	}
	return partyports.PartyView{ID: p.ID, TenantID: p.TenantID, FullName: p.FullName}, nil
}

type vertical struct {
	db     *database.DB
	party  *partyapp.Service
	policy *policyapp.Service
	tenant string
}

func setup(t *testing.T) vertical {
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

	for _, ddl := range []string{outbox.Schema, partyadapters.Schema, productadapters.Schema, policyadapters.Schema} {
		if _, err := db.Pool().Exec(ctx, ddl); err != nil {
			t.Fatalf("schema: %v", err)
		}
	}

	// product catalog + rate provider
	productRepo := productadapters.NewProductRepository(db)
	if err := productRepo.Upsert(ctx, productadapters.MotorProduct()); err != nil {
		t.Fatalf("seed product: %v", err)
	}
	rateEngine := ratingdomain.NewEngine(productadapters.NewRateProvider(productRepo),
		ratingdomain.TaxPolicy{VATRate: decimal.NewFromFloat(0.15), StampDuty: money.FromInt(35)})

	partySvc := partyapp.NewService(db, partyadapters.NewPartyRepository(db))
	uwEngine := uwdomain.NewEngine(uwdomain.Rules{ReferAbove: money.FromInt(100000), MaxPriorClaims: 1})

	policySvc := policyapp.NewService(policyapp.Deps{
		DB: db, Quotes: policyadapters.NewQuoteRepository(db), Policies: policyadapters.NewPolicyRepository(db),
		Rating: rateEngine, Party: partyReader{svc: partySvc}, Underwriting: uwEngine, Insurer: "EIC",
	})

	return vertical{db: db, party: partySvc, policy: policySvc, tenant: "eic"}
}

func (v vertical) newParty(t *testing.T, nationalID string) string {
	t.Helper()
	id, err := v.party.Register(context.Background(), partyapp.RegisterInput{
		TenantID: v.tenant, FullName: "Abebe Bikila", PhoneE164: "+251911000000", NationalID: nationalID,
		Address: partydomain.Address{Region: "Addis Ababa", Zone: "Bole", Woreda: "03"},
	})
	if err != nil {
		t.Fatalf("register party: %v", err)
	}
	return id
}

func (v vertical) outboxCount(t *testing.T, topic string) int {
	t.Helper()
	var n int
	if err := v.db.Pool().QueryRow(context.Background(),
		`SELECT count(*) FROM outbox WHERE topic=$1`, topic).Scan(&n); err != nil {
		t.Fatalf("outbox count: %v", err)
	}
	return n
}

func TestVertical_QuoteBindIssuesPolicyAtomically(t *testing.T) {
	v := setup(t)
	ctx := context.Background()
	partyID := v.newParty(t, "ETH-1")

	// Quote — real rating (adult OD 1200 + TPL 800 = 2000 net, +VAT 300 +stamp 35 = 2335 gross).
	q, err := v.policy.CreateQuote(ctx, policyapp.CreateQuoteInput{
		TenantID: v.tenant, PartyID: partyID, ProductCode: "MOT",
		Coverages: []string{"OD", "TPL"}, RiskDimensions: map[string]string{"age_band": "adult"},
	})
	if err != nil {
		t.Fatalf("create quote: %v", err)
	}
	if q.GrossPremium.Minor() != 233500 {
		t.Fatalf("quote gross = %d, want 233500", q.GrossPremium.Minor())
	}
	if q.Status != policydomain.QuoteQuoted || q.CalculationID == "" {
		t.Fatalf("unexpected quote: %+v", q)
	}

	// Bind — underwriting auto-accepts, policy issued atomically.
	policy, err := v.policy.BindQuote(ctx, v.tenant, q.ID)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	wantNumber := fmt.Sprintf("EIC/MOT/%d/000001", time.Now().UTC().Year())
	if policy.PolicyNumber != wantNumber {
		t.Fatalf("policy number = %q, want %q", policy.PolicyNumber, wantNumber)
	}
	if policy.GrossPremium.Minor() != 233500 || policy.Status != policydomain.StatusIssued {
		t.Fatalf("unexpected policy: %+v", policy)
	}

	// Quote is now BOUND, and a policy.issued event was written in the same tx.
	gotQuote, _ := v.policy.GetQuote(ctx, v.tenant, q.ID)
	if gotQuote.Status != policydomain.QuoteBound {
		t.Fatalf("quote status = %s, want BOUND", gotQuote.Status)
	}
	if v.outboxCount(t, policydomain.TopicPolicyIssued) != 1 {
		t.Fatalf("expected 1 policy.issued outbox event, got %d", v.outboxCount(t, policydomain.TopicPolicyIssued))
	}
}

func TestVertical_RebindFailsAndIsAtomic(t *testing.T) {
	v := setup(t)
	ctx := context.Background()
	partyID := v.newParty(t, "ETH-2")

	q, _ := v.policy.CreateQuote(ctx, policyapp.CreateQuoteInput{
		TenantID: v.tenant, PartyID: partyID, ProductCode: "MOT", Coverages: []string{"OD"},
		RiskDimensions: map[string]string{"age_band": "adult"},
	})
	if _, err := v.policy.BindQuote(ctx, v.tenant, q.ID); err != nil {
		t.Fatalf("first bind: %v", err)
	}
	before := v.outboxCount(t, policydomain.TopicPolicyIssued)

	// Second bind must fail (already bound) and write nothing new.
	if _, err := v.policy.BindQuote(ctx, v.tenant, q.ID); err == nil {
		t.Fatal("expected rebind to fail")
	}
	if after := v.outboxCount(t, policydomain.TopicPolicyIssued); after != before {
		t.Fatalf("rebind changed outbox: before=%d after=%d", before, after)
	}
}

func TestVertical_PolicyNumberSequenceIncrements(t *testing.T) {
	v := setup(t)
	ctx := context.Background()
	year := time.Now().UTC().Year()

	for i := 1; i <= 3; i++ {
		partyID := v.newParty(t, fmt.Sprintf("ETH-SEQ-%d", i))
		q, _ := v.policy.CreateQuote(ctx, policyapp.CreateQuoteInput{
			TenantID: v.tenant, PartyID: partyID, ProductCode: "MOT", Coverages: []string{"OD"},
			RiskDimensions: map[string]string{"age_band": "adult"},
		})
		p, err := v.policy.BindQuote(ctx, v.tenant, q.ID)
		if err != nil {
			t.Fatalf("bind %d: %v", i, err)
		}
		want := fmt.Sprintf("EIC/MOT/%d/%06d", year, i)
		if p.PolicyNumber != want {
			t.Fatalf("policy #%d number = %q, want %q", i, p.PolicyNumber, want)
		}
	}
}

func TestVertical_ReferredRiskDoesNotIssue(t *testing.T) {
	v := setup(t)
	ctx := context.Background()
	partyID := v.newParty(t, "ETH-3")

	// prior_claims=2 exceeds the STP threshold → REFER, no policy.
	q, _ := v.policy.CreateQuote(ctx, policyapp.CreateQuoteInput{
		TenantID: v.tenant, PartyID: partyID, ProductCode: "MOT", Coverages: []string{"OD"},
		RiskDimensions: map[string]string{"age_band": "adult", "prior_claims": "2"},
	})
	if _, err := v.policy.BindQuote(ctx, v.tenant, q.ID); err != policydomain.ErrReferred {
		t.Fatalf("expected ErrReferred, got %v", err)
	}
	if v.outboxCount(t, policydomain.TopicPolicyIssued) != 0 {
		t.Fatal("referred risk must not issue a policy")
	}
}

func TestVertical_QuotePartyNotFound(t *testing.T) {
	v := setup(t)
	if _, err := v.policy.CreateQuote(context.Background(), policyapp.CreateQuoteInput{
		TenantID: v.tenant, PartyID: "ghost", ProductCode: "MOT", Coverages: []string{"OD"},
	}); err != policyapp.ErrPartyNotFound {
		t.Fatalf("expected ErrPartyNotFound, got %v", err)
	}
}
