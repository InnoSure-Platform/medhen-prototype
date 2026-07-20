package domain_test

import (
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/domain"
)

func TestNewUser_Valid(t *testing.T) {
	u, err := domain.NewUser("eic", "kc|abc", "a@eic.et", "Adjuster A", []string{"claims", "staff"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID == "" || !u.HasRole("claims") || u.HasRole("admin") {
		t.Fatalf("unexpected user: %+v", u)
	}
}

func TestNewUser_Validation(t *testing.T) {
	if _, err := domain.NewUser("eic", "", "e", "n", []string{"x"}); err != domain.ErrSubjectRequired {
		t.Fatalf("no subject: got %v", err)
	}
	if _, err := domain.NewUser("eic", "s", "e", "n", nil); err != domain.ErrNoRoles {
		t.Fatalf("no roles: got %v", err)
	}
}
