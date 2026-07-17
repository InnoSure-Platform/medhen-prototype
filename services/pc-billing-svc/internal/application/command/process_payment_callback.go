package command

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/medhen/pc-billing-svc/internal/application/port"
	"github.com/medhen/pc-billing-svc/internal/domain/aggregate"
	"github.com/shopspring/decimal"
)

type ProcessPaymentCallbackCmd struct {
	TenantID             string
	GatewayTransactionID string
	Method               string
	Amount               decimal.Decimal
	InvoiceID            *uuid.UUID // Optional: if provided, allocate directly
}

type ProcessPaymentCallbackHandler struct {
	uow           port.UnitOfWork
	paymentRepo   port.PaymentRepository
	invoiceRepo   port.InvoiceRepository
	accountRepo   port.BillingAccountRepository
	ledgerRepo    port.LedgerRepository
}

func NewProcessPaymentCallbackHandler(
	uow           port.UnitOfWork,
	paymentRepo   port.PaymentRepository,
	invoiceRepo   port.InvoiceRepository,
	accountRepo   port.BillingAccountRepository,
	ledgerRepo    port.LedgerRepository,
) *ProcessPaymentCallbackHandler {
	return &ProcessPaymentCallbackHandler{
		uow:           uow,
		paymentRepo:   paymentRepo,
		invoiceRepo:   invoiceRepo,
		accountRepo:   accountRepo,
		ledgerRepo:    ledgerRepo,
	}
}

func (h *ProcessPaymentCallbackHandler) Handle(ctx context.Context, cmd ProcessPaymentCallbackCmd) error {
	return h.uow.Execute(ctx, func(ctx context.Context) error {
		// 1. Check if payment already exists (Idempotency inside UoW)
		existingPayment, err := h.paymentRepo.GetByGatewayTxID(ctx, cmd.GatewayTransactionID)
		if err == nil && existingPayment != nil {
			// Already processed, return nil to ack webhook
			return nil
		}

		// 2. Create Payment
		payment := aggregate.NewPayment(cmd.TenantID, cmd.Method, cmd.GatewayTransactionID, cmd.Amount)
		payment.MarkSuccess()

		// 3. Double-entry ledger for cash clearing
		ledgerTx := aggregate.NewLedgerTransaction(cmd.TenantID, payment.ID, "PAYMENT")
		ledgerTx.AddDebit("1100", cmd.Amount) // Debit Cash/Bank

		// 4. Allocate if InvoiceID is provided
		if cmd.InvoiceID != nil {
			invoice, err := h.invoiceRepo.GetByID(ctx, *cmd.InvoiceID)
			if err != nil {
				return err
			}

			// Allocate what we can to the invoice
			remainingInvoiceBal := invoice.TotalAmount.Sub(invoice.AmountPaid)
			allocationAmount := cmd.Amount
			if allocationAmount.GreaterThan(remainingInvoiceBal) {
				allocationAmount = remainingInvoiceBal
			}

			if allocationAmount.GreaterThan(decimal.Zero) {
				err = payment.Allocate(invoice.ID, allocationAmount)
				if err != nil {
					return err
				}

				_, err = invoice.ApplyPayment(allocationAmount)
				if err != nil {
					return err
				}
				
				err = h.invoiceRepo.Save(ctx, invoice)
				if err != nil {
					return err
				}

				// Credit A/R for the allocated amount
				ledgerTx.AddCredit("1000", allocationAmount)
			}

			// Handle overpayment
			if payment.UnallocatedAmount.GreaterThan(decimal.Zero) {
				account, err := h.accountRepo.GetByID(ctx, invoice.BillingAccountID)
				if err != nil {
					return err
				}
				account.AddSuspense(payment.UnallocatedAmount)
				
				err = h.accountRepo.Save(ctx, account)
				if err != nil {
					return err
				}

				// Credit Suspense for overpayment
				ledgerTx.AddCredit("2200", payment.UnallocatedAmount)
			}
		} else {
			// No invoice provided, all goes to suspense
			return errors.New("auto-allocation without invoice ID not implemented yet")
		}

		// 5. Validate and save ledger
		err = ledgerTx.Validate()
		if err != nil {
			return err
		}

		err = h.ledgerRepo.Save(ctx, ledgerTx)
		if err != nil {
			return err
		}

		// 6. Save Payment
		err = h.paymentRepo.Save(ctx, payment)
		if err != nil {
			return err
		}

		return nil
	})
}
