// Package document is the composition seam for the document context. It subscribes
// to policy.issued and generates the Certificate of Insurance.
package document

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/InnoSure-Platform/medhen-prototype/internal/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/adapters"
	docapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/rest"
	partyports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/eventbus"
)

const topicPolicyIssued = "policy.issued"

// Module wires the document context into the monolith.
type Module struct {
	svc    *docapp.Service
	party  partyports.Reader
	logger *slog.Logger
}

// New builds the module from the database and the party Reader.
func New(db *database.DB, party partyports.Reader) *Module {
	return &Module{svc: docapp.NewService(db, adapters.NewDocumentRepository(db)), party: party}
}

// Name identifies the module.
func (m *Module) Name() string { return "document" }

// Init subscribes to policy.issued to generate the certificate.
func (m *Module) Init(k *app.Kernel) error {
	m.logger = k.Logger
	k.Events.Subscribe(topicPolicyIssued, m.onPolicyIssued)
	return nil
}

// Mount returns the module's route prefix and handler.
func (m *Module) Mount() (string, http.Handler) {
	return "document", rest.New(m.svc).Routes()
}

func (m *Module) onPolicyIssued(ctx context.Context, e eventbus.Event) error {
	payload := []byte("{}")
	if p, ok := e.(interface{ Payload() []byte }); ok {
		payload = p.Payload()
	}
	var p struct {
		TenantID     string `json:"tenant_id"`
		PolicyID     string `json:"policy_id"`
		PolicyNumber string `json:"policy_number"`
		PartyID      string `json:"party_id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	name := p.PartyID
	if pv, err := m.party.GetParty(ctx, p.TenantID, p.PartyID); err == nil {
		name = pv.FullName
	}

	doc, err := m.svc.GenerateCertificate(ctx, p.TenantID, p.PolicyID, p.PolicyNumber, name)
	if err != nil {
		return err
	}
	m.logger.Info("certificate generated", "document", doc.ID, "number", doc.Number)
	return nil
}
