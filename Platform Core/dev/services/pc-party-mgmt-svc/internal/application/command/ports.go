package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type PartyRepository interface {
	Save(ctx context.Context, party *domain.Party) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Party, error)
	FindByNationalID(ctx context.Context, tenantID, nationalID string) (*domain.Party, error)
}

type OutboxPublisher interface {
	Publish(ctx context.Context, event domain.DomainEvent) error
}

type SearchRepository interface {
	FuzzyMatch(ctx context.Context, tenantID, firstName, lastName string, dob time.Time) ([]string, error)
}

type FaydaClient interface {
	VerifyIdentity(ctx context.Context, nationalID string) (bool, error)
}

type RegisterIndividualCommand struct {
	TenantID              string
	FirstName             string
	LastName              string
	DOB                   time.Time
	Gender                string
	NationalIDType        string
	NationalIDNumber      string
	TIN                   string
	OverrideDuplicateFlag bool
}

type MergePartyCommand struct {
	TenantID      string
	SourcePartyID uuid.UUID
	TargetPartyID uuid.UUID
	MergedBy      string
}
