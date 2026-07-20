// Package notification is the composition seam for the notification context. It
// subscribes to domain events, resolves the recipient via party.Reader, queues an
// SMS in the relay tx, and dispatches queued messages in a background loop via
// the integration sender.
package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	intports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/integration/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification/adapters"
	notifapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification/app"
	partyports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/eventbus"
)

// Event topics consumed (must match the producing modules' contracts).
const (
	topicPolicyIssued  = "policy.issued"
	topicClaimSettled  = "claims.settled"
	dispatchInterval   = 2 * time.Second
	dispatchBatchLimit = 50
)

// Module wires the notification context into the monolith.
type Module struct {
	svc    *notifapp.Service
	party  partyports.Reader
	logger *slog.Logger
}

// New builds the module from the database, the party Reader and the SMS sender.
func New(db *database.DB, party partyports.Reader, sms intports.SmsSender, logger *slog.Logger) *Module {
	return &Module{
		svc:    notifapp.NewService(adapters.NewNotificationRepository(db), sms, logger),
		party:  party,
		logger: logger,
	}
}

// Name identifies the module.
func (m *Module) Name() string { return "notification" }

// Init subscribes to the events that trigger customer messages.
func (m *Module) Init(k *app.Kernel) error {
	m.logger = k.Logger
	k.Events.Subscribe(topicPolicyIssued, m.onPolicyIssued)
	k.Events.Subscribe(topicClaimSettled, m.onClaimSettled)
	return nil
}

// Mount exposes no query routes for the prototype.
func (m *Module) Mount() (string, http.Handler) { return "", nil }

// RunBackground dispatches queued notifications until the context is cancelled.
func (m *Module) RunBackground(ctx context.Context) {
	ticker := time.NewTicker(dispatchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := m.svc.Dispatch(ctx, dispatchBatchLimit); err != nil {
				m.logger.Warn("notification dispatch error", "err", err)
			}
		}
	}
}

func payloadOf(e eventbus.Event) []byte {
	if p, ok := e.(interface{ Payload() []byte }); ok {
		return p.Payload()
	}
	return []byte("{}")
}

func (m *Module) onPolicyIssued(ctx context.Context, e eventbus.Event) error {
	var p struct {
		TenantID     string `json:"tenant_id"`
		PartyID      string `json:"party_id"`
		PolicyNumber string `json:"policy_number"`
	}
	if err := json.Unmarshal(payloadOf(e), &p); err != nil {
		return err
	}
	phone := m.recipient(ctx, p.TenantID, p.PartyID)
	return m.svc.EnqueueSMS(ctx, p.TenantID, phone,
		fmt.Sprintf("Your Medhen policy %s has been issued.", p.PolicyNumber))
}

func (m *Module) onClaimSettled(ctx context.Context, e eventbus.Event) error {
	var p struct {
		TenantID    string `json:"tenant_id"`
		PartyID     string `json:"party_id"`
		AmountMinor int64  `json:"amount_minor"`
	}
	if err := json.Unmarshal(payloadOf(e), &p); err != nil {
		return err
	}
	phone := m.recipient(ctx, p.TenantID, p.PartyID)
	return m.svc.EnqueueSMS(ctx, p.TenantID, phone,
		fmt.Sprintf("Your claim has been settled for %d.%02d ETB.", p.AmountMinor/100, p.AmountMinor%100))
}

func (m *Module) recipient(ctx context.Context, tenantID, partyID string) string {
	pv, err := m.party.GetParty(ctx, tenantID, partyID)
	if err != nil {
		m.logger.Warn("notification: party lookup failed", "party", partyID, "err", err)
		return ""
	}
	return pv.PhoneE164
}
