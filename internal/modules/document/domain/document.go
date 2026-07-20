// Package domain is the document bounded context: generated artifacts such as the
// Certificate of Insurance (COI). Rendering to PDF is an infra concern (Phase 8);
// here we produce and store the certificate content deterministically.
package domain

import (
	"fmt"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
)

// Type is the kind of document.
type Type string

const (
	TypeCertificate Type = "CERTIFICATE_OF_INSURANCE"
)

// Topic published when a document is generated.
const TopicDocumentGenerated = "document.generated"

// Document is a generated artifact tied to a policy.
type Document struct {
	ID        string
	TenantID  string
	PolicyID  string
	Type      Type
	Number    string
	Content   string
	CreatedAt time.Time
}

// NewCertificate builds a Certificate of Insurance for an issued policy.
func NewCertificate(tenantID, policyID, policyNumber, partyName string) *Document {
	number := "COI-" + policyNumber
	content := fmt.Sprintf(
		"CERTIFICATE OF INSURANCE\nInsurer: Ethiopian Insurance Corporation\n"+
			"Certificate No: %s\nPolicy No: %s\nInsured: %s\nProduct: Motor\n"+
			"This certifies that the above policy is in force.",
		number, policyNumber, partyName)
	return &Document{
		ID: ids.New(), TenantID: tenantID, PolicyID: policyID, Type: TypeCertificate,
		Number: number, Content: content, CreatedAt: time.Now().UTC(),
	}
}

// DocumentGenerated is emitted when a document is created.
type DocumentGenerated struct {
	DocumentID string    `json:"document_id"`
	TenantID   string    `json:"tenant_id"`
	PolicyID   string    `json:"policy_id"`
	Type       Type      `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (DocumentGenerated) EventName() string { return TopicDocumentGenerated }
