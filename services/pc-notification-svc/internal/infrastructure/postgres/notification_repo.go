package postgres

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"pc-notification-svc/internal/domain/notification"
	// "github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepo struct {
	// pool *pgxpool.Pool // normally we inject this
}

func NewNotificationRepo() *NotificationRepo {
	return &NotificationRepo{}
}

func (r *NotificationRepo) Save(ctx context.Context, n *notification.Notification) error {
	/*
	query := `INSERT INTO notifications (
		id, tenant_id, party_id, template_code, channel, category, 
		status, recipient_address, rendered_content, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	
	_, err := r.pool.Exec(ctx, query, 
		n.ID, n.TenantID, n.PartyID, n.TemplateCode, n.Channel, n.Category,
		n.Status, n.RecipientAddress, n.RenderedContent, n.CreatedAt, n.UpdatedAt,
	)
	return err
	*/
	return nil
}

func (r *NotificationRepo) UpdateStatus(ctx context.Context, n *notification.Notification) error {
	/*
	query := `UPDATE notifications 
			  SET status = $1, vendor_receipt_id = $2, error_reason = $3, updated_at = $4 
			  WHERE id = $5`
	_, err := r.pool.Exec(ctx, query, n.Status, n.VendorReceiptID, n.ErrorReason, n.UpdatedAt, n.ID)
	return err
	*/
	return nil
}

func (r *NotificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*notification.Notification, error) {
	/*
	query := `SELECT id, tenant_id, party_id, template_code, channel, category, status, 
					 recipient_address, rendered_content, vendor_receipt_id, error_reason, created_at, updated_at 
			  FROM notifications WHERE id = $1`
	var n notification.Notification
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.TenantID, &n.PartyID, &n.TemplateCode, &n.Channel, &n.Category, &n.Status,
		&n.RecipientAddress, &n.RenderedContent, &n.VendorReceiptID, &n.ErrorReason, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &n, nil
	*/
	return nil, fmt.Errorf("not implemented")
}

func (r *NotificationRepo) ScrubPartyData(ctx context.Context, partyID uuid.UUID) error {
	/*
	query := `UPDATE notifications 
			  SET recipient_address = '***', rendered_content = '***', updated_at = NOW() 
			  WHERE party_id = $1`
	_, err := r.pool.Exec(ctx, query, partyID)
	return err
	*/
	return nil
}
