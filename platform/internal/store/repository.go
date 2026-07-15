package store

import (
	"context"
	"time"
)

// Repository is the persistence port for all Phase 0 BC aggregates.
// Postgres and in-memory adapters implement this interface.
type Repository interface {
	// Party
	SaveParty(ctx context.Context, p *Party) error
	GetParty(ctx context.Context, id string) (*Party, error)

	// Quote / Policy
	SaveQuote(ctx context.Context, q *Quote) error
	GetQuote(ctx context.Context, id string) (*Quote, error)
	UpdateQuoteStatus(ctx context.Context, id, status string) error
	SavePolicy(ctx context.Context, p *Policy) error
	GetPolicy(ctx context.Context, id string) (*Policy, error)
	NextPolicyNumber(ctx context.Context, year int) (string, error)
	ListIssuedPolicies(ctx context.Context) ([]Policy, error)

	// Billing
	SaveInvoice(ctx context.Context, inv *Invoice) error
	GetInvoice(ctx context.Context, id string) (*Invoice, error)
	SaveReceipt(ctx context.Context, r *Receipt) error

	// Documents
	SaveDocument(ctx context.Context, d *Document) error
	ListDocumentsByPolicy(ctx context.Context, policyID string) ([]Document, error)

	// Claims
	SaveClaim(ctx context.Context, c *Claim) error
	GetClaim(ctx context.Context, id string) (*Claim, error)
	NextClaimNumber(ctx context.Context, year int) (string, error)
	ListClaims(ctx context.Context) ([]Claim, error)

	// Audit
	AppendAudit(ctx context.Context, e AuditEntry) error
	QueryAudit(ctx context.Context, entityType, entityID string, limit int) ([]AuditEntry, error)

	// Outbox (transactional — PG only; memory no-ops)
	InsertOutbox(ctx context.Context, aggregateType, aggregateID, eventType string, payload any) error

	// Lifecycle
	EnsureSchema(ctx context.Context) error
	Close()
}

// MemoryRepository wraps the in-process Memory store.
type MemoryRepository struct {
	*Memory
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{Memory: NewMemory()}
}

func (m *MemoryRepository) SaveParty(_ context.Context, p *Party) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Parties[p.ID] = p
	return nil
}

func (m *MemoryRepository) GetParty(_ context.Context, id string) (*Party, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	p, ok := m.Parties[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (m *MemoryRepository) SaveQuote(_ context.Context, q *Quote) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Quotes[q.ID] = q
	return nil
}

func (m *MemoryRepository) GetQuote(_ context.Context, id string) (*Quote, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	q, ok := m.Quotes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return q, nil
}

func (m *MemoryRepository) UpdateQuoteStatus(_ context.Context, id, status string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	q, ok := m.Quotes[id]
	if !ok {
		return ErrNotFound
	}
	q.Status = status
	return nil
}

func (m *MemoryRepository) SavePolicy(_ context.Context, p *Policy) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Policies[p.ID] = p
	return nil
}

func (m *MemoryRepository) GetPolicy(_ context.Context, id string) (*Policy, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	p, ok := m.Policies[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (m *MemoryRepository) NextPolicyNumber(_ context.Context, year int) (string, error) {
	return m.Memory.NextPolicyNumber(year), nil
}

func (m *MemoryRepository) ListIssuedPolicies(_ context.Context) ([]Policy, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	var out []Policy
	for _, p := range m.Policies {
		if p.Status == "ISSUED" {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (m *MemoryRepository) SaveInvoice(_ context.Context, inv *Invoice) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Invoices[inv.ID] = inv
	return nil
}

func (m *MemoryRepository) GetInvoice(_ context.Context, id string) (*Invoice, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	inv, ok := m.Invoices[id]
	if !ok {
		return nil, ErrNotFound
	}
	return inv, nil
}

func (m *MemoryRepository) SaveReceipt(_ context.Context, r *Receipt) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Receipts[r.ID] = r
	return nil
}

func (m *MemoryRepository) SaveDocument(_ context.Context, d *Document) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Documents[d.ID] = d
	return nil
}

func (m *MemoryRepository) ListDocumentsByPolicy(_ context.Context, policyID string) ([]Document, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	var out []Document
	for _, d := range m.Documents {
		if d.PolicyID == policyID {
			out = append(out, *d)
		}
	}
	return out, nil
}

func (m *MemoryRepository) SaveClaim(_ context.Context, c *Claim) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Claims[c.ID] = c
	return nil
}

func (m *MemoryRepository) GetClaim(_ context.Context, id string) (*Claim, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	c, ok := m.Claims[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (m *MemoryRepository) NextClaimNumber(_ context.Context, year int) (string, error) {
	return m.Memory.NextClaimNumber(year), nil
}

func (m *MemoryRepository) ListClaims(_ context.Context) ([]Claim, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	var out []Claim
	for _, c := range m.Claims {
		out = append(out, *c)
	}
	return out, nil
}

func (m *MemoryRepository) AppendAudit(_ context.Context, e AuditEntry) error {
	m.Memory.AppendAudit(e)
	return nil
}

func (m *MemoryRepository) QueryAudit(_ context.Context, entityType, entityID string, limit int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	var out []AuditEntry
	for i := len(m.Audit) - 1; i >= 0 && len(out) < limit; i-- {
		e := m.Audit[i]
		if entityType != "" && e.EntityType != entityType {
			continue
		}
		if entityID != "" && e.EntityID != entityID {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func (m *MemoryRepository) InsertOutbox(_ context.Context, _, _, _ string, _ any) error { return nil }
func (m *MemoryRepository) EnsureSchema(_ context.Context) error                       { return nil }
func (m *MemoryRepository) Close()                                                       {}

// ErrNotFound is returned when an aggregate is missing.
var ErrNotFound = errNotFound("not found")

type errNotFound string

func (e errNotFound) Error() string { return string(e) }

// AuditEntry helper for services.
func NewAuditEntry(entityType, entityID, action, actor, detail string) AuditEntry {
	return AuditEntry{
		ID: uuidNew(), TenantID: "eic", EntityType: entityType, EntityID: entityID,
		Action: action, Actor: actor, Detail: detail, At: time.Now().UTC(),
	}
}

func uuidNew() string {
	// avoid importing uuid in store package from memory - use google/uuid
	return newUUID()
}
