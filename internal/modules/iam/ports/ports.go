// Package ports is the published contract of the IAM module.
package ports

import "context"

// Reader resolves an application user's roles by IdP subject, for cross-module
// authorization decisions.
type Reader interface {
	RolesForSubject(ctx context.Context, tenantID, subject string) ([]string, error)
}
