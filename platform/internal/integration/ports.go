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

type FaydaClient interface {
	Verify(nationalID string) (bool, error)
}

type MockFayda struct{}

func (MockFayda) Verify(nationalID string) (bool, error) {
	return nationalID != "", nil
}
