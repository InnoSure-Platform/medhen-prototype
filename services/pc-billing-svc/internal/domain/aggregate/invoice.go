package aggregate

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InvoiceStatus string
type InvoiceType string

const (
	InvoiceStatusDraft         InvoiceStatus = "DRAFT"
	InvoiceStatusIssued        InvoiceStatus = "ISSUED"
	InvoiceStatusPartiallyPaid InvoiceStatus = "PARTIALLY_PAID"
	InvoiceStatusPaid          InvoiceStatus = "PAID"
	InvoiceStatusCancelled     InvoiceStatus = "CANCELLED"

	InvoiceTypeNewBusiness InvoiceType = "NEW_BUSINESS"
	InvoiceTypeRenewal     InvoiceType = "RENEWAL"
	InvoiceTypeDebitNote   InvoiceType = "DEBIT_NOTE"
	InvoiceTypeCreditNote  InvoiceType = "CREDIT_NOTE"
)

type Invoice struct {
	ID               uuid.UUID
	BillingAccountID uuid.UUID
	PolicyID         uuid.UUID
	Type             InvoiceType
	TotalAmount      decimal.Decimal
	AmountPaid       decimal.Decimal
	Status               InvoiceStatus
	DueDate              time.Time
	CoverageStartDate    time.Time
	CoverageEndDate      time.Time
	LineItems            []InvoiceLineItem
	CreatedAt            time.Time
}

type InvoiceLineItem struct {
	ID            int64
	InvoiceID     uuid.UUID
	Description   string
	Amount        decimal.Decimal
	TaxAmount     decimal.Decimal
	GLAccountCode string
}

func NewInvoice(accountID, policyID uuid.UUID, invType InvoiceType, dueDate, covStart, covEnd time.Time) *Invoice {
	return &Invoice{
		ID:                uuid.New(),
		BillingAccountID:  accountID,
		PolicyID:          policyID,
		Type:              invType,
		TotalAmount:       decimal.NewFromInt(0),
		AmountPaid:        decimal.NewFromInt(0),
		Status:            InvoiceStatusDraft,
		DueDate:           dueDate,
		CoverageStartDate: covStart,
		CoverageEndDate:   covEnd,
		LineItems:         make([]InvoiceLineItem, 0),
		CreatedAt:         time.Now().UTC(),
	}
}

func (i *Invoice) AddLineItem(desc string, amount, tax decimal.Decimal, glCode string) {
	i.LineItems = append(i.LineItems, InvoiceLineItem{
		InvoiceID:     i.ID,
		Description:   desc,
		Amount:        amount,
		TaxAmount:     tax,
		GLAccountCode: glCode,
	})
	i.TotalAmount = i.TotalAmount.Add(amount).Add(tax)
}

func (i *Invoice) Issue() error {
	if i.Status != InvoiceStatusDraft {
		return errors.New("only draft invoices can be issued")
	}
	if i.TotalAmount.LessThanOrEqual(decimal.Zero) {
		return errors.New("total amount must be greater than zero")
	}
	i.Status = InvoiceStatusIssued
	return nil
}

func (i *Invoice) ApplyPayment(amount decimal.Decimal) (unallocated decimal.Decimal, err error) {
	if i.Status == InvoiceStatusPaid || i.Status == InvoiceStatusCancelled {
		return amount, errors.New("cannot apply payment to paid or cancelled invoice")
	}

	remaining := i.TotalAmount.Sub(i.AmountPaid)
	
	if amount.GreaterThanOrEqual(remaining) {
		i.AmountPaid = i.TotalAmount
		i.Status = InvoiceStatusPaid
		return amount.Sub(remaining), nil // Return unallocated
	}

	i.AmountPaid = i.AmountPaid.Add(amount)
	i.Status = InvoiceStatusPartiallyPaid
	return decimal.Zero, nil
}
