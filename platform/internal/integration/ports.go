package integration

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TelebirrClient is the ACL port for Ethiopian mobile money (BC-MDH-18).
type TelebirrClient interface {
	Charge(phone string, amountMinor int64, reference string) (receiptID string, err error)
}

type MockTelebirr struct{}

func (MockTelebirr) Charge(phone string, amountMinor int64, reference string) (string, error) {
	if amountMinor <= 0 {
		return "", fmt.Errorf("invalid amount")
	}
	_ = phone
	_ = reference
	return "TBL-" + uuid.NewString()[:8], nil
}

type SMSClient interface {
	Send(to, body string) error
}

type MockSMS struct {
	Sent []string
}

func (m *MockSMS) Send(to, body string) error {
	m.Sent = append(m.Sent, fmt.Sprintf("%s|%s|%s", time.Now().UTC().Format(time.RFC3339), to, body))
	return nil
}

type FaydaProfile struct {
	Status   string `json:"status"` // ACTIVE, INACTIVE, DECEASED
	FullName string `json:"fullName"`
	DOB      string `json:"dob"`
}

type InnoGuardClient interface {
	ScreenEntity(entityID, entityType string) (ScreeningResult, error)
}

type ScreeningResult struct {
	Status string // CLEARED, FLAG_SANCTION, FLAG_PEP
	Details string
}

type FaydaClient interface {
	Verify(nationalID string) (*FaydaProfile, error)
}

type MockFayda struct{}

func (MockFayda) Verify(nationalID string) (*FaydaProfile, error) {
	if nationalID == "1234567890" {
		return &FaydaProfile{
			FullName: "Abebe Bikila",
			Status:   "ACTIVE",
		}, nil
	}
	if nationalID == "" {
		return nil, fmt.Errorf("invalid national ID")
	}
	return &FaydaProfile{
		FullName: "Verified Citizen",
		Status:   "ACTIVE",
	}, nil
}

type MockInnoGuard struct{}

func (m MockInnoGuard) ScreenEntity(entityID, entityType string) (ScreeningResult, error) {
	return ScreeningResult{Status: "CLEARED", Details: "Mock cleared"}, nil
}
