package quote

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft    Status = "DRAFT"
	StatusQuoted   Status = "QUOTED"
	StatusReferred Status = "REFERRED"
	StatusAccepted Status = "ACCEPTED"
	StatusBound    Status = "BOUND"
	StatusDeclined Status = "DECLINED"
	StatusExpired  Status = "EXPIRED"
)

var (
	ErrQuoteExpired       = errors.New("quote has expired")
	ErrQuoteNotAccepted   = errors.New("quote must be in ACCEPTED state to bind")
	ErrProductNotActive   = errors.New("selected product version is not active")
)

type Quote struct {
	ID          uuid.UUID
	TenantID    string
	ProductID   uuid.UUID
	PartyID     uuid.UUID
	Status      Status
	RiskPayload []byte
	Premium     float64
	ExpiresAt   time.Time
	Notes       []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewQuote(tenantID string, productID, partyID uuid.UUID, riskPayload []byte) *Quote {
	now := time.Now()
	return &Quote{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ProductID:   productID,
		PartyID:     partyID,
		Status:      StatusDraft,
		RiskPayload: riskPayload,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (q *Quote) Calculate(premium float64, requiresReferral bool, validDays int) {
	q.Premium = premium
	if requiresReferral {
		q.Status = StatusReferred
	} else {
		q.Status = StatusQuoted
	}
	q.ExpiresAt = time.Now().AddDate(0, 0, validDays)
	q.UpdatedAt = time.Now()
}

func (q *Quote) Accept() error {
	if q.Status != StatusQuoted {
		return errors.New("can only accept quotes in QUOTED state")
	}
	if time.Now().After(q.ExpiresAt) {
		q.Status = StatusExpired
		return ErrQuoteExpired
	}
	q.Status = StatusAccepted
	q.UpdatedAt = time.Now()
	return nil
}

func (q *Quote) MarkBound() error {
	if q.Status != StatusAccepted {
		return ErrQuoteNotAccepted
	}
	q.Status = StatusBound
	q.UpdatedAt = time.Now()
	return nil
}

// Clone creates a new draft quote based on this quote's properties.
func (q *Quote) Clone() *Quote {
	return NewQuote(q.TenantID, q.ProductID, q.PartyID, q.RiskPayload)
}

// AddNote adds an underwriting or general note to the quote.
func (q *Quote) AddNote(note string) {
	q.Notes = append(q.Notes, note)
	q.UpdatedAt = time.Now()
}
