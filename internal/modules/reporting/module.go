// Package reporting is the composition seam for the reporting context. It builds a
// KPI read model by projecting policy.issued and claims.settled events.
package reporting

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/reporting/adapters"
	reportapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/reporting/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/reporting/rest"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/eventbus"
)

const (
	topicPolicyIssued = "policy.issued"
	topicClaimSettled = "claims.settled"
)

// Module wires the reporting context into the monolith.
type Module struct {
	svc    *reportapp.Service
	logger *slog.Logger
}

// New builds the module from the shared database.
func New(db *database.DB) *Module {
	return &Module{svc: reportapp.NewService(adapters.NewKPIRepository(db))}
}

// Name identifies the module.
func (m *Module) Name() string { return "reporting" }

// Init subscribes to the events that feed the KPI projection.
func (m *Module) Init(k *app.Kernel) error {
	m.logger = k.Logger
	k.Events.Subscribe(topicPolicyIssued, m.onPolicyIssued)
	k.Events.Subscribe(topicClaimSettled, m.onClaimSettled)
	return nil
}

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "reporting", rest.New(m.svc).Routes()
}

func payloadOf(e eventbus.Event) []byte {
	if p, ok := e.(interface{ Payload() []byte }); ok {
		return p.Payload()
	}
	return []byte("{}")
}

func (m *Module) onPolicyIssued(ctx context.Context, e eventbus.Event) error {
	var p struct {
		TenantID   string `json:"tenant_id"`
		GrossMinor int64  `json:"gross_minor"`
	}
	if err := json.Unmarshal(payloadOf(e), &p); err != nil {
		return err
	}
	return m.svc.RecordPolicy(ctx, p.TenantID, p.GrossMinor)
}

func (m *Module) onClaimSettled(ctx context.Context, e eventbus.Event) error {
	var p struct {
		TenantID    string `json:"tenant_id"`
		AmountMinor int64  `json:"amount_minor"`
	}
	if err := json.Unmarshal(payloadOf(e), &p); err != nil {
		return err
	}
	return m.svc.RecordClaim(ctx, p.TenantID, p.AmountMinor)
}
