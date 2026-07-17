package command

import (
	"context"
	"log"

	"github.com/medhen/pc-billing-svc/internal/application/port"
	"github.com/medhen/pc-billing-svc/internal/domain/aggregate"
)

type ProcessAutoCollectionsCmd struct {
	TenantID string
}

type ProcessAutoCollectionsHandler struct {
	invoiceRepo port.InvoiceRepository
	// In a real system, we'd need an active mandate repository and a gateway service port
	// mandateRepo port.PaymentMandateRepository
	// gatewayPort port.PaymentGatewayPort
}

// In a fully built out repository, we'd have a method to fetch DUE invoices with ACTIVE mandates
type InvoiceQuery interface {
	GetDueInvoicesWithMandates(ctx context.Context, tenantID string) ([]*aggregate.Invoice, error)
}

func NewProcessAutoCollectionsHandler(invoiceRepo port.InvoiceRepository) *ProcessAutoCollectionsHandler {
	return &ProcessAutoCollectionsHandler{
		invoiceRepo: invoiceRepo,
	}
}

// Handle executes the scheduled cron job to pull funds from tokenized mandates.
func (h *ProcessAutoCollectionsHandler) Handle(ctx context.Context, cmd ProcessAutoCollectionsCmd) error {
	log.Printf("Starting auto-collection batch for tenant %s\n", cmd.TenantID)

	// Pseudocode for the actual business logic:
	// 1. invoices := h.invoiceQuery.GetDueInvoicesWithMandates(ctx, cmd.TenantID)
	// 2. For each invoice:
	//    a. mandate := h.mandateRepo.GetActiveByAccount(invoice.BillingAccountID)
	//    b. req := GatewayChargeRequest{ Token: mandate.ProviderToken, Amount: invoice.RemainingAmount() }
	//    c. resp, err := h.gatewayPort.Charge(ctx, req)
	//    d. if resp.Success {
	//         Dispatch ProcessPaymentCallbackCmd internally to execute the UoW.
	//       } else {
	//         invoice.MarkOverdue() or trigger Dunning
	//       }

	return nil
}
