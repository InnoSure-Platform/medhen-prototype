package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type RegisterPartyHandler struct {
	uow    UnitOfWork
	search SearchRepository
	fayda  FaydaClient
}

// We inject the UnitOfWork now instead of repo and outbox separately
func NewRegisterPartyHandler(uow UnitOfWork, search SearchRepository, fayda FaydaClient) *RegisterPartyHandler {
	return &RegisterPartyHandler{
		uow:    uow,
		search: search,
		fayda:  fayda,
	}
}

func (h *RegisterPartyHandler) HandleIndividual(ctx context.Context, cmd RegisterIndividualCommand) (*domain.Party, error) {
	// Fuzzy match check (simplified, outside transaction since it's an external read)
	candidates, err := h.search.FuzzyMatch(ctx, cmd.TenantID, cmd.FirstName, cmd.LastName, cmd.DOB)
	if err == nil && len(candidates) > 0 && !cmd.OverrideDuplicateFlag {
		return nil, domain.ErrDuplicatePartyDetected
	}

	party, err := domain.NewIndividual(
		cmd.TenantID, cmd.FirstName, cmd.LastName, cmd.DOB, cmd.Gender, cmd.NationalIDType, cmd.NationalIDNumber, cmd.TIN,
	)
	if err != nil {
		return nil, err
	}

	// Verify KYC asynchronously or synchronously based on Fayda
	if cmd.NationalIDNumber != "" {
		verified, _ := h.fayda.VerifyIdentity(ctx, cmd.NationalIDNumber)
		if verified {
			party.KYCStatus = domain.KYCStatusVerified
		} else {
			// Degrades to pending if network failed, or remains pending if Fayda returned false
			party.KYCStatus = domain.KYCStatusPending
		}
	}

	event := domain.PartyCreatedEvent{
		ID:             uuid.New(),
		TenantID:       party.TenantID,
		PartyID:        party.ID,
		Type:           party.Type,
		OccurredAtTime: time.Now(),
	}

	// Transactional outbox pattern using Unit of Work
	err = h.uow.Do(ctx, func(ctx context.Context, repo PartyRepository, outbox OutboxPublisher) error {
		// Exact match check inside the transaction for concurrency safety
		existing, err := repo.FindByNationalID(ctx, cmd.TenantID, cmd.NationalIDNumber)
		if err == nil && existing != nil && !cmd.OverrideDuplicateFlag {
			return domain.ErrDuplicatePartyDetected
		}
		
		if err := repo.Save(ctx, party); err != nil {
			return err
		}
		return outbox.Publish(ctx, event)
	})

	if err != nil {
		return nil, err
	}

	return party, nil
}
