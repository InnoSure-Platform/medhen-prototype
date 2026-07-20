package domain_test

import (
	"strings"
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/domain"
)

func TestNewCertificate(t *testing.T) {
	d := domain.NewCertificate("eic", "pol-1", "EIC/MOT/2026/000001", "Abebe Bikila")
	if d.Type != domain.TypeCertificate || d.Number != "COI-EIC/MOT/2026/000001" || d.ID == "" {
		t.Fatalf("unexpected document: %+v", d)
	}
	if !strings.Contains(d.Content, "Abebe Bikila") || !strings.Contains(d.Content, "EIC/MOT/2026/000001") {
		t.Fatalf("content missing insured/policy: %q", d.Content)
	}
}
