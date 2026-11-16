package dpdlv

import "testing"

func TestBaseURLForMarkets(t *testing.T) {
	if got := BaseURLFor("production", "lt"); got != "https://esiunta.dpd.lt/api/v1" {
		t.Fatalf("lt: %s", got)
	}
	if got := BaseURLFor("production", "ee"); got != "https://telli.dpd.ee/api/v1" {
		t.Fatalf("ee: %s", got)
	}
	if got := BaseURLFor("sandbox", "lt"); got != "https://sandbox-eserviss.dpd.lv/api/v1" {
		t.Fatalf("sandbox: %s", got)
	}
	if NormalizeMarket("lithuania") != MarketLT {
		t.Fatal("alias")
	}
}
