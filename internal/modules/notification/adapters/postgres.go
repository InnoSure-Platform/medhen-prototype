// Package adapters holds the notification module's Postgres repository.
package adapters

import (
	"context"
	"fmt"

	notifapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Schema is the DDL for the notification module's table.
const Schema = `
CREATE TABLE IF NOT EXISTS notifications (
    id         TEXT PRIMARY KEY,
    tenant_id  TEXT NOT NULL,
    channel    TEXT NOT NULL,
    recipient  TEXT NOT NULL,
    subject    TEXT NOT NULL DEFAULT '',
    body       TEXT NOT NULL,
    status     TEXT NOT NULL,
    attempts   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL,
    sent_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications (status, created_at);
`

// NotificationRepository implements app.Repository.
type NotificationRepository struct{ db *database.DB }

// NewNotificationRepository builds the repository.
func NewNotificationRepository(db *database.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

var _ notifapp.Repository = (*NotificationRepository)(nil)

// Save upserts a notification using the ambient connection.
func (r *NotificationRepository) Save(ctx context.Context, n *domain.Notification) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO notifications (id, tenant_id, channel, recipient, subject, body, status, attempts, created_at, sent_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, attempts=EXCLUDED.attempts, sent_at=EXCLUDED.sent_at`,
		n.ID, n.TenantID, n.Channel, n.Recipient, n.Subject, n.Body, n.Status, n.Attempts, n.CreatedAt, n.SentAt)
	if err != nil {
		return fmt.Errorf("notification repo: save: %w", err)
	}
	return nil
}

// ListQueued returns queued notifications oldest-first.
func (r *NotificationRepository) ListQueued(ctx context.Context, limit int) ([]*domain.Notification, error) {
	rows, err := r.db.Conn(ctx).Query(ctx,
		`SELECT id, tenant_id, channel, recipient, subject, body, status, attempts, created_at, sent_at
		   FROM notifications WHERE status=$1 ORDER BY created_at LIMIT $2`, domain.StatusQueued, limit)
	if err != nil {
		return nil, fmt.Errorf("notification repo: list queued: %w", err)
	}
	defer rows.Close()

	var out []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.TenantID, &n.Channel, &n.Recipient, &n.Subject,
			&n.Body, &n.Status, &n.Attempts, &n.CreatedAt, &n.SentAt); err != nil {
			return nil, err
		}
		out = append(out, &n)
	}
	return out, rows.Err()
}
