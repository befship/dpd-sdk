// Package catalog maps 17TRACK DPD carrier keys to dpd-sdk packages / markets.
//
// Source: https://res.17track.net/asset/carrier/info/apicarrier.all.json
// Only carriers that share a known shipping API surface are marked Implemented.
package catalog

// TrackKey is a 17TRACK numeric carrier key.
type TrackKey int

// 17TRACK keys for DPD-branded carriers.
const (
	KeyDE       TrackKey = 100007 // DPD (DE)
	KeyUK       TrackKey = 100010 // DPD (UK)
	KeyBRT      TrackKey = 100026 // BRT Bartolini(DPD)
	KeyRU       TrackKey = 100071 // DPD (RU)
	KeyFR       TrackKey = 100072 // DPD (FR)
	KeyPL       TrackKey = 100111 // DPD (PL)
	KeyIE       TrackKey = 100142 // DPD (IE)
	KeyRO       TrackKey = 100177 // DPD (RO)
	KeySpeedy   TrackKey = 100198 // Speedy (DPD)
	KeyPT       TrackKey = 100204 // DPD (PT)
	KeyDawnWing TrackKey = 100288 // Dawn Wing (DPD Laser)
	KeyBE       TrackKey = 100321 // DPD (BE)
	KeyGR       TrackKey = 100404 // DPD (GR)
	KeyCZ       TrackKey = 100483 // DPD (CZ)
	KeyAT       TrackKey = 100556 // DPD (AT)
	KeyHU       TrackKey = 100584 // DPD (HU)
	KeyLocal    TrackKey = 100709 // DPD Local
	KeyHR       TrackKey = 100807 // DPD (HR)
	KeyEE       TrackKey = 100808 // DPD (EE)
	KeyLV       TrackKey = 100809 // DPD (LV)
	KeyLT       TrackKey = 100810 // DPD (LT)
	KeyNL       TrackKey = 100811 // DPD (NL)
	KeySK       TrackKey = 100812 // DPD (SK)
	KeySI       TrackKey = 100813 // DPD (SI)
	KeyLU       TrackKey = 100814 // DPD (LU)
	KeyCH       TrackKey = 100815 // DPD (CH)
	KeyLenton   TrackKey = 100886 // Lenton Group (DPD)
	KeyCN       TrackKey = 101032 // DPD (CN)
	KeyKZ       TrackKey = 101087 // DPD (KZ)
	KeyUZ       TrackKey = 101088 // DPD (UZ)
	KeyGroup    TrackKey = 101108 // DPD Group
	KeyDpdEx    TrackKey = 191032 // DpdEx
)

// Entry describes how a 17TRACK DPD carrier maps onto this module.
type Entry struct {
	Key         TrackKey
	Name        string
	Package     string // Go import path suffix under github.com/befship/dpd-sdk/
	Market      string // e.g. lv|lt|ee for dpdlv; empty otherwise
	Implemented bool   // shipping API client available in this module
	Notes       string
}

// All is the 17TRACK DPD-branded carrier table for this SDK.
var All = []Entry{
	{KeyLV, "DPD (LV)", "dpdlv", "lv", true, "Baltics eServiss"},
	{KeyLT, "DPD (LT)", "dpdlv", "lt", true, "Baltics eServiss (esiunta.dpd.lt)"},
	{KeyEE, "DPD (EE)", "dpdlv", "ee", true, "Baltics eServiss (telli.dpd.ee)"},
	{KeyUK, "DPD (UK)", "dpduk", "", true, "Geopost UK developer portal"},
	{KeyLocal, "DPD Local", "dpduk", "", true, "Same UK portal; network codes differ"},
	{KeyIE, "DPD (IE)", "dpduk", "", true, "Domestic IE via UK portal create domestic"},

	// Listed on 17TRACK but no shipping client in this module yet (country portals differ).
	{KeyDE, "DPD (DE)", "", "", false, "Separate DE Cloud / myDPD APIs"},
	{KeyFR, "DPD (FR)", "", "", false, "Separate FR portal"},
	{KeyPL, "DPD (PL)", "", "", false, "Separate PL portal"},
	{KeyNL, "DPD (NL)", "", "", false, "Separate NL portal"},
	{KeyBE, "DPD (BE)", "", "", false, ""},
	{KeyAT, "DPD (AT)", "", "", false, ""},
	{KeyCH, "DPD (CH)", "", "", false, ""},
	{KeyCZ, "DPD (CZ)", "", "", false, ""},
	{KeySK, "DPD (SK)", "", "", false, ""},
	{KeyHU, "DPD (HU)", "", "", false, ""},
	{KeyRO, "DPD (RO)", "", "", false, ""},
	{KeyHR, "DPD (HR)", "", "", false, ""},
	{KeySI, "DPD (SI)", "", "", false, ""},
	{KeyLU, "DPD (LU)", "", "", false, ""},
	{KeyPT, "DPD (PT)", "", "", false, ""},
	{KeyGR, "DPD (GR)", "", "", false, ""},
	{KeyRU, "DPD (RU)", "", "", false, ""},
	{KeyKZ, "DPD (KZ)", "", "", false, ""},
	{KeyUZ, "DPD (UZ)", "", "", false, ""},
	{KeyCN, "DPD (CN)", "", "", false, ""},
	{KeyGroup, "DPD Group", "", "", false, "Network brand / tracking aggregate"},
	{KeyBRT, "BRT Bartolini(DPD)", "", "", false, "Italy partner brand"},
	{KeySpeedy, "Speedy (DPD)", "", "", false, "Bulgaria partner brand"},
	{KeyDawnWing, "Dawn Wing (DPD Laser)", "", "", false, "South Africa partner"},
	{KeyLenton, "Lenton Group (DPD)", "", "", false, ""},
	{KeyDpdEx, "DpdEx", "", "", false, "China-oriented brand"},
}

// Lookup returns the catalog entry for a 17TRACK key.
func Lookup(key TrackKey) (Entry, bool) {
	for _, e := range All {
		if e.Key == key {
			return e, true
		}
	}
	return Entry{}, false
}

// Implemented returns only carriers with a shipping client in this module.
func Implemented() []Entry {
	out := make([]Entry, 0, 8)
	for _, e := range All {
		if e.Implemented {
			out = append(out, e)
		}
	}
	return out
}
