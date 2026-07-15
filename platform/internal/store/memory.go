package store

import (
	"sync"
	"time"
)

// Memory is a Phase 0 in-process aggregate store (swap for Postgres adapters per BC).
type Memory struct {
	Mu sync.RWMutex

	Parties     map[string]*Party
	Quotes      map[string]*Quote
	Policies    map[string]*Policy
	Invoices    map[string]*Invoice
	Receipts    map[string]*Receipt
	Documents   map[string]*Document
	Claims      map[string]*Claim
	Audit       []AuditEntry
	Notifications []Notification
	PolicySeq   int
	ClaimSeq    int
}

func NewMemory() *Memory {
	return &Memory{
		Parties:   map[string]*Party{},
		Quotes:    map[string]*Quote{},
		Policies:  map[string]*Policy{},
		Invoices:  map[string]*Invoice{},
		Receipts:  map[string]*Receipt{},
		Documents: map[string]*Document{},
		Claims:    map[string]*Claim{},
	}
}

type Address struct {
	Region string `json:"region,omitempty"`
	Zone   string `json:"zone,omitempty"`
	Woreda string `json:"woreda,omitempty"`
	Kebele string `json:"kebele,omitempty"`
	Line1  string `json:"line1,omitempty"`
	City   string `json:"city,omitempty"`
}

type Party struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenantId"`
	FullName   string    `json:"fullName"`
	FullNameAm string    `json:"fullNameAm,omitempty"`
	PhoneE164  string    `json:"phoneE164"`
	Email      string    `json:"email,omitempty"`
	Status     string    `json:"status"`
	FaydaID    string    `json:"faydaId,omitempty"`
	KYCStatus  string    `json:"kycStatus"`
	Address    *Address  `json:"address,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type MotorRisk struct {
	PlateNumber     string `json:"plateNumber"`
	ChassisNumber   string `json:"chassisNumber,omitempty"`
	Make            string `json:"make"`
	Model           string `json:"model"`
	Year            int    `json:"year"`
	Usage           string `json:"usage"`
	CoverType       string `json:"coverType"`
	SumInsuredMinor int64  `json:"sumInsuredMinor"`
}

type PremiumLine struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	LabelAm     string `json:"labelAm"`
	AmountMinor int64  `json:"amountMinor"`
}

type Quote struct {
	ID          string        `json:"id"`
	TenantID    string        `json:"tenantId"`
	PartyID     string        `json:"partyId"`
	ProductCode string        `json:"productCode"`
	Status      string        `json:"status"`
	Risk        MotorRisk     `json:"risk"`
	Lines       []PremiumLine `json:"lines"`
	TotalMinor  int64         `json:"totalMinor"`
	Currency    string        `json:"currency"`
	UWDecision  string        `json:"uwDecision"`
	ExpiresAt   time.Time     `json:"expiresAt"`
	CreatedAt   time.Time     `json:"createdAt"`
}

type Policy struct {
	ID             string        `json:"id"`
	TenantID       string        `json:"tenantId"`
	PolicyNumber   string        `json:"policyNumber"`
	QuoteID        string        `json:"quoteId"`
	PartyID        string        `json:"partyId"`
	ProductCode    string        `json:"productCode"`
	Status         string        `json:"status"`
	Risk           MotorRisk     `json:"risk"`
	Lines          []PremiumLine `json:"lines"`
	TotalMinor     int64         `json:"totalMinor"`
	Currency       string        `json:"currency"`
	EffectiveFrom  string        `json:"effectiveFrom"`
	EffectiveTo    string        `json:"effectiveTo"`
	IssuedAt       *time.Time    `json:"issuedAt,omitempty"`
	InvoiceID      string        `json:"invoiceId,omitempty"`
	ParentPolicyID string        `json:"parentPolicyId,omitempty"`
	Version        int           `json:"version"`
}

type Invoice struct {
	ID                string `json:"id"`
	TenantID          string `json:"tenantId"`
	PolicyID          string `json:"policyId"`
	AmountMinor       int64  `json:"amountMinor"`
	Currency          string `json:"currency"`
	Status            string `json:"status"`
	DueDate           string `json:"dueDate,omitempty"`
	InstallmentNumber int    `json:"installmentNumber,omitempty"`
}

type Receipt struct {
	ID       string    `json:"id"`
	InvoiceID string   `json:"invoiceId"`
	Channel  string    `json:"channel"`
	Status   string    `json:"status"`
	PaidAt   time.Time `json:"paidAt"`
}

type Document struct {
	ID         string `json:"id"`
	PolicyID   string `json:"policyId"`
	Type       string `json:"type"`
	Locale     string `json:"locale"`
	URL        string `json:"url"`
	ObjectKey  string `json:"objectKey,omitempty"`
	Body       string `json:"-"`
}

type Claim struct {
	ID                   string     `json:"id"`
	ClaimNumber          string     `json:"claimNumber"`
	TenantID             string     `json:"tenantId"`
	PolicyID             string     `json:"policyId"`
	Status               string     `json:"status"`
	Track                string     `json:"track"`
	Description          string     `json:"description"`
	Latitude             float64    `json:"latitude,omitempty"`
	Longitude            float64    `json:"longitude,omitempty"`
	EstimatedAmountMinor int64      `json:"estimatedAmountMinor,omitempty"`
	ReserveMinor         int64      `json:"reserveMinor,omitempty"`
	RecoveryMinor        int64      `json:"recoveryMinor,omitempty"`
	SettlementMinor      int64      `json:"settlementMinor,omitempty"`
	Currency             string     `json:"currency"`
	PhotoObjectKeys      []string   `json:"photoObjectKeys,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	SettledAt            *time.Time `json:"settledAt,omitempty"`
}

type AuditEntry struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenantId"`
	EntityType string    `json:"entityType"`
	EntityID   string    `json:"entityId"`
	Action     string    `json:"action"`
	Actor      string    `json:"actor"`
	Detail     string    `json:"detail,omitempty"`
	At         time.Time `json:"at"`
}

type Notification struct {
	ID      string    `json:"id"`
	Channel string    `json:"channel"`
	To      string    `json:"to"`
	Body    string    `json:"body"`
	At      time.Time `json:"at"`
}

func (m *Memory) AppendAudit(e AuditEntry) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Audit = append(m.Audit, e)
}

func (m *Memory) NextPolicyNumber(year int) string {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.PolicySeq++
	return formatSeq("EIC/MOT", year, m.PolicySeq)
}

func (m *Memory) NextClaimNumber(year int) string {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.ClaimSeq++
	return formatSeq("EIC/CLM", year, m.ClaimSeq)
}

func formatSeq(prefix string, year, seq int) string {
	return prefix + "/" + itoa(year) + "/" + pad6(seq)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func pad6(n int) string {
	s := itoa(n)
	for len(s) < 6 {
		s = "0" + s
	}
	return s
}
