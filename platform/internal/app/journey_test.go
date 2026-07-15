package app_test

import (
	"context"
	"testing"

	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-platform/internal/usecase"
	"github.com/InnoSure-Platform/pc-shared-go/i18n"
)

func TestBuyClaimJourney(t *testing.T) {
	m := usecase.NewDefault()
	ctx := context.Background()
	party, err := m.RegisterParty(ctx, usecase.RegisterPartyCmd{
		FullName: "Test User", PhoneE164: "+251900000001", Actor: "test", IdemKey: "p1",
	})
	if err != nil {
		t.Fatal(err)
	}
	q, err := m.CreateQuote(ctx, usecase.CreateQuoteCmd{
		PartyID: party.ID, ProductCode: "MOTOR-PRIVATE-COMP", Locale: i18n.EN, Actor: "test", IdemKey: "q1",
		Risk: store.MotorRisk{
			PlateNumber: "AA-1-1", Make: "Toyota", Model: "Yaris", Year: 2020,
			Usage: "private", CoverType: "comprehensive", SumInsuredMinor: 80_000_000,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	bind, err := m.BindQuote(ctx, q.ID, "test", "b1")
	if err != nil {
		t.Fatal(err)
	}
	pay, err := m.PayInvoice(ctx, bind.Invoice.ID, "telebirr", party.PhoneE164, "test", "pay1")
	if err != nil {
		t.Fatal(err)
	}
	if pay.Policy.Status != "ISSUED" {
		t.Fatalf("bad pay %+v", pay)
	}
	cl, err := m.SubmitFNOL(ctx, usecase.FNOLCmd{
		PolicyID: pay.Policy.ID, Description: "scratch", EstimatedAmountMinor: 1_000_000,
		Actor: "test", IdemKey: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}
	cl, err = m.SettleFastTrack(ctx, cl.ID, "test", "s1")
	if err != nil {
		t.Fatal(err)
	}
	if cl.Status != "SETTLED" {
		t.Fatalf("%+v", cl)
	}
	k, err := m.KPIs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if k.PoliciesInForce != 1 || k.ClaimsSettled != 1 {
		t.Fatalf("%+v", k)
	}
}
