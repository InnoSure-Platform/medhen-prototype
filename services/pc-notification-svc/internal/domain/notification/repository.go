package notification

import (
	"context"
	"github.com/google/uuid"
	"pc-notification-svc/internal/domain/preference"
	"pc-notification-svc/internal/domain/template"
)

type NotificationRepository interface {
	Save(ctx context.Context, n *Notification) error
	UpdateStatus(ctx context.Context, n *Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	ScrubPartyData(ctx context.Context, partyID uuid.UUID) error
}

type TemplateRepository interface {
	GetActive(ctx context.Context, code string, channel string, locale string) (*template.NotificationTemplate, error)
}

type PreferenceRepository interface {
	GetByPartyID(ctx context.Context, partyID uuid.UUID) (*preference.RoutingPreference, error)
}
