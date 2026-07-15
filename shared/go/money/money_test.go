package money

import "testing"

func TestETBRoundTrip(t *testing.T) {
	a := NewETB(1234.56)
	if a.Minor != 123456 {
		t.Fatalf("got %d", a.Minor)
	}
	if a.Format() != "1234.56 ETB" {
		t.Fatalf("format %q", a.Format())
	}
}

func TestVATAndStamp(t *testing.T) {
	net := NewETB(10000)
	vat := VATOn(net)
	if vat.Minor != 150000 {
		t.Fatalf("vat %d", vat.Minor)
	}
	stamp := StampDutyMotor(net)
	if stamp.Minor != 5000 { // 0.5% of 10000 = 50 ETB
		t.Fatalf("stamp %d", stamp.Minor)
	}
}
