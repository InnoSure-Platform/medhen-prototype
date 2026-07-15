package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/InnoSure-Platform/pc-platform/internal/integration"
	"github.com/InnoSure-Platform/pc-platform/internal/product"
	"github.com/InnoSure-Platform/pc-platform/internal/rating"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-platform/internal/underwriting"
	"github.com/InnoSure-Platform/pc-shared-go/calendar"
	pcerr "github.com/InnoSure-Platform/pc-shared-go/errors"
	"github.com/InnoSure-Platform/pc-shared-go/i18n"
	"github.com/InnoSure-Platform/pc-shared-go/idempotency"
	"github.com/InnoSure-Platform/pc-shared-go/tenant"
)

// Platform is the Phase 0 composition root — BC use-cases behind one deployable host.
// Each method maps to a Bounded Context capability; adapters (HTTP) call this surface.
type Platform struct {
	DB      *store.Memory
	Idem    *idempotency.MemoryStore
	Product product.Product
	Pay     integration.TelebirrClient
	SMS     *integration.MockSMS
	DocsDir string
}

func New() *Platform {
	dir := os.Getenv("MEDHEN_DOCS_DIR")
	if dir == "" {
		dir = "./data/docs"
	}
	_ = os.MkdirAll(dir, 0o755)
	return &Platform{
		DB:      store.NewMemory(),
		Idem:    idempotency.NewMemoryStore(),
		Product: product.SeedMotor(),
		Pay:     integration.MockTelebirr{},
		SMS:     &integration.MockSMS{},
		DocsDir: dir,
	}
}

func (p *Platform) audit(entityType, entityID, action, actor, detail string) {
	p.DB.AppendAudit(store.AuditEntry{
		ID: uuid.NewString(), TenantID: tenant.EIC, EntityType: entityType, EntityID: entityID,
		Action: action, Actor: actor, Detail: detail, At: time.Now().UTC(),
	})
}

type RegisterPartyCmd struct {
	FullName   string
	FullNameAm string
	PhoneE164  string
	Email      string
	Address    *store.Address
	Actor      string
	IdemKey    string
}

func (p *Platform) RegisterParty(cmd RegisterPartyCmd) (*store.Party, error) {
	if cmd.FullName == "" || cmd.PhoneE164 == "" {
		return nil, pcerr.E(pcerr.CodeValidation, "fullName and phoneE164 required")
	}
	if cached, replay, err := p.Idem.Begin(nil, "party.register", cmd.IdemKey); err != nil {
		return nil, err
	} else if replay {
		var existing store.Party
		if len(cached) > 0 {
			_ = jsonUnmarshal(cached, &existing)
			return &existing, nil
		}
		return nil, pcerr.E(pcerr.CodeIdempotency, "duplicate request in progress")
	}
	party := &store.Party{
		ID: uuid.NewString(), TenantID: tenant.EIC, FullName: cmd.FullName, FullNameAm: cmd.FullNameAm,
		PhoneE164: cmd.PhoneE164, Email: cmd.Email, Status: "ACTIVE", Address: cmd.Address, CreatedAt: time.Now().UTC(),
	}
	p.DB.Mu.Lock()
	p.DB.Parties[party.ID] = party
	p.DB.Mu.Unlock()
	p.audit("party", party.ID, "REGISTERED", cmd.Actor, party.FullName)
	_ = p.Idem.Complete(nil, "party.register", cmd.IdemKey, party)
	return party, nil
}

func (p *Platform) GetParty(id string) (*store.Party, error) {
	p.DB.Mu.RLock()
	defer p.DB.Mu.RUnlock()
	party, ok := p.DB.Parties[id]
	if !ok {
		return nil, pcerr.E(pcerr.CodeNotFound, "party not found")
	}
	return party, nil
}

type CreateQuoteCmd struct {
	PartyID     string
	ProductCode string
	Risk        store.MotorRisk
	Locale      i18n.Locale
	Actor       string
	IdemKey     string
}

func (p *Platform) CreateQuote(cmd CreateQuoteCmd) (*store.Quote, error) {
	if cmd.ProductCode == "" {
		cmd.ProductCode = p.Product.Code
	}
	if cmd.ProductCode != p.Product.Code {
		return nil, pcerr.E(pcerr.CodeValidation, "unknown product")
	}
	if _, err := p.GetParty(cmd.PartyID); err != nil {
		return nil, err
	}
	if cmd.Risk.Usage == "" {
		cmd.Risk.Usage = "private"
	}
	if cached, replay, _ := p.Idem.Begin(nil, "quote.create", cmd.IdemKey); replay && len(cached) > 0 {
		var q store.Quote
		_ = jsonUnmarshal(cached, &q)
		return &q, nil
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
	p.DB.Mu.Lock()
	p.DB.Quotes[q.ID] = q
	p.DB.Mu.Unlock()
	p.audit("quote", q.ID, "QUOTED", cmd.Actor, fmt.Sprintf("total=%d %s uw=%s", q.TotalMinor, q.Currency, q.UWDecision))
	_ = p.Idem.Complete(nil, "quote.create", cmd.IdemKey, q)
	return q, nil
}

func (p *Platform) GetQuote(id string) (*store.Quote, error) {
	p.DB.Mu.RLock()
	defer p.DB.Mu.RUnlock()
	q, ok := p.DB.Quotes[id]
	if !ok {
		return nil, pcerr.E(pcerr.CodeNotFound, "quote not found")
	}
	return q, nil
}

type BindResult struct {
	Policy  *store.Policy  `json:"policy"`
	Invoice *store.Invoice `json:"invoice"`
}

func (p *Platform) BindQuote(quoteID, actor, idemKey string) (*BindResult, error) {
	if cached, replay, _ := p.Idem.Begin(nil, "quote.bind", idemKey); replay && len(cached) > 0 {
		var br BindResult
		_ = jsonUnmarshal(cached, &br)
		return &br, nil
	}
	q, err := p.GetQuote(quoteID)
	if err != nil {
		return nil, err
	}
	if q.Status != "QUOTED" {
		return nil, pcerr.E(pcerr.CodeConflict, "quote not in QUOTED status")
	}
	now := time.Now().UTC()
	from := now.Format("2006-01-02")
	to := now.AddDate(1, 0, -1).Format("2006-01-02")
	pol := &store.Policy{
		ID: uuid.NewString(), TenantID: tenant.EIC, PolicyNumber: "PENDING", QuoteID: q.ID, PartyID: q.PartyID,
		ProductCode: q.ProductCode, Status: "PENDING_PAYMENT", Risk: q.Risk, Lines: q.Lines,
		TotalMinor: q.TotalMinor, Currency: q.Currency, EffectiveFrom: from, EffectiveTo: to,
	}
	inv := &store.Invoice{
		ID: uuid.NewString(), TenantID: tenant.EIC, PolicyID: pol.ID, AmountMinor: pol.TotalMinor,
		Currency: pol.Currency, Status: "OPEN",
	}
	pol.InvoiceID = inv.ID
	p.DB.Mu.Lock()
	q.Status = "BOUND"
	p.DB.Policies[pol.ID] = pol
	p.DB.Invoices[inv.ID] = inv
	p.DB.Mu.Unlock()
	p.audit("policy", pol.ID, "BOUND", actor, "awaiting payment")
	p.audit("invoice", inv.ID, "CREATED", actor, fmt.Sprintf("%d %s", inv.AmountMinor, inv.Currency))
	br := &BindResult{Policy: pol, Invoice: inv}
	_ = p.Idem.Complete(nil, "quote.bind", idemKey, br)
	return br, nil
}

type PaymentResult struct {
	ReceiptID string            `json:"receiptId"`
	Channel   string            `json:"channel"`
	Status    string            `json:"status"`
	Policy    *store.Policy     `json:"policy"`
	Documents []store.Document  `json:"documents"`
}

func (p *Platform) PayInvoice(invoiceID, channel, phone, actor, idemKey string) (*PaymentResult, error) {
	if channel == "" {
		channel = "telebirr"
	}
	if cached, replay, _ := p.Idem.Begin(nil, "invoice.pay", idemKey); replay && len(cached) > 0 {
		var pr PaymentResult
		_ = jsonUnmarshal(cached, &pr)
		return &pr, nil
	}
	p.DB.Mu.RLock()
	inv, ok := p.DB.Invoices[invoiceID]
	p.DB.Mu.RUnlock()
	if !ok {
		return nil, pcerr.E(pcerr.CodeNotFound, "invoice not found")
	}
	if inv.Status != "OPEN" {
		return nil, pcerr.E(pcerr.CodeConflict, "invoice not open")
	}
	p.DB.Mu.RLock()
	pol := p.DB.Policies[inv.PolicyID]
	party := p.DB.Parties[pol.PartyID]
	p.DB.Mu.RUnlock()
	if phone == "" && party != nil {
		phone = party.PhoneE164
	}
	receiptRef, err := p.Pay.Charge(phone, inv.AmountMinor, inv.ID)
	if err != nil {
		return nil, pcerr.Wrap(pcerr.CodePaymentFailed, "telebirr charge failed", err)
	}
	receipt := &store.Receipt{ID: receiptRef, InvoiceID: inv.ID, Channel: channel, Status: "COMPLETED", PaidAt: time.Now().UTC()}
	year := time.Now().UTC().Year()
	pol.PolicyNumber = p.DB.NextPolicyNumber(year)
	pol.Status = "ISSUED"
	now := time.Now().UTC()
	pol.IssuedAt = &now
	inv.Status = "PAID"

	docs := p.generateDocs(pol, party)

	p.DB.Mu.Lock()
	p.DB.Receipts[receipt.ID] = receipt
	p.DB.Policies[pol.ID] = pol
	p.DB.Invoices[inv.ID] = inv
	for i := range docs {
		d := docs[i]
		p.DB.Documents[d.ID] = &d
	}
	p.DB.Mu.Unlock()

	msg := i18n.T("policy.issued", i18n.EN) + " " + pol.PolicyNumber
	_ = p.SMS.Send(phone, msg)
	p.DB.Mu.Lock()
	p.DB.Notifications = append(p.DB.Notifications, store.Notification{
		ID: uuid.NewString(), Channel: "sms", To: phone, Body: msg, At: time.Now().UTC(),
	})
	p.DB.Mu.Unlock()

	p.audit("payment", receipt.ID, "COMPLETED", actor, channel)
	p.audit("policy", pol.ID, "ISSUED", actor, pol.PolicyNumber)

	outDocs := make([]store.Document, len(docs))
	copy(outDocs, docs)
	pr := &PaymentResult{ReceiptID: receipt.ID, Channel: channel, Status: "COMPLETED", Policy: pol, Documents: outDocs}
	_ = p.Idem.Complete(nil, "invoice.pay", idemKey, pr)
	return pr, nil
}

func (p *Platform) generateDocs(pol *store.Policy, party *store.Party) []store.Document {
	eth := calendar.FromGregorian(time.Now().UTC())
	name := ""
	if party != nil {
		name = party.FullName
		if party.FullNameAm != "" {
			name = party.FullName + " / " + party.FullNameAm
		}
	}
	mk := func(docType, locale, title string) store.Document {
		id := uuid.NewString()
		body := fmt.Sprintf(`ETHIOPIAN INSURANCE CORPORATION
መድህን · Medhen Platform — Phase 0

%s (%s)
Policy: %s
Insured: %s
Vehicle: %s %s %d · Plate %s
Cover: %s · Sum insured: %.2f ETB
Premium: %.2f ETB
Period: %s → %s
Ethiopian date: %s / %s
QR: medhen://policy/%s
`, title, locale, pol.PolicyNumber, name, pol.Risk.Make, pol.Risk.Model, pol.Risk.Year, pol.Risk.PlateNumber,
			pol.Risk.CoverType, float64(pol.Risk.SumInsuredMinor)/100, float64(pol.TotalMinor)/100,
			pol.EffectiveFrom, pol.EffectiveTo, eth.FormatEN(), eth.FormatAM(), pol.ID)
		fname := fmt.Sprintf("%s-%s-%s.txt", pol.PolicyNumber, docType, locale)
		safe := filepath.Join(p.DocsDir, sanitize(fname))
		_ = os.WriteFile(safe, []byte(body), 0o644)
		return store.Document{ID: id, PolicyID: pol.ID, Type: docType, Locale: locale, URL: "/files/" + sanitize(fname), Body: body}
	}
	return []store.Document{
		mk("schedule", "en", i18n.T("doc.schedule", i18n.EN)),
		mk("schedule", "am", i18n.T("doc.schedule", i18n.AM)),
		mk("coi", "en", i18n.T("doc.coi", i18n.EN)),
		mk("coi", "am", i18n.T("doc.coi", i18n.AM)),
		mk("sticker", "en", i18n.T("doc.sticker", i18n.EN)),
	}
}

func sanitize(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			out = append(out, c)
		} else if c == '/' {
			out = append(out, '-')
		}
	}
	return string(out)
}

func (p *Platform) GetPolicy(id string) (*store.Policy, error) {
	p.DB.Mu.RLock()
	defer p.DB.Mu.RUnlock()
	pol, ok := p.DB.Policies[id]
	if !ok {
		return nil, pcerr.E(pcerr.CodeNotFound, "policy not found")
	}
	return pol, nil
}

func (p *Platform) ListPolicyDocuments(policyID string) ([]store.Document, error) {
	if _, err := p.GetPolicy(policyID); err != nil {
		return nil, err
	}
	p.DB.Mu.RLock()
	defer p.DB.Mu.RUnlock()
	var out []store.Document
	for _, d := range p.DB.Documents {
		if d.PolicyID == policyID {
			out = append(out, *d)
		}
	}
	return out, nil
}

type FNOLCmd struct {
	PolicyID             string
	LossDate             time.Time
	Description          string
	Latitude             float64
	Longitude            float64
	EstimatedAmountMinor int64
	PhotoObjectKeys      []string
	Actor                string
	IdemKey              string
}

func (p *Platform) SubmitFNOL(cmd FNOLCmd) (*store.Claim, error) {
	pol, err := p.GetPolicy(cmd.PolicyID)
	if err != nil {
		return nil, err
	}
	if pol.Status != "ISSUED" {
		return nil, pcerr.E(pcerr.CodeValidation, "policy not active/issued")
	}
	if cmd.Description == "" {
		return nil, pcerr.E(pcerr.CodeValidation, "description required")
	}
	if cached, replay, _ := p.Idem.Begin(nil, "claim.fnol", cmd.IdemKey); replay && len(cached) > 0 {
		var c store.Claim
		_ = jsonUnmarshal(cached, &c)
		return &c, nil
	}
	track := "STANDARD"
	if cmd.EstimatedAmountMinor > 0 && cmd.EstimatedAmountMinor <= 5_000_000 {
		track = "FAST_TRACK"
	}
	cl := &store.Claim{
		ID: uuid.NewString(), ClaimNumber: p.DB.NextClaimNumber(time.Now().UTC().Year()),
		TenantID: tenant.EIC, PolicyID: pol.ID, Status: "REGISTERED", Track: track,
		Description: cmd.Description, Latitude: cmd.Latitude, Longitude: cmd.Longitude,
		EstimatedAmountMinor: cmd.EstimatedAmountMinor, Currency: "ETB",
		PhotoObjectKeys: cmd.PhotoObjectKeys, CreatedAt: time.Now().UTC(),
	}
	p.DB.Mu.Lock()
	p.DB.Claims[cl.ID] = cl
	p.DB.Mu.Unlock()
	p.audit("claim", cl.ID, "REGISTERED", cmd.Actor, cl.ClaimNumber+" "+track)
	_ = p.Idem.Complete(nil, "claim.fnol", cmd.IdemKey, cl)
	return cl, nil
}

func (p *Platform) SettleFastTrack(claimID, actor, idemKey string) (*store.Claim, error) {
	if cached, replay, _ := p.Idem.Begin(nil, "claim.settle", idemKey); replay && len(cached) > 0 {
		var c store.Claim
		_ = jsonUnmarshal(cached, &c)
		return &c, nil
	}
	p.DB.Mu.Lock()
	cl, ok := p.DB.Claims[claimID]
	if !ok {
		p.DB.Mu.Unlock()
		return nil, pcerr.E(pcerr.CodeNotFound, "claim not found")
	}
	if cl.Status == "SETTLED" {
		p.DB.Mu.Unlock()
		return cl, nil
	}
	if cl.Track != "FAST_TRACK" {
		p.DB.Mu.Unlock()
		return nil, pcerr.E(pcerr.CodeValidation, "claim not on fast-track")
	}
	cl.Status = "SETTLED"
	cl.SettlementMinor = cl.EstimatedAmountMinor
	now := time.Now().UTC()
	cl.SettledAt = &now
	p.DB.Mu.Unlock()

	p.DB.Mu.RLock()
	pol := p.DB.Policies[cl.PolicyID]
	party := p.DB.Parties[pol.PartyID]
	p.DB.Mu.RUnlock()
	phone := ""
	if party != nil {
		phone = party.PhoneE164
	}
	msg := i18n.T("claim.settled", i18n.EN)
	_ = p.SMS.Send(phone, msg)
	p.audit("claim", cl.ID, "SETTLED", actor, fmt.Sprintf("%d ETB cents", cl.SettlementMinor))
	_ = p.Idem.Complete(nil, "claim.settle", idemKey, cl)
	return cl, nil
}

func (p *Platform) QueryAudit(entityType, entityID string, limit int) []store.AuditEntry {
	if limit <= 0 {
		limit = 50
	}
	p.DB.Mu.RLock()
	defer p.DB.Mu.RUnlock()
	var out []store.AuditEntry
	for i := len(p.DB.Audit) - 1; i >= 0 && len(out) < limit; i-- {
		e := p.DB.Audit[i]
		if entityType != "" && e.EntityType != entityType {
			continue
		}
		if entityID != "" && e.EntityID != entityID {
			continue
		}
		out = append(out, e)
	}
	return out
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

func (p *Platform) KPIs() KPIs {
	p.DB.Mu.RLock()
	defer p.DB.Mu.RUnlock()
	k := KPIs{Currency: "ETB", TenantID: tenant.EIC, ProductCode: p.Product.Code}
	for _, pol := range p.DB.Policies {
		if pol.Status == "ISSUED" {
			k.PoliciesInForce++
			k.GWPMinor += pol.TotalMinor
		}
	}
	for _, c := range p.DB.Claims {
		if c.Status == "SETTLED" {
			k.ClaimsSettled++
		} else {
			k.ClaimsOpen++
		}
	}
	return k
}
