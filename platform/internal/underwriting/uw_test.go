package underwriting_test

import (
	"testing"

	"github.com/InnoSure-Platform/pc-platform/internal/underwriting"
)

func TestSTPAccept(t *testing.T) {
	d := underwriting.EvaluateSTP(2018, 100_000_000, "comprehensive")
	if d.Outcome != "ACCEPT" {
		t.Fatalf("%+v", d)
	}
}

func TestSTPDeclineOld(t *testing.T) {
	d := underwriting.EvaluateSTP(2005, 100_000_000, "comprehensive")
	if d.Outcome != "DECLINE" {
		t.Fatalf("%+v", d)
	}
}
