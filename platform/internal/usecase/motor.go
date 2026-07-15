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
	Fayda   integration.FaydaClient
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

func (m *Motor) VerifyKYC(ctx context.Context, partyID, faydaID, actor string) error {
	p, err := m.Repo.GetParty(ctx, partyID)
	if err != nil {
		return err
	}
	if p.KYCStatus == "VERIFIED" {
		return nil
	}
	profile, err := m.Fayda.Verify(faydaID)
	if err != nil || profile == nil {
		return pcerr.E(pcerr.CodeValidation, "invalid fayda ID")
	}
	if profile.Status != "ACTIVE" {
		return pcerr.E(pcerr.CodeValidation, "fayda profile inactive")
	}
	p.FaydaID = faydaID
	p.KYCStatus = "VERIFIED"
	// We might also update the name based on verified data
	p.FullName = profile.FullName

	if err := m.Repo.SaveParty(ctx, p); err != nil {
		return err
	}

	m.audit(ctx, "party", p.ID, "KYC_VERIFIED", actor, faydaID)
	// Optionally emit outbox event here
	return nil
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
	if uw.Outcome == "DECLINE" {
		return nil, pcerr.E(pcerr.CodeUWDeclined, uw.Reason)
	}

	rated := rating.CalculateMotor(rating.Input{
		CoverType: cmd.Risk.CoverType, Usage: cmd.Risk.Usage, Year: cmd.Risk.Year,
		SumInsuredMinor: cmd.Risk.SumInsuredMinor, Locale: cmd.Locale,
	})

	status := "QUOTED"
	if uw.Outcome == "REFER" {
		status = "REFERRED"
	}

	q := &store.Quote{
		ID: uuid.NewString(), TenantID: tenant.EIC, PartyID: cmd.PartyID, ProductCode: cmd.ProductCode,
		Status: status, Risk: cmd.Risk, Lines: rated.Lines, TotalMinor: rated.TotalMinor,
		Currency: rated.Currency, UWDecision: uw.Outcome, ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	if err := m.Repo.SaveQuote(ctx, q); err != nil {
		return nil, err
	}
	m.audit(ctx, "quote", q.ID, status, cmd.Actor, fmt.Sprintf("total=%d %s, uw=%s", q.TotalMinor, q.Currency, uw.Outcome))
	if status == "QUOTED" {
		m.outbox(ctx, "quote", q.ID, events.PolicyQuoted, map[string]any{"quoteId": q.ID, "totalMinor": q.TotalMinor})
	}
	return q, nil
}

func (m *Motor) ApproveQuote(ctx context.Context, quoteID, actor string) error {
	q, err := m.Repo.GetQuote(ctx, quoteID)
	if err != nil {
		return err
	}
	if q.Status != "REFERRED" {
		return pcerr.E(pcerr.CodeConflict, "quote is not REFERRED")
	}
	if err := m.Repo.UpdateQuoteStatus(ctx, quoteID, "QUOTED"); err != nil {
		return err
	}
	m.audit(ctx, "quote", quoteID, "APPROVED", actor, "underwriter approved referral")
	m.outbox(ctx, "quote", quoteID, events.PolicyQuoted, map[string]any{"quoteId": quoteID, "totalMinor": q.TotalMinor})
	return nil
}

func (m *Motor) DeclineQuote(ctx context.Context, quoteID, actor string) error {
	q, err := m.Repo.GetQuote(ctx, quoteID)
	if err != nil {
		return err
	}
	if q.Status != "REFERRED" {
		return pcerr.E(pcerr.CodeConflict, "quote is not REFERRED")
	}
	if err := m.Repo.UpdateQuoteStatus(ctx, quoteID, "DECLINED"); err != nil {
		return err
	}
	m.audit(ctx, "quote", quoteID, "DECLINED", actor, "underwriter declined referral")
	return nil
}

func (m *Motor) ListReferredQuotes(ctx context.Context) ([]*store.Quote, error) {
	return m.Repo.ListQuotesByStatus(ctx, "REFERRED")
}

func (m *Motor) GetQuote(ctx context.Context, id string) (*store.Quote, error) {
	q, err := m.Repo.GetQuote(ctx, id)
	if err == store.ErrNotFound {
		return nil, pcerr.E(pcerr.CodeNotFound, "quote not found")
	}
	return q, err
}

type BindResult struct {
	Policy   *store.Policy    `json:"policy"`
	Invoice  *store.Invoice   `json:"invoice"`
	Invoices []*store.Invoice `json:"invoices"`
}

func (m *Motor) BindQuote(ctx context.Context, quoteID, installmentPlan, actor, idemKey string) (*BindResult, error) {
	q, err := m.GetQuote(ctx, quoteID)
	if err != nil {
		return nil, err
	}
	if q.Status != "QUOTED" {
		return nil, pcerr.E(pcerr.CodeConflict, "quote is not in QUOTED status")
	}

	p, err := m.GetParty(ctx, q.PartyID)
	if err != nil {
		return nil, err
	}
	if p.KYCStatus != "VERIFIED" {
		return nil, pcerr.E(pcerr.CodeConflict, "kyc verification is required before binding")
	}
	now := time.Now().UTC()
	pol := &store.Policy{
		ID: uuid.NewString(), TenantID: tenant.EIC, PolicyNumber: "PENDING", QuoteID: q.ID, PartyID: q.PartyID,
		ProductCode: q.ProductCode, Status: "PENDING_PAYMENT", Risk: q.Risk, Lines: q.Lines,
		TotalMinor: q.TotalMinor, Currency: q.Currency,
		EffectiveFrom: now.Format("2006-01-02"), EffectiveTo: now.AddDate(1, 0, -1).Format("2006-01-02"),
	}
	var invoices []*store.Invoice
	if installmentPlan == "40_30_30" {
		down := (pol.TotalMinor * 40) / 100
		p2 := (pol.TotalMinor * 30) / 100
		p3 := pol.TotalMinor - down - p2 // remainder
		
		invoices = append(invoices, &store.Invoice{ID: uuid.NewString(), TenantID: tenant.EIC, PolicyID: pol.ID, AmountMinor: down, Currency: pol.Currency, Status: "OPEN", DueDate: now.Format("2006-01-02"), InstallmentNumber: 1})
		invoices = append(invoices, &store.Invoice{ID: uuid.NewString(), TenantID: tenant.EIC, PolicyID: pol.ID, AmountMinor: p2, Currency: pol.Currency, Status: "OPEN", DueDate: now.AddDate(0, 1, 0).Format("2006-01-02"), InstallmentNumber: 2})
		invoices = append(invoices, &store.Invoice{ID: uuid.NewString(), TenantID: tenant.EIC, PolicyID: pol.ID, AmountMinor: p3, Currency: pol.Currency, Status: "OPEN", DueDate: now.AddDate(0, 2, 0).Format("2006-01-02"), InstallmentNumber: 3})
	} else {
		// Default to 100_UPFRONT
		invoices = append(invoices, &store.Invoice{ID: uuid.NewString(), TenantID: tenant.EIC, PolicyID: pol.ID, AmountMinor: pol.TotalMinor, Currency: pol.Currency, Status: "OPEN", DueDate: now.Format("2006-01-02"), InstallmentNumber: 1})
	}

	pol.InvoiceID = invoices[0].ID
	if err := m.Repo.UpdateQuoteStatus(ctx, q.ID, "BOUND"); err != nil {
		return nil, err
	}
	if err := m.Repo.SavePolicy(ctx, pol); err != nil {
		return nil, err
	}
	for _, inv := range invoices {
		if err := m.Repo.SaveInvoice(ctx, inv); err != nil {
			return nil, err
		}
	}
	m.audit(ctx, "policy", pol.ID, "BOUND", actor, "awaiting downpayment")
	m.outbox(ctx, "policy", pol.ID, events.PolicyBound, map[string]string{"policyId": pol.ID})
	return &BindResult{Policy: pol, Invoice: invoices[0], Invoices: invoices}, nil
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
	now := time.Now().UTC()
	inv.Status = "PAID"
	rec := &store.Receipt{ID: receiptRef, InvoiceID: inv.ID, Channel: channel, Status: "COMPLETED", PaidAt: now}
	if err := m.Repo.SaveReceipt(ctx, rec); err != nil {
		return nil, err
	}
	if err := m.Repo.SaveInvoice(ctx, inv); err != nil {
		return nil, err
	}

	if inv.InstallmentNumber <= 1 {
		pn, err := m.Repo.NextPolicyNumber(ctx, time.Now().UTC().Year())
		if err != nil {
			return nil, err
		}
		pol.PolicyNumber = pn
		pol.Status = "ISSUED"
		pol.IssuedAt = &now
		if err := m.Repo.SavePolicy(ctx, pol); err != nil {
			return nil, err
		}
	} else {
		// Log that a subsequent installment was paid
		m.audit(ctx, "policy", pol.ID, "INSTALLMENT_PAID", actor, fmt.Sprintf("installment #%d paid", inv.InstallmentNumber))
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

func (m *Motor) AdjustReserve(ctx context.Context, claimID, actor string, amountMinor int64) error {
	cl, err := m.Repo.GetClaim(ctx, claimID)
	if err != nil {
		return err
	}
	if cl.Status == "SETTLED" {
		return pcerr.E(pcerr.CodeValidation, "cannot adjust reserve on settled claim")
	}
	oldReserve := cl.ReserveMinor
	cl.ReserveMinor = amountMinor
	if err := m.Repo.SaveClaim(ctx, cl); err != nil {
		return err
	}
	m.audit(ctx, "claim", cl.ID, "RESERVE_ADJUSTED", actor, fmt.Sprintf("reserve changed from %d to %d", oldReserve, cl.ReserveMinor))
	return nil
}

func (m *Motor) RecordRecovery(ctx context.Context, claimID, actor string, amountMinor int64) error {
	cl, err := m.Repo.GetClaim(ctx, claimID)
	if err != nil {
		return err
	}
	cl.RecoveryMinor += amountMinor
	if err := m.Repo.SaveClaim(ctx, cl); err != nil {
		return err
	}
	m.audit(ctx, "claim", cl.ID, "RECOVERY_RECORDED", actor, fmt.Sprintf("recorded recovery of %d", amountMinor))
	return nil
}

func (m *Motor) SettleClaim(ctx context.Context, claimID, actor, idemKey string, settlementMinor int64) (*store.Claim, error) {
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

	pol, err := m.Repo.GetPolicy(ctx, cl.PolicyID)
	if err != nil {
		return nil, err
	}

	if cl.Track == "TOTAL_LOSS" && settlementMinor > pol.Risk.SumInsuredMinor {
		return nil, pcerr.E(pcerr.CodeValidation, "settlement exceeds sum insured for total loss")
	}

	cl.Status = "SETTLED"
	cl.SettlementMinor = settlementMinor
	now := time.Now().UTC()
	cl.SettledAt = &now
	if err := m.Repo.SaveClaim(ctx, cl); err != nil {
		return nil, err
	}
	
	if cl.Track == "TOTAL_LOSS" {
		_ = m.CancelPolicy(ctx, CancelPolicyCmd{
			PolicyID: pol.ID,
			Actor:    actor,
			Reason:   "Total Loss Exhaustion",
		})
	}

	if party, err := m.Repo.GetParty(ctx, pol.PartyID); err == nil && m.SMS != nil {
		_ = m.SMS.Send(party.PhoneE164, i18n.T("claim.settled", i18n.EN))
	}
	m.audit(ctx, "claim", cl.ID, "SETTLED", actor, fmt.Sprintf("%d", cl.SettlementMinor))
	m.outbox(ctx, "claim", cl.ID, events.ClaimSettled, map[string]any{"claimId": cl.ID})
	return cl, nil
}

type EndorsePolicyCmd struct {
	PolicyID string
	Risk     store.MotorRisk
	Actor    string
	IdemKey  string
}

func (m *Motor) EndorsePolicy(ctx context.Context, cmd EndorsePolicyCmd) (*store.Policy, error) {
	old, err := m.Repo.GetPolicy(ctx, cmd.PolicyID)
	if err != nil {
		return nil, err
	}
	if old.Status != "ISSUED" {
		return nil, pcerr.E(pcerr.CodeConflict, "can only endorse ISSUED policy")
	}

	uw := underwriting.EvaluateSTP(cmd.Risk.Year, cmd.Risk.SumInsuredMinor, cmd.Risk.CoverType)
	if err := underwriting.RequireAccept(uw); err != nil {
		return nil, err
	}

	rated := rating.CalculateMotor(rating.Input{
		CoverType: cmd.Risk.CoverType, Usage: cmd.Risk.Usage, Year: cmd.Risk.Year,
		SumInsuredMinor: cmd.Risk.SumInsuredMinor, Locale: i18n.EN,
	})

	// Simplistic pro-rata logic for demo: diff of totals
	diff := rated.TotalMinor - old.TotalMinor

	now := time.Now().UTC()
	newPol := &store.Policy{
		ID:             uuid.NewString(),
		TenantID:       tenant.EIC,
		PolicyNumber:   old.PolicyNumber,
		QuoteID:        old.QuoteID,
		PartyID:        old.PartyID,
		ProductCode:    old.ProductCode,
		Status:         "ISSUED",
		Risk:           cmd.Risk,
		Lines:          rated.Lines,
		TotalMinor:     rated.TotalMinor,
		Currency:       rated.Currency,
		EffectiveFrom:  old.EffectiveFrom,
		EffectiveTo:    old.EffectiveTo,
		IssuedAt:       &now,
		ParentPolicyID: old.ID,
		Version:        old.Version + 1,
	}

	// Create invoice for difference if > 0
	if diff > 0 {
		inv := &store.Invoice{
			ID:          uuid.NewString(),
			TenantID:    tenant.EIC,
			PolicyID:    newPol.ID,
			AmountMinor: diff,
			Currency:    newPol.Currency,
			Status:      "OPEN",
		}
		if err := m.Repo.SaveInvoice(ctx, inv); err != nil {
			return nil, err
		}
		newPol.InvoiceID = inv.ID
	}

	old.Status = "SUPERSEDED"
	if err := m.Repo.SavePolicy(ctx, old); err != nil {
		return nil, err
	}
	if err := m.Repo.SavePolicy(ctx, newPol); err != nil {
		return nil, err
	}

	m.audit(ctx, "policy", newPol.ID, "ENDORSED", cmd.Actor, fmt.Sprintf("diff=%d %s", diff, newPol.Currency))
	m.outbox(ctx, "policy", newPol.ID, events.PolicyIssued, map[string]any{"policyId": newPol.ID})
	return newPol, nil
}

type RenewPolicyCmd struct {
	PolicyID string
	Actor    string
	IdemKey  string
}

func (m *Motor) RenewPolicy(ctx context.Context, cmd RenewPolicyCmd) (*store.Policy, error) {
	old, err := m.Repo.GetPolicy(ctx, cmd.PolicyID)
	if err != nil {
		return nil, err
	}
	if old.Status != "ISSUED" {
		return nil, pcerr.E(pcerr.CodeConflict, "can only renew ISSUED policy")
	}

	// Calculate new effective dates (1 year from old end)
	oldTo, err := time.Parse("2006-01-02", old.EffectiveTo)
	if err != nil {
		return nil, err
	}
	newTo := oldTo.AddDate(1, 0, 0)

	rated := rating.CalculateMotor(rating.Input{
		CoverType: old.Risk.CoverType, Usage: old.Risk.Usage, Year: old.Risk.Year,
		SumInsuredMinor: old.Risk.SumInsuredMinor, Locale: i18n.EN,
	})


	newPol := &store.Policy{
		ID:             uuid.NewString(),
		TenantID:       tenant.EIC,
		PolicyNumber:   old.PolicyNumber, // keep same policy number across renewals
		QuoteID:        old.QuoteID,
		PartyID:        old.PartyID,
		ProductCode:    old.ProductCode,
		Status:         "PENDING_PAYMENT",
		Risk:           old.Risk,
		Lines:          rated.Lines,
		TotalMinor:     rated.TotalMinor,
		Currency:       rated.Currency,
		EffectiveFrom:  oldTo.Format("2006-01-02"),
		EffectiveTo:    newTo.Format("2006-01-02"),
		ParentPolicyID: old.ID,
		Version:        old.Version + 1,
	}

	inv := &store.Invoice{
		ID:          uuid.NewString(),
		TenantID:    tenant.EIC,
		PolicyID:    newPol.ID,
		AmountMinor: newPol.TotalMinor,
		Currency:    newPol.Currency,
		Status:      "OPEN",
	}
	newPol.InvoiceID = inv.ID

	if err := m.Repo.SaveInvoice(ctx, inv); err != nil {
		return nil, err
	}
	if err := m.Repo.SavePolicy(ctx, newPol); err != nil {
		return nil, err
	}

	m.audit(ctx, "policy", newPol.ID, "RENEWAL_GENERATED", cmd.Actor, "pending payment")
	return newPol, nil
}

type CancelPolicyCmd struct {
	PolicyID string
	Actor    string
	IdemKey  string
	Reason   string
}

func (m *Motor) CancelPolicy(ctx context.Context, cmd CancelPolicyCmd) error {
	pol, err := m.Repo.GetPolicy(ctx, cmd.PolicyID)
	if err != nil {
		return err
	}
	if pol.Status != "ISSUED" {
		return pcerr.E(pcerr.CodeConflict, "can only cancel ISSUED policy")
	}

	pol.Status = "CANCELLED"
	now := time.Now().UTC()
	pol.EffectiveTo = now.Format("2006-01-02")
	if err := m.Repo.SavePolicy(ctx, pol); err != nil {
		return err
	}

	// Emit cancellation refund logic here (simplistic for demo: flat refund)
	m.audit(ctx, "policy", pol.ID, "CANCELLED", cmd.Actor, cmd.Reason)
	m.outbox(ctx, "policy", pol.ID, events.PolicyCancelled, map[string]string{"policyId": pol.ID, "reason": cmd.Reason})
	return nil
}

func (m *Motor) RunEndOfDayReconciliation(ctx context.Context, date string) (map[string]interface{}, error) {
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	receipts, err := m.Repo.ListDailyReceipts(ctx, date)
	if err != nil {
		return nil, err
	}

	var total int64
	var count int
	for _, r := range receipts {
		if r.Status == "COMPLETED" {
			count++
			// Look up invoice to get amount
			inv, err := m.Repo.GetInvoice(ctx, r.InvoiceID)
			if err == nil && inv != nil {
				total += inv.AmountMinor
			}
		}
	}

	res := map[string]interface{}{
		"date":                 date,
		"totalReceipts":        count,
		"totalAmountMinor":     total,
		"status":               "RECONCILED",
		"erpJournalRef":        uuid.NewString(),
		"reconciliationTime": time.Now().UTC().Format(time.RFC3339),
	}
	
	// Simulate ERP export event
	m.outbox(ctx, "system", date, "ERP_RECONCILIATION_COMPLETED", map[string]string{
		"date": date,
		"totalMinor": fmt.Sprintf("%d", total),
	})

	m.audit(ctx, "system", date, "ERP_RECONCILIATION", "system", fmt.Sprintf("Reconciled %d receipts for total %d minor", count, total))

	return res, nil
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
		Fayda:   integration.MockFayda{},
	}
}
