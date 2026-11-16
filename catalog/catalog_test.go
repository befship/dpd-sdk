package catalog

import "testing"

func TestLookupImplemented(t *testing.T) {
	e, ok := Lookup(KeyUK)
	if !ok || !e.Implemented || e.Package != "dpduk" {
		t.Fatalf("%+v", e)
	}
	e, ok = Lookup(KeyLT)
	if !ok || e.Market != "lt" || e.Package != "dpdlv" {
		t.Fatalf("%+v", e)
	}
	got := Implemented()
	if len(got) < 5 {
		t.Fatalf("implemented=%d", len(got))
	}
}
