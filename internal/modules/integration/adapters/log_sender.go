// Package adapters holds integration's provider adapters. For the prototype the
// SMS/email senders log the message; real gateway (Ethio Telecom SMS, SMTP) and
// Telebirr clients drop in here behind the same ports without touching callers.
package adapters

import (
	"context"
	"log/slog"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/integration/ports"
)

// LogSender is a development SmsSender/EmailSender that records outbound messages
// to the log instead of calling an external provider.
type LogSender struct{ logger *slog.Logger }

// NewLogSender builds the logging sender.
func NewLogSender(logger *slog.Logger) *LogSender { return &LogSender{logger: logger} }

var (
	_ ports.SmsSender   = (*LogSender)(nil)
	_ ports.EmailSender = (*LogSender)(nil)
)

// SendSMS logs the SMS.
func (s *LogSender) SendSMS(_ context.Context, to, body string) error {
	s.logger.Info("SMS dispatched", "to", to, "body", body)
	return nil
}

// SendEmail logs the email.
func (s *LogSender) SendEmail(_ context.Context, to, subject, body string) error {
	s.logger.Info("email dispatched", "to", to, "subject", subject)
	return nil
}
