package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/InnoSure-Platform/pc-platform/internal/integration"
	"github.com/InnoSure-Platform/pc-platform/internal/pdf"
	"github.com/InnoSure-Platform/pc-platform/internal/product"
	"github.com/InnoSure-Platform/pc-platform/internal/rating"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-platform/internal/underwriting"
	pcerr "github.com/InnoSure-Platform/pc-shared-go/errors"
	"github.com/InnoSure-Platform/pc-shared-go/events"
	"github.com/InnoSure-Platform/pc-shared-go/i18n"
	"github.com/InnoSure-Platform/pc-shared-go/idempotency"
	"github.com/InnoSure-Platform/pc-shared-go/tenant"
)

// Motor implements Phase 0 Motor BC use-cases (shared by monolith + mesh).
type Motor struct {
	Repo    store.Repository
	Idem    idempotencyStore
	Product product.Product
	Pay     integration.TelebirrClient
	SMS     integration.SMSClient
	PDF     *pdf.Generator
}

type idempotencyStore interface {
	Begin(ctx context.Context, scope, key string) ([]byte, bool, error)
	Complete(ctx context.Context, scope, key string, body any) error
}

func (m *Motor) audit(ctx context.Context, entityType, entityID, action, actor, detail string) {
	_ = m.Repo.AppendAudit(ctx, store.NewAuditEntry(entityType, entityID, action, actor, detail))
}

func (m *Motor) outbox(ctx context.Context, aggType, aggID, event string, payload any) {
	_ = m.Repo.InsertOutbox(ctx, aggType, aggID, event, payload)
}

type RegisterPartyCmd struct {
	FullName, FullNameAm, PhoneE164, Email string
	Address                                *store.Address
	Actor, IdemKey                         string
}

func (m *Motor) RegisterParty(ctx context.Context, cmd RegisterPartyCmd) (*store.Party, error) {
	if cmd.FullName == "" || cmd.PhoneE164 == "" {
		return nil, pcerr.E(pcerr.CodeValidation, "fullName and phoneE164 required")
	}
	if m.Idem != nil {
		if cached, replay, _ := m.Idem.Begin(ctx, "party.register", cmd.IdemKey); replay && len(cached) > 0 {
			var p store.Party
			_ = jsonUnmarshal(cached, &p)
			return &p, nil
		}
	}
	p := &store.Party{
		ID: uuid.NewString(), TenantID: tenant.EIC, FullName: cmd.FullName, FullNameAm: cmd.FullNameAm,
		PhoneE164: cmd.PhoneE164, Email: cmd.Email, Status: "ACTIVE", Address: cmd.Address, CreatedAt: time.Now().UTC(),
	}
	if err := m.Repo.SaveParty(ctx, p); err != nil {
		return nil, err
	}
	m.audit(ctx, "party", p.ID, "REGISTERED", cmd.Actor, p.FullName)
	m.outbox(ctx, "party", p.ID, events.PartyRegistered, map[string]string{"partyId": p.ID})
	if m.Idem != nil {
		_ = m.Idem.Complete(ctx, "party.register", cmd.IdemKey, p)
	}
	return p, nil
}

func (m *Motor) GetParty(ctx context.Context, id string) (*store.Party, error) {
	p, err := m.Repo.GetParty(ctx, id)
	if err == store.ErrNotFound {
		return nil, pcerr.E(pcerr.CodeNotFound, "party not found")
	}
	return p, err
}

type CreateQuoteCmd struct {
	PartyID, ProductCode, Actor, IdemKey string
	Risk                                 store.MotorRisk
	Locale                               i18n.Locale
}

func (m *Motor) CreateQuote(ctx context.Context, cmd CreateQuoteCmd) (*store.Quote, error) {
	if cmd.ProductCode == "" {
		cmd.ProductCode = m.Product.Code
	}
	if cmd.ProductCode != m.Product.Code {
		return nil, pcerr.E(pcerr.CodeValidation, "unknown product")
	}
	if _, err := m.GetParty(ctx, cmd.PartyID); err != nil {
		return nil, err
	}
	if cmd.Risk.Usage == "" {
		cmd.Risk.Usage = "private"
	}
	uw := underwriting.EvaluateSTP(cmd.Risk.Year, cmd.Risk.SumInsuredMinor, cmd.Risk.CoverType)
	if err := underwriting.RequireAccept(uw); err != nil {
		return nil, err
	}
	rated := rating.CalculateMotor(rating.Input{
		CoverType: cmd.Risk.CoverType, Usage: cmd.Risk.Usage, Year: cmd.Risk.Year,
		SumInsuredMinor: cmd.Risk.SumInsuredMinor, Locale: cmd.Locale,
	})
	q := &store.Quote{
		ID: uuid.NewString(), TenantID: tenant.EIC, PartyID: cmd.PartyID, ProductCode: cmd.ProductCode,
		Status: "QUOTED", Risk: cmd.Risk, Lines: rated.Lines, TotalMinor: rated.TotalMinor,
		Currency: rated.Currency, UWDecision: uw.Outcome, ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	if err := m.Repo.SaveQuote(ctx, q); err != nil {
		return nil, err
	}
	m.audit(ctx, "quote", q.ID, "QUOTED", cmd.Actor, fmt.Sprintf("total=%d %s", q.TotalMinor, q.Currency))
	m.outbox(ctx, "quote", q.ID, events.PolicyQuoted, map[string]any{"quoteId": q.ID, "totalMinor": q.TotalMinor})
	return q, nil
}

func (m *Motor) GetQuote(ctx context.Context, id string) (*store.Quote, error) {
	q, err := m.Repo.GetQuote(ctx, id)
	if err == store.ErrNotFound {
		return nil, pcerr.E(pcerr.CodeNotFound, "quote not found")
	}
	return q, err
}

type BindResult struct {
	Policy  *store.Policy  `json:"policy"`
	Invoice *store.Invoice `json:"invoice"`
}

func (m *Motor) BindQuote(ctx context.Context, quoteID, actor, idemKey string) (*BindResult, error) {
	q, err := m.GetQuote(ctx, quoteID)
	if err != nil {
		return nil, err
	}
	if q.Status != "QUOTED" {
		return nil, pcerr.E(pcerr.CodeConflict, "quote not in QUOTED status")
	}
	now := time.Now().UTC()
	pol := &store.Policy{
		ID: uuid.NewString(), TenantID: tenant.EIC, PolicyNumber: "PENDING", QuoteID: q.ID, PartyID: q.PartyID,
		ProductCode: q.ProductCode, Status: "PENDING_PAYMENT", Risk: q.Risk, Lines: q.Lines,
		TotalMinor: q.TotalMinor, Currency: q.Currency,
		EffectiveFrom: now.Format("2006-01-02"), EffectiveTo: now.AddDate(1, 0, -1).Format("2006-01-02"),
	}
	inv := &store.Invoice{ID: uuid.NewString(), TenantID: tenant.EIC, PolicyID: pol.ID, AmountMinor: pol.TotalMinor, Currency: pol.Currency, Status: "OPEN"}
	pol.InvoiceID = inv.ID
	if err := m.Repo.UpdateQuoteStatus(ctx, q.ID, "BOUND"); err != nil {
		return nil, err
	}
	if err := m.Repo.SavePolicy(ctx, pol); err != nil {
		return nil, err
	}
	if err := m.Repo.SaveInvoice(ctx, inv); err != nil {
		return nil, err
	}
	m.audit(ctx, "policy", pol.ID, "BOUND", actor, "awaiting payment")
	m.outbox(ctx, "policy", pol.ID, events.PolicyBound, map[string]string{"policyId": pol.ID})
	return &BindResult{Policy: pol, Invoice: inv}, nil
}

type PaymentResult struct {
	ReceiptID string           `json:"receiptId"`
	Channel   string           `json:"channel"`
	Status    string           `json:"status"`
	Policy    *store.Policy    `json:"policy"`
	Documents []store.Document `json:"documents"`
}

func (m *Motor) PayInvoice(ctx context.Context, invoiceID, channel, phone, actor, idemKey string) (*PaymentResult, error) {
	if channel == "" {
		channel = "telebirr"
	}
	inv, err := m.Repo.GetInvoice(ctx, invoiceID)
	if err == store.ErrNotFound {
		return nil, pcerr.E(pcerr.CodeNotFound, "invoice not found")
	}
	if err != nil {
		return nil, err
	}
	if inv.Status != "OPEN" {
		return nil, pcerr.E(pcerr.CodeConflict, "invoice not open")
	}
	pol, err := m.Repo.GetPolicy(ctx, inv.PolicyID)
	if err != nil {
		return nil, err
	}
	party, _ := m.Repo.GetParty(ctx, pol.PartyID)
	if phone == "" && party != nil {
		phone = party.PhoneE164
	}
	receiptRef, err := m.Pay.Charge(phone, inv.AmountMinor, inv.ID)
	if err != nil {
		return nil, pcerr.Wrap(pcerr.CodePaymentFailed, "telebirr charge failed", err)
	}
	pn, err := m.Repo.NextPolicyNumber(ctx, time.Now().UTC().Year())
	if err != nil {
		return nil, err
	}
	pol.PolicyNumber = pn
	pol.Status = "ISSUED"
	now := time.Now().UTC()
	pol.IssuedAt = &now
	inv.Status = "PAID"
	rec := &store.Receipt{ID: receiptRef, InvoiceID: inv.ID, Channel: channel, Status: "COMPLETED", PaidAt: now}
	if err := m.Repo.SaveReceipt(ctx, rec); err != nil {
		return nil, err
	}
	if err := m.Repo.SavePolicy(ctx, pol); err != nil {
		return nil, err
	}
	if err := m.Repo.SaveInvoice(ctx, inv); err != nil {
		return nil, err
	}
	var docs []store.Document
	if m.PDF != nil {
		docs, err = m.PDF.Pack(pol, party)
		if err != nil {
			return nil, err
		}
		for i := range docs {
			if err := m.Repo.SaveDocument(ctx, &docs[i]); err != nil {
				return nil, err
			}
		}
	}
	if m.SMS != nil && party != nil {
		msg := i18n.T("policy.issued", i18n.EN) + " " + pol.PolicyNumber
		_ = m.SMS.Send(party.PhoneE164, msg)
	}
	m.audit(ctx, "payment", rec.ID, "COMPLETED", actor, channel)
	m.audit(ctx, "policy", pol.ID, "ISSUED", actor, pol.PolicyNumber)
	m.outbox(ctx, "policy", pol.ID, events.PolicyIssued, map[string]any{"policyId": pol.ID, "policyNumber": pol.PolicyNumber})
	return &PaymentResult{ReceiptID: rec.ID, Channel: channel, Status: "COMPLETED", Policy: pol, Documents: docs}, nil
}

func (m *Motor) GetPolicy(ctx context.Context, id string) (*store.Policy, error) {
	p, err := m.Repo.GetPolicy(ctx, id)
	if err == store.ErrNotFound {
		return nil, pcerr.E(pcerr.CodeNotFound, "policy not found")
	}
	return p, err
}

func (m *Motor) ListPolicyDocuments(ctx context.Context, policyID string) ([]store.Document, error) {
	if _, err := m.GetPolicy(ctx, policyID); err != nil {
		return nil, err
	}
	return m.Repo.ListDocumentsByPolicy(ctx, policyID)
}

type FNOLCmd struct {
	PolicyID, Description, Actor, IdemKey string
	LossDate                              time.Time
	Latitude, Longitude                   float64
	EstimatedAmountMinor                  int64
	PhotoObjectKeys                       []string
}

func (m *Motor) SubmitFNOL(ctx context.Context, cmd FNOLCmd) (*store.Claim, error) {
	pol, err := m.GetPolicy(ctx, cmd.PolicyID)
	if err != nil {
		return nil, err
	}
	if pol.Status != "ISSUED" {
		return nil, pcerr.E(pcerr.CodeValidation, "policy not active/issued")
	}
	track := "STANDARD"
	if cmd.EstimatedAmountMinor > 0 && cmd.EstimatedAmountMinor <= 5_000_000 {
		track = "FAST_TRACK"
	}
	cn, err := m.Repo.NextClaimNumber(ctx, time.Now().UTC().Year())
	if err != nil {
		return nil, err
	}
	cl := &store.Claim{
		ID: uuid.NewString(), ClaimNumber: cn, TenantID: tenant.EIC, PolicyID: pol.ID,
		Status: "REGISTERED", Track: track, Description: cmd.Description,
		Latitude: cmd.Latitude, Longitude: cmd.Longitude, EstimatedAmountMinor: cmd.EstimatedAmountMinor,
		Currency: "ETB", PhotoObjectKeys: cmd.PhotoObjectKeys, CreatedAt: time.Now().UTC(),
	}
	if err := m.Repo.SaveClaim(ctx, cl); err != nil {
		return nil, err
	}
	m.audit(ctx, "claim", cl.ID, "REGISTERED", cmd.Actor, cn+" "+track)
	m.outbox(ctx, "claim", cl.ID, events.ClaimRegistered, map[string]string{"claimId": cl.ID})
	return cl, nil
}

func (m *Motor) SettleFastTrack(ctx context.Context, claimID, actor, idemKey string) (*store.Claim, error) {
	cl, err := m.Repo.GetClaim(ctx, claimID)
	if err == store.ErrNotFound {
		return nil, pcerr.E(pcerr.CodeNotFound, "claim not found")
	}
	if err != nil {
		return nil, err
	}
	if cl.Status == "SETTLED" {
		return cl, nil
	}
	if cl.Track != "FAST_TRACK" {
		return nil, pcerr.E(pcerr.CodeValidation, "claim not on fast-track")
	}
	cl.Status = "SETTLED"
	cl.SettlementMinor = cl.EstimatedAmountMinor
	now := time.Now().UTC()
	cl.SettledAt = &now
	if err := m.Repo.SaveClaim(ctx, cl); err != nil {
		return nil, err
	}
	if pol, err := m.Repo.GetPolicy(ctx, cl.PolicyID); err == nil {
		if party, err := m.Repo.GetParty(ctx, pol.PartyID); err == nil && m.SMS != nil {
			_ = m.SMS.Send(party.PhoneE164, i18n.T("claim.settled", i18n.EN))
		}
	}
	m.audit(ctx, "claim", cl.ID, "SETTLED", actor, fmt.Sprintf("%d", cl.SettlementMinor))
	m.outbox(ctx, "claim", cl.ID, events.ClaimSettled, map[string]any{"claimId": cl.ID})
	return cl, nil
}

func (m *Motor) QueryAudit(ctx context.Context, entityType, entityID string, limit int) ([]store.AuditEntry, error) {
	return m.Repo.QueryAudit(ctx, entityType, entityID, limit)
}

type KPIs struct {
	PoliciesInForce int    `json:"policiesInForce"`
	GWPMinor        int64  `json:"gwpMinor"`
	Currency        string `json:"currency"`
	ClaimsOpen      int    `json:"claimsOpen"`
	ClaimsSettled   int    `json:"claimsSettled"`
	TenantID        string `json:"tenantId"`
	ProductCode     string `json:"productCode"`
}

func (m *Motor) KPIs(ctx context.Context) (KPIs, error) {
	k := KPIs{Currency: "ETB", TenantID: tenant.EIC, ProductCode: m.Product.Code}
	pols, err := m.Repo.ListIssuedPolicies(ctx)
	if err != nil {
		return k, err
	}
	for _, p := range pols {
		k.PoliciesInForce++
		k.GWPMinor += p.TotalMinor
	}
	claims, err := m.Repo.ListClaims(ctx)
	if err != nil {
		return k, err
	}
	for _, c := range claims {
		if c.Status == "SETTLED" {
			k.ClaimsSettled++
		} else {
			k.ClaimsOpen++
		}
	}
	return k, nil
}

func NewDefault() *Motor {
	return &Motor{
		Repo:    store.NewMemoryRepository(),
		Idem:    idempotency.NewMemoryStore(),
		Product: product.SeedMotor(),
		Pay:     integration.NewTelebirrFromEnv(),
		SMS:     &integration.MockSMS{},
		PDF:     pdf.NewGenerator("", "/files"),
	}
}
