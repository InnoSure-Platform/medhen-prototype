package command

import (
	"context"
	"fmt"
	"github.com/google/uuid"

	"pc-notification-svc/internal/domain/notification"
	"pc-notification-svc/internal/infrastructure/template"
	"pc-notification-svc/internal/infrastructure/acl"
)

type DispatchCommand struct {
	TenantID     string
	PartyID      uuid.UUID
	EventName    string
	Payload      map[string]interface{}
	TargetLocale string
}

type DispatchHandler struct {
	notificationRepo notification.NotificationRepository
	templateRepo     notification.TemplateRepository
	preferenceRepo   notification.PreferenceRepository
	templateEngine   *template.Engine
	aclClient        acl.Client
}

func NewDispatchHandler(nr notification.NotificationRepository, tr notification.TemplateRepository, pr notification.PreferenceRepository, te *template.Engine, ac acl.Client) *DispatchHandler {
	return &DispatchHandler{
		notificationRepo: nr,
		templateRepo:     tr,
		preferenceRepo:   pr,
		templateEngine:   te,
		aclClient:        ac,
	}
}

func (h *DispatchHandler) Handle(ctx context.Context, cmd DispatchCommand) error {
	// 1. Fetch preferences
	pref, err := h.preferenceRepo.GetByPartyID(ctx, cmd.PartyID)
	if err != nil {
		return fmt.Errorf("failed to get preference: %w", err)
	}

	// 2. Decide channel (simplistic priority: SMS then Email)
	// In reality, this would be more complex or provided by the event/preference
	selectedChannel := notification.ChannelSMS
	
	// 3. Fetch template
	tpl, err := h.templateRepo.GetActive(ctx, cmd.EventName, string(selectedChannel), cmd.TargetLocale)
	if err != nil {
		// fallback to en-US
		tpl, err = h.templateRepo.GetActive(ctx, cmd.EventName, string(selectedChannel), "en-US")
		if err != nil {
			return fmt.Errorf("failed to find template: %w", err)
		}
	}

	// 4. Check suppression
	targetAddress := extractAddress(selectedChannel, cmd.Payload)
	notif := notification.NewNotification(cmd.TenantID, cmd.PartyID, tpl.Code, selectedChannel, notification.Category(tpl.Category), targetAddress, "")

	if pref.IsOptedOut(string(selectedChannel), tpl.Category) {
		notif.MarkSuppressed("User opted out of " + string(selectedChannel) + " for " + tpl.Category)
		return h.notificationRepo.Save(ctx, notif)
	}

	// 5. Render template
	rendered, err := h.templateEngine.Render(tpl.BodyTemplate, cmd.Payload)
	if err != nil {
		notif.MarkFailed("Template render failed: " + err.Error())
		return h.notificationRepo.Save(ctx, notif)
	}
	notif.RenderedContent = rendered

	// 6. Save PENDING to DB
	if err := h.notificationRepo.Save(ctx, notif); err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	// 7. Dispatch to ACL
	receipt, err := h.aclClient.Dispatch(ctx, notif.Channel, notif.RecipientAddress, notif.RenderedContent)
	if err != nil {
		notif.MarkFailed("ACL dispatch failed: " + err.Error())
		_ = h.notificationRepo.UpdateStatus(ctx, notif)
		return err
	}

	notif.MarkDispatched()
	// Optionally mark delivered immediately if synchronous, but normally we wait for webhook
	_ = notif.MarkDelivered(receipt)
	
	return h.notificationRepo.UpdateStatus(ctx, notif)
}

func extractAddress(channel notification.Channel, payload map[string]interface{}) string {
	// Mock extraction
	if channel == notification.ChannelSMS {
		return "+1234567890"
	}
	return "test@example.com"
}
