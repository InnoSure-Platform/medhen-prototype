package command

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"pc-notification-svc/internal/domain/notification"
)

type HandleDeliveryReceiptCommand struct {
	NotificationID uuid.UUID
	Status         string // DELIVERED, FAILED
	VendorReceipt  string
	Reason         string
}

type HandleDeliveryReceiptHandler struct {
	repo notification.NotificationRepository
}

func NewHandleDeliveryReceiptHandler(repo notification.NotificationRepository) *HandleDeliveryReceiptHandler {
	return &HandleDeliveryReceiptHandler{repo: repo}
}

func (h *HandleDeliveryReceiptHandler) Handle(ctx context.Context, cmd HandleDeliveryReceiptCommand) error {
	notif, err := h.repo.GetByID(ctx, cmd.NotificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	if cmd.Status == "DELIVERED" {
		if err := notif.MarkDelivered(cmd.VendorReceipt); err != nil {
			return err
		}
	} else if cmd.Status == "FAILED" {
		notif.MarkFailed(cmd.Reason)
	}

	return h.repo.UpdateStatus(ctx, notif)
}
