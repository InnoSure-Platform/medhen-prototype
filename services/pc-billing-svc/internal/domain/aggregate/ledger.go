package aggregate

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LedgerTransaction struct {
	ID            uuid.UUID
	TenantID      string
	ReferenceID   uuid.UUID
	ReferenceType string
	PostedAt      time.Time
	Entries       []JournalEntry
}

type JournalEntry struct {
	ID                  int64
	LedgerTransactionID uuid.UUID
	AccountCode         string
	DebitAmount         decimal.Decimal
	CreditAmount        decimal.Decimal
}

func NewLedgerTransaction(tenantID string, referenceID uuid.UUID, refType string) *LedgerTransaction {
	return &LedgerTransaction{
		ID:            uuid.New(),
		TenantID:      tenantID,
		ReferenceID:   referenceID,
		ReferenceType: refType,
		PostedAt:      time.Now().UTC(),
		Entries:       make([]JournalEntry, 0),
	}
}

func (lt *LedgerTransaction) AddDebit(accountCode string, amount decimal.Decimal) {
	lt.Entries = append(lt.Entries, JournalEntry{
		LedgerTransactionID: lt.ID,
		AccountCode:         accountCode,
		DebitAmount:         amount,
		CreditAmount:        decimal.Zero,
	})
}

func (lt *LedgerTransaction) AddCredit(accountCode string, amount decimal.Decimal) {
	lt.Entries = append(lt.Entries, JournalEntry{
		LedgerTransactionID: lt.ID,
		AccountCode:         accountCode,
		DebitAmount:         decimal.Zero,
		CreditAmount:        amount,
	})
}

func (lt *LedgerTransaction) Validate() error {
	totalDebits := decimal.Zero
	totalCredits := decimal.Zero

	for _, entry := range lt.Entries {
		totalDebits = totalDebits.Add(entry.DebitAmount)
		totalCredits = totalCredits.Add(entry.CreditAmount)
	}

	if !totalDebits.Equal(totalCredits) {
		return errors.New("ledger transaction unbalanced: debits must equal credits")
	}

	return nil
}
