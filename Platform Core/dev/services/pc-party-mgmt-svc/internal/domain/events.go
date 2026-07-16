package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventTypePartyCreated      EventType = "PartyCreated"
	EventTypePartyUpdated      EventType = "PartyUpdated"
	EventTypePartyMerged       EventType = "PartyMerged"
	EventTypeKYCStatusEvaluated EventType = "KYCStatusEvaluated"
)

type DomainEvent interface {
	EventID() uuid.UUID
	EventType() EventType
	OccurredAt() time.Time
}

type PartyCreatedEvent struct {
	ID         uuid.UUID
	TenantID   string
	PartyID    uuid.UUID
	Type       PartyType
	OccurredAtTime time.Time
}

func (e PartyCreatedEvent) EventID() uuid.UUID { return e.ID }
func (e PartyCreatedEvent) EventType() EventType { return EventTypePartyCreated }
func (e PartyCreatedEvent) OccurredAt() time.Time { return e.OccurredAtTime }

type PartyMergedEvent struct {
	ID             uuid.UUID
	TenantID       string
	SourcePartyID  uuid.UUID
	TargetPartyID  uuid.UUID
	MergedBy       string
	OccurredAtTime time.Time
}

func (e PartyMergedEvent) EventID() uuid.UUID { return e.ID }
func (e PartyMergedEvent) EventType() EventType { return EventTypePartyMerged }
func (e PartyMergedEvent) OccurredAt() time.Time { return e.OccurredAtTime }
