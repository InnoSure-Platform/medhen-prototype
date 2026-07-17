package app_test

import (
	"context"
	"testing"
	"github.com/google/uuid"

	"pc-notification-svc/internal/application/command"
	"pc-notification-svc/internal/domain/notification"
	"pc-notification-svc/internal/domain/preference"
	"pc-notification-svc/internal/domain/template"
	tplengine "pc-notification-svc/internal/infrastructure/template"
)

// Mocks
type mockNotifRepo struct {
	saved *notification.Notification
}
func (m *mockNotifRepo) Save(ctx context.Context, n *notification.Notification) error {
	m.saved = n
	return nil
}
func (m *mockNotifRepo) UpdateStatus(ctx context.Context, n *notification.Notification) error {
	m.saved = n
	return nil
}
func (m *mockNotifRepo) GetByID(ctx context.Context, id uuid.UUID) (*notification.Notification, error) {
	return m.saved, nil
}

type mockTplRepo struct {
	mockTpl *template.NotificationTemplate
}
func (m *mockTplRepo) GetActive(ctx context.Context, code string, channel string, locale string) (*template.NotificationTemplate, error) {
	return m.mockTpl, nil
}

type mockPrefRepo struct {
	mockPref *preference.RoutingPreference
}
func (m *mockPrefRepo) GetByPartyID(ctx context.Context, partyID uuid.UUID) (*preference.RoutingPreference, error) {
	return m.mockPref, nil
}

type mockAcl struct {
	called bool
}
func (m *mockAcl) Dispatch(ctx context.Context, channel notification.Channel, address string, content string) (string, error) {
	m.called = true
	return "receipt-123", nil
}

// BDD Scenario 1: Successful Event-Driven SMS
func TestScenario_NOT_BDD_01_SuccessfulEventDrivenSMS(t *testing.T) {
	partyID := uuid.New()
	
	nRepo := &mockNotifRepo{}
	tRepo := &mockTplRepo{
		mockTpl: &template.NotificationTemplate{
			Code: "PaymentReceived",
			Category: "TRANSACTIONAL",
			BodyTemplate: "Payment {{.amount}} received.",
		},
	}
	pRepo := &mockPrefRepo{
		mockPref: &preference.RoutingPreference{
			PartyID: partyID,
			OptedOutSMS: false, // No opt-out
		},
	}
	acl := &mockAcl{}
	engine := tplengine.NewEngine()

	handler := command.NewDispatchHandler(nRepo, tRepo, pRepo, engine, acl)

	cmd := command.DispatchCommand{
		TenantID: "t-1",
		PartyID: partyID,
		EventName: "PaymentReceived",
		Payload: map[string]interface{}{"amount": "100.00"},
		TargetLocale: "en-US",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if nRepo.saved.Status != notification.StatusDelivered {
		t.Errorf("Expected status DELIVERED, got %s", nRepo.saved.Status)
	}
	if !acl.called {
		t.Errorf("Expected ACL Dispatch to be called")
	}
	if nRepo.saved.RenderedContent != "Payment 100.00 received." {
		t.Errorf("Template hydration failed: %s", nRepo.saved.RenderedContent)
	}
}

// BDD Scenario 2: Preference Suppression (Marketing Opt-Out)
func TestScenario_NOT_BDD_02_PreferenceSuppression(t *testing.T) {
	partyID := uuid.New()
	
	nRepo := &mockNotifRepo{}
	tRepo := &mockTplRepo{
		mockTpl: &template.NotificationTemplate{
			Code: "PromoOffer",
			Category: "MARKETING",
			BodyTemplate: "Special offer!",
		},
	}
	pRepo := &mockPrefRepo{
		mockPref: &preference.RoutingPreference{
			PartyID: partyID,
			MarketingOptIn: false, // Opted out
		},
	}
	acl := &mockAcl{}
	engine := tplengine.NewEngine()

	handler := command.NewDispatchHandler(nRepo, tRepo, pRepo, engine, acl)

	cmd := command.DispatchCommand{
		TenantID: "t-1",
		PartyID: partyID,
		EventName: "PromoOffer",
		TargetLocale: "en-US",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if nRepo.saved.Status != notification.StatusSuppressed {
		t.Errorf("Expected status SUPPRESSED, got %s", nRepo.saved.Status)
	}
	if acl.called {
		t.Errorf("Expected ACL Dispatch NOT to be called")
	}
}

// BDD Scenario 3: Statutory Override
func TestScenario_NOT_BDD_03_StatutoryOverride(t *testing.T) {
	partyID := uuid.New()
	
	nRepo := &mockNotifRepo{}
	tRepo := &mockTplRepo{
		mockTpl: &template.NotificationTemplate{
			Code: "CancellationNotice",
			Category: "STATUTORY",
			BodyTemplate: "Policy Cancelled.",
		},
	}
	pRepo := &mockPrefRepo{
		mockPref: &preference.RoutingPreference{
			PartyID: partyID,
			OptedOutSMS: true, // Opted out, but should be overridden
		},
	}
	acl := &mockAcl{}
	engine := tplengine.NewEngine()

	handler := command.NewDispatchHandler(nRepo, tRepo, pRepo, engine, acl)

	cmd := command.DispatchCommand{
		TenantID: "t-1",
		PartyID: partyID,
		EventName: "CancellationNotice",
		TargetLocale: "en-US",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if nRepo.saved.Status != notification.StatusDelivered {
		t.Errorf("Expected status DELIVERED, got %s", nRepo.saved.Status)
	}
	if !acl.called {
		t.Errorf("Expected ACL Dispatch to be called despite opt-out")
	}
}
