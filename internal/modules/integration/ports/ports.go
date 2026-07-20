// Package ports is the published contract of the integration module — the
// anti-corruption layer over external providers (SMS gateway, email, Telebirr).
// Other modules depend on these interfaces, never on concrete providers.
package ports

import "context"

// SmsSender delivers an SMS to an E.164 number.
type SmsSender interface {
	SendSMS(ctx context.Context, to, body string) error
}

// EmailSender delivers an email.
type EmailSender interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}
