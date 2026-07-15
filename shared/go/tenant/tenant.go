// Package tenant holds tenancy constants and helpers (ADR-PC-015).
package tenant

const EIC = "eic"

type Context struct {
	TenantID string
	UserID   string
	Roles    []string
	Locale   string
}

func (c Context) IsEIC() bool { return c.TenantID == EIC }

func Require(tenantID string) error {
	if tenantID == "" {
		return ErrMissingTenant
	}
	return nil
}

type tenantError string

func (e tenantError) Error() string { return string(e) }

const ErrMissingTenant = tenantError("tenant_id required")
