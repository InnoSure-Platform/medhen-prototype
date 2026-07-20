package domain_test

import (
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/domain"
)

func validAddr() domain.Address {
	return domain.Address{Region: "Addis Ababa", Zone: "Bole", Woreda: "03", Kebele: "05"}
}

func TestNewIndividual_Valid(t *testing.T) {
	p, err := domain.NewIndividual("eic", "Abebe Bikila", "አበበ ቢቂላ", "+251911000000", "ETH-1", validAddr())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type != domain.Individual || p.Status != domain.StatusActive || p.Version != 1 {
		t.Fatalf("unexpected party: %+v", p)
	}
	if p.ID == "" || p.FullNameAmharic != "አበበ ቢቂላ" {
		t.Fatalf("id/amharic not set: %+v", p)
	}
}

func TestNewIndividual_Validation(t *testing.T) {
	cases := map[string]struct {
		name, phone, nid string
		addr             domain.Address
		want             error
	}{
		"no name":     {"", "+2519", "n1", validAddr(), domain.ErrNameRequired},
		"no phone":    {"A", "", "n1", validAddr(), domain.ErrPhoneRequired},
		"no national": {"A", "+2519", "", validAddr(), domain.ErrNationalIDRequired},
		"bad address": {"A", "+2519", "n1", domain.Address{Region: "AA"}, domain.ErrInvalidAddress},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := domain.NewIndividual("eic", c.name, "", c.phone, c.nid, c.addr); err != c.want {
				t.Fatalf("got %v, want %v", err, c.want)
			}
		})
	}
}

func TestNewOrganization_Validation(t *testing.T) {
	if _, err := domain.NewOrganization("eic", "", "", "+2519", "REG-1", validAddr()); err != domain.ErrNameRequired {
		t.Fatalf("empty legal name: got %v", err)
	}
	if _, err := domain.NewOrganization("eic", "Acme", "", "+2519", "", validAddr()); err != domain.ErrNationalIDRequired {
		t.Fatalf("empty reg no: got %v", err)
	}
	p, err := domain.NewOrganization("eic", "Acme plc", "አክሜ", "+2519", "REG-1", validAddr())
	if err != nil || p.Type != domain.Organization {
		t.Fatalf("valid org failed: %+v %v", p, err)
	}
}

func TestSuspend(t *testing.T) {
	p, _ := domain.NewIndividual("eic", "A", "", "+2519", "n1", validAddr())
	if err := p.Suspend(); err != nil {
		t.Fatalf("suspend active: %v", err)
	}
	if p.Status != domain.StatusSuspended || p.Version != 2 {
		t.Fatalf("unexpected after suspend: %+v", p)
	}
	if err := p.Suspend(); err != domain.ErrInvalidTransition {
		t.Fatalf("re-suspend should fail, got %v", err)
	}
}
