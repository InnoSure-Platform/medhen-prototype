// Package events defines canonical Phase 0 domain event type names.
package events

const (
	PartyRegistered         = "pc.party.registered.v1"
	PolicyQuoted            = "pc.policy.quoted.v1"
	PolicyBound             = "pc.policy.bound.v1"
	PolicyIssued            = "pc.policy.issued.v1"
	PaymentCompleted        = "pc.billing.payment.completed.v1"
	PaymentFailed           = "pc.billing.payment.failed.v1"
	DocumentGenerated       = "pc.document.generated.v1"
	ClaimRegistered         = "pc.claim.registered.v1"
	ClaimSettled            = "pc.claim.settled.v1"
	NotificationRequested   = "pc.notification.requested.v1"
	AuditAppended           = "pc.audit.appended.v1"
)
