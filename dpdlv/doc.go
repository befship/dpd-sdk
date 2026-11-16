// Package dpdlv is an unofficial Go client for DPD Baltics eServiss APIs (LV/LT/EE).
//
// Covers 17TRACK carriers: DPD (LV)=100809, DPD (LT)=100810, DPD (EE)=100808.
// Select production host with Config.Market / Creds.Market (lv|lt|ee).
//
// # Overview
//
//   - Basic auth → POST /auth/tokens (or static Bearer token)
//   - Sandbox / production base URLs (LV sandbox; LV/LT/EE production portals)
//   - Typed OpenAPI client via API / Authed
//   - Re-exported schema types for Ship / Label / Track
//
// Generated OpenAPI clients live in internal/generated and are not imported by
// downstream modules directly; use this package instead.
//
// This project is not affiliated with DPD. See DISCLAIMER.md at the module root.
package dpdlv
