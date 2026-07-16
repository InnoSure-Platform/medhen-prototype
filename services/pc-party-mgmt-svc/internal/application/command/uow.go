package command

import "context"

// UnitOfWork defines the transactional boundary for application commands.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, repo PartyRepository, outbox OutboxPublisher) error) error
}
