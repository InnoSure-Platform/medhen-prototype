package domain_test

import (
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

func TestNewInvoice(t *testing.T) {
	inv := domain.NewInvoice("eic", "pol-1", "party-1", money.FromInt(2680))
	if inv.Status != domain.InvoiceOpen || !inv.AmountPaid.IsZero() || inv.Outstanding().Minor() != 268000 {
		t.Fatalf("unexpected invoice: %+v", inv)
	}
}

func TestInvoiceApply_PartialThenFull(t *testing.T) {
	inv := domain.NewInvoice("eic", "pol-1", "party-1", money.FromInt(2680))

	if err := inv.Apply(money.FromInt(1000)); err != nil {
		t.Fatalf("partial apply: %v", err)
	}
	if inv.Status != domain.InvoicePartiallyPaid || inv.Outstanding().Minor() != 168000 {
		t.Fatalf("after partial: %+v", inv)
	}
	if err := inv.Apply(money.FromInt(1680)); err != nil {
		t.Fatalf("final apply: %v", err)
	}
	if inv.Status != domain.InvoicePaid || !inv.Outstanding().IsZero() {
		t.Fatalf("after full: %+v", inv)
	}
}

func TestInvoiceApply_Overpayment(t *testing.T) {
	inv := domain.NewInvoice("eic", "pol-1", "party-1", money.FromInt(100))
	_ = inv.Apply(money.FromInt(150))
	if inv.Status != domain.InvoicePaid || !inv.Outstanding().IsZero() {
		t.Fatalf("overpayment should be PAID with zero outstanding: %+v", inv)
	}
}

func TestInvoiceApply_RejectsNonPositive(t *testing.T) {
	inv := domain.NewInvoice("eic", "pol-1", "party-1", money.FromInt(100))
	if err := inv.Apply(money.Zero()); err != domain.ErrNonPositivePayment {
		t.Fatalf("zero: got %v", err)
	}
	if err := inv.Apply(money.FromInt(-5)); err != domain.ErrNonPositivePayment {
		t.Fatalf("negative: got %v", err)
	}
}
