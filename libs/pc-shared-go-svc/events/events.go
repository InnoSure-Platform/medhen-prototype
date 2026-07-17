package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Envelope represents a CloudEvent v1.0 wrapper.
type Envelope struct {
	SpecVersion     string          `json:"specversion"`
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	Subject         string          `json:"subject,omitempty"`
	Time            time.Time       `json:"time"`
	DataContentType string          `json:"datacontenttype"`
	Data            json.RawMessage `json:"data"`
}

// NewEnvelope creates a new CloudEvent envelope for a given payload.
func NewEnvelope(source, eventType string, payload interface{}) (*Envelope, error) {
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Envelope{
		SpecVersion:     "1.0",
		ID:              uuid.New().String(),
		Source:          source,
		Type:            eventType,
		Time:            time.Now().UTC(),
		DataContentType: "application/json",
		Data:            dataBytes,
	}, nil
}
