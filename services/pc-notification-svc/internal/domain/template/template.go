package template

import "github.com/google/uuid"

type NotificationTemplate struct {
	ID              uuid.UUID
	Code            string
	Channel         string
	Locale          string
	Category        string
	SubjectTemplate string
	BodyTemplate    string
	Version         int
}
