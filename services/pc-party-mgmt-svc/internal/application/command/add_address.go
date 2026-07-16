package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type AddAddressCommand struct {
	PartyID     uuid.UUID
	Type        string
	IsPrimary   bool
	Region      string
	Zone        string
	Woreda      string
	Kebele      string
	HouseNumber string
}

type AddAddressHandler struct {
	uow UnitOfWork
}

func NewAddAddressHandler(uow UnitOfWork) *AddAddressHandler {
	return &AddAddressHandler{uow: uow}
}

func (h *AddAddressHandler) Handle(ctx context.Context, cmd AddAddressCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repo PartyRepository, outbox OutboxPublisher) error {
		party, err := repo.FindByID(ctx, cmd.PartyID)
		if err != nil {
			return err
		}

		address, err := domain.NewAddress(
			cmd.PartyID, domain.AddressType(cmd.Type), cmd.Region, cmd.Zone, cmd.Woreda, cmd.Kebele, cmd.HouseNumber, cmd.IsPrimary,
		)
		if err != nil {
			return err
		}

		party.Addresses = append(party.Addresses, *address)
		party.Version++ // Increment aggregate root version for optimistic locking

		if err := repo.Save(ctx, party); err != nil {
			return err
		}

		// Typically, adding an address might not generate a global event unless it's a domain-significant change (like billing address change)
		// but we'll assume PartyUpdatedEvent for now.
		return nil
	})
}
