package dpduk

import (
	"encoding/json"
	"testing"
)

func TestBaseURL(t *testing.T) {
	if got := BaseURL("sandbox", ""); got != HostSandbox {
		t.Fatalf("sandbox: %s", got)
	}
	if got := BaseURL("production", ""); got != HostProduction {
		t.Fatalf("production: %s", got)
	}
	if got := BaseURL("production", " https://custom.example/ "); got != "https://custom.example" {
		t.Fatalf("override: %s", got)
	}
}

func TestParseTokenJSON(t *testing.T) {
	pair, err := parseTokenJSON([]byte(`{"accessToken":"a","refreshToken":"r","expiry":1700000000}`))
	if err != nil || pair.AccessToken != "a" {
		t.Fatalf("%+v %v", pair, err)
	}
	pair, err = parseTokenJSON([]byte(`{"data":{"accessToken":"b"}}`))
	if err != nil || pair.AccessToken != "b" {
		t.Fatalf("%+v %v", pair, err)
	}
}

func TestDecodeDataAuthError(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"error": []map[string]any{{
			"code": 102, "status": 401, "errorCode": "auth/invalid-token", "message": "bad",
		}},
	})
	err := decodeData(raw, &struct{}{})
	ae, ok := err.(*APIError)
	if !ok || !ae.Auth {
		t.Fatalf("%v", err)
	}
}
