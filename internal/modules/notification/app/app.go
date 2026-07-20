// Package app holds the notification use cases: enqueue (from events, in the
// relay tx) and dispatch (a background loop that sends queued messages via the
// integration sender).
package app

import (
	"context"
	"log/slog"

	intports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/integration/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification/domain"
)

// Repository persists notifications.
type Repository interface {
	Save(ctx context.Context, n *domain.Notification) error
	ListQueued(ctx context.Context, limit int) ([]*domain.Notification, error)
}

// Service enqueues and dispatches notifications.
type Service struct {
	repo   Repository
	sms    intports.SmsSender
	logger *slog.Logger
}

// NewService builds the service.
func NewService(repo Repository, sms intports.SmsSender, logger *slog.Logger) *Service {
	return &Service{repo: repo, sms: sms, logger: logger}
}

// EnqueueSMS persists a QUEUED SMS. Called from event handlers inside the relay
// transaction, so the queue write commits atomically with the event.
func (s *Service) EnqueueSMS(ctx context.Context, tenantID, recipient, body string) error {
	if recipient == "" {
		return nil // nothing to send to; skip silently
	}
	return s.repo.Save(ctx, domain.NewSMS(tenantID, recipient, body))
}

// Dispatch sends up to `limit` queued notifications via the provider and marks
// their outcome. Runs outside any request/relay transaction (no network I/O in a
// DB tx). At-least-once: a crash between send and mark may re-send.
func (s *Service) Dispatch(ctx context.Context, limit int) (int, error) {
	queued, err := s.repo.ListQueued(ctx, limit)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, n := range queued {
		if err := s.sms.SendSMS(ctx, n.Recipient, n.Body); err != nil {
			n.MarkFailed()
			s.logger.Warn("notification send failed", "id", n.ID, "err", err)
		} else {
			n.MarkSent()
			sent++
		}
		if err := s.repo.Save(ctx, n); err != nil {
			return sent, err
		}
	}
	return sent, nil
}
