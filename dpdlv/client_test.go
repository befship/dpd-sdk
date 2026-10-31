package dpdlv

import (
	"strings"
	"testing"
)

func TestBaseURL(t *testing.T) {
	if got := BaseURL(EnvProduction); got != "https://eserviss.dpd.lv/api/v1" {
		t.Fatal(got)
	}
	if got := BaseURL(EnvSandbox); !strings.Contains(got, "sandbox") {
		t.Fatal(got)
	}
	if got := BaseURL(EnvSandbox, "https://example.test/"); got != "https://example.test" {
		t.Fatal(got)
	}
}

func TestNewDefaults(t *testing.T) {
	c := New(Config{Username: "u", Password: "p"})
	if c.Environment() != EnvSandbox {
		t.Fatal(c.Environment())
	}
}
