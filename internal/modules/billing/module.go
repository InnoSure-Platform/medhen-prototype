// Package billing is the composition seam for the billing bounded context. It is
// the event-consumer reference: on Init it subscribes to policy.issued and raises
// the first invoice; it also verifies and applies Telebirr payments.
package billing

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/adapters"
	billingapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/rest"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/eventbus"
)

// topicPolicyIssued must match policy/domain.TopicPolicyIssued (event contract).
const topicPolicyIssued = "policy.issued"

// Module wires the billing context into the monolith.
type Module struct {
	svc           *billingapp.Service
	webhookSecret string
	logger        *slog.Logger
}

// New builds the module from the database and the Telebirr webhook secret.
func New(db *database.DB, webhookSecret string) *Module {
	svc := billingapp.NewService(billingapp.Deps{
		DB:       db,
		Invoices: adapters.NewInvoiceRepository(db),
		Payments: adapters.NewPaymentRepository(db),
	})
	return &Module{svc: svc, webhookSecret: webhookSecret}
}

// Name identifies the module.
func (m *Module) Name() string { return "billing" }

// Init subscribes to policy.issued so an invoice is raised for every issued
// policy (idempotently, so redelivery does not double-bill).
func (m *Module) Init(k *app.Kernel) error {
	m.logger = k.Logger
	k.Events.Subscribe(topicPolicyIssued, m.onPolicyIssued)
	return nil
}

// policyIssuedPayload is the subset of policy.issued this module consumes.
type policyIssuedPayload struct {
	PolicyID   string `json:"policy_id"`
	TenantID   string `json:"tenant_id"`
	PartyID    string `json:"party_id"`
	GrossMinor int64  `json:"gross_minor"`
}

func (m *Module) onPolicyIssued(ctx context.Context, e eventbus.Event) error {
	payloaded, ok := e.(interface{ Payload() []byte })
	if !ok {
		m.logger.Warn("policy.issued event carried no payload")
		return nil
	}
	var p policyIssuedPayload
	if err := json.Unmarshal(payloaded.Payload(), &p); err != nil {
		return err
	}
	inv, err := m.svc.RaiseInvoiceForPolicy(ctx, p.TenantID, p.PolicyID, p.PartyID, p.GrossMinor)
	if err != nil {
		return err
	}
	m.logger.Info("invoice raised", "invoice", inv.ID, "policy", p.PolicyID, "due_minor", inv.AmountDue.Minor())
	return nil
}

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "billing", rest.New(m.svc, m.webhookSecret).Routes()
}
