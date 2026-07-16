package product

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// LifecycleAction represents the specific trigger for the lifecycle event.
type LifecycleAction string

const (
	ActionCreated   LifecycleAction = "CREATED"
	ActionSubmitted LifecycleAction = "SUBMITTED"
	ActionApproved  LifecycleAction = "APPROVED"
	ActionActivated LifecycleAction = "ACTIVATED"
	ActionSuspended LifecycleAction = "SUSPENDED"
	ActionRetired   LifecycleAction = "RETIRED"
)

// ProductLifecycleEvent defines the payload published to Kafka (via Outbox)
// matching the Avro schema defined in the SSD.
type ProductLifecycleEvent struct {
	EventID       uuid.UUID       `json:"event_id"`
	TenantID      string          `json:"tenant_id"`
	ProductID     string          `json:"product_id"`
	LOB           string          `json:"lob"`
	Action        LifecycleAction `json:"action"`
	Status        string          `json:"status"`
	Version       int             `json:"version"`
	EffectiveFrom int64           `json:"effective_from"`
	EffectiveTo   *int64          `json:"effective_to,omitempty"`
	OccurredAt    int64           `json:"occurred_at"`
}

// NewProductLifecycleEvent constructs the event from the aggregate root.
func NewProductLifecycleEvent(p *Product, action LifecycleAction) *ProductLifecycleEvent {
	var effectiveTo *int64
	if p.EffectiveTo != nil {
		ts := p.EffectiveTo.UnixMilli()
		effectiveTo = &ts
	}

	return &ProductLifecycleEvent{
		EventID:       uuid.New(),
		TenantID:      p.TenantID,
		ProductID:     p.ID.String(),
		LOB:           p.LOB,
		Action:        action,
		Status:        string(p.Status),
		Version:       p.Version,
		EffectiveFrom: p.EffectiveFrom.UnixMilli(),
		EffectiveTo:   effectiveTo,
		OccurredAt:    time.Now().UTC().UnixMilli(),
	}
}

// Serialize converts the event to JSON for the Outbox. 
// Note: In a true production environment with Apicurio, we would serialize this using `goavro` to binary.
func (e *ProductLifecycleEvent) Serialize() ([]byte, error) {
	return json.Marshal(e)
}
