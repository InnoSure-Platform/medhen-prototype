// Package app holds the party use cases (commands/queries) and the repository
// port they depend on. Commands run inside a Unit-of-Work so the aggregate write
// and the outbox event commit atomically.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
)

// ErrNotFound is returned when a party does not exist.
var ErrNotFound = errors.New("party: not found")

// Repository persists and loads parties. Implementations use the ambient
// transaction (database.DB.Conn) so they participate in the caller's UoW.
type Repository interface {
	Save(ctx context.Context, p *domain.Party) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Party, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Party, error)
}

// RegisterInput is the command payload for registering an individual.
type RegisterInput struct {
	TenantID        string         `json:"tenant_id"`
	FullName        string         `json:"full_name"`
	FullNameAmharic string         `json:"full_name_amharic"`
	PhoneE164       string         `json:"phone_e164"`
	NationalID      string         `json:"national_id"`
	Address         domain.Address `json:"address"`
}

// Service implements the party use cases.
type Service struct {
	db   *database.DB
	repo Repository
}

// NewService builds the party service.
func NewService(db *database.DB, repo Repository) *Service {
	return &Service{db: db, repo: repo}
}

// Register creates a party and emits PartyRegistered — atomically. The party row
// and the outbox event either both commit or neither does.
func (s *Service) Register(ctx context.Context, in RegisterInput) (string, error) {
	party, err := domain.NewIndividual(in.TenantID, in.FullName, in.FullNameAmharic, in.PhoneE164, in.NationalID, in.Address)
	if err != nil {
		return "", err
	}

	err = s.db.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Save(ctx, party); err != nil {
			return err
		}
		evt := domain.PartyRegistered{
			PartyID: party.ID, TenantID: party.TenantID, Type: party.Type,
			FullName: party.FullName, OccurredAt: time.Now().UTC(),
		}
		payload, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("party: marshal event: %w", err)
		}
		return outbox.Write(ctx, s.db.Conn(ctx), outbox.Message{
			ID:            ids.New(),
			Topic:         domain.TopicPartyRegistered,
			AggregateType: "party",
			AggregateID:   party.ID,
			Payload:       payload,
		})
	})
	if err != nil {
		return "", err
	}
	return party.ID, nil
}

// Get loads a party by id within a tenant.
func (s *Service) Get(ctx context.Context, tenantID, id string) (*domain.Party, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// List returns a tenant's parties (newest first), paginated.
func (s *Service) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Party, error) {
	return s.repo.List(ctx, tenantID, limit, offset)
}
