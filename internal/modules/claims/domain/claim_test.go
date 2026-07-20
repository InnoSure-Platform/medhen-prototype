package domain_test

import (
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

func TestNewClaim(t *testing.T) {
	c, err := domain.NewClaim("eic", "pol-1", "party-1", "collision", 9.03, 38.74, money.FromInt(40000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != domain.StatusFiled || c.Reserve.Minor() != 4000000 || !c.SettledAmount.IsZero() {
		t.Fatalf("unexpected claim: %+v", c)
	}
}

func TestNewClaim_NegativeReserve(t *testing.T) {
	if _, err := domain.NewClaim("eic", "pol-1", "party-1", "x", 0, 0, money.FromInt(-1)); err != domain.ErrNegativeAmount {
		t.Fatalf("got %v, want ErrNegativeAmount", err)
	}
}

func TestSettle_WithinAuthority(t *testing.T) {
	c, _ := domain.NewClaim("eic", "pol-1", "party-1", "x", 0, 0, money.FromInt(40000))
	if err := c.Settle(money.FromInt(30000), money.FromInt(50000)); err != nil {
		t.Fatalf("settle: %v", err)
	}
	if c.Status != domain.StatusSettled || c.SettledAmount.Minor() != 3000000 {
		t.Fatalf("unexpected after settle: %+v", c)
	}
}

func TestSettle_OverAuthority(t *testing.T) {
	c, _ := domain.NewClaim("eic", "pol-1", "party-1", "x", 0, 0, money.FromInt(90000))
	if err := c.Settle(money.FromInt(80000), money.FromInt(50000)); err != domain.ErrAuthorityExceeded {
		t.Fatalf("got %v, want ErrAuthorityExceeded", err)
	}
	if c.Status != domain.StatusFiled {
		t.Fatalf("claim should stay FILED after refer, got %s", c.Status)
	}
}

func TestSettle_Guards(t *testing.T) {
	c, _ := domain.NewClaim("eic", "pol-1", "party-1", "x", 0, 0, money.FromInt(40000))
	if err := c.Settle(money.FromInt(-1), money.FromInt(50000)); err != domain.ErrNegativeAmount {
		t.Fatalf("negative settle: got %v", err)
	}
	_ = c.Settle(money.FromInt(100), money.FromInt(50000))
	if err := c.Settle(money.FromInt(100), money.FromInt(50000)); err != domain.ErrAlreadySettled {
		t.Fatalf("re-settle: got %v", err)
	}
}
