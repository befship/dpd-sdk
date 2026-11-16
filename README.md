# dpd-sdk

English | [中文](./README.zh-CN.md)

Unofficial Go SDK for **DPD / Geopost country APIs** used by BefShip.

> **Not affiliated with DPD.** You need your own account credentials. See [DISCLAIMER.md](./DISCLAIMER.md).

## Packages

| Package | 17TRACK carriers | API |
|---------|------------------|-----|
| [`dpdlv`](./dpdlv) | DPD (LV) `100809`, DPD (LT) `100810`, DPD (EE) `100808` | Baltics eServiss OpenAPI (`Market=lv\|lt\|ee`) |
| [`dpduk`](./dpduk) | DPD (UK) `100010`, DPD Local `100709`, DPD (IE) domestic via UK portal | [developers.api.dpd.co.uk](https://developers.api.dpd.co.uk/) |
| [`catalog`](./catalog) | Full 17TRACK DPD-branded key table | Mapping only — marks which keys are implemented |

Other 17TRACK DPD brands (DE/FR/PL/NL/…) use **different** national portals; they are listed in `catalog` as not implemented until a dedicated client exists.

## Layout

| Path | Purpose |
|------|---------|
| `dpdlv/` | Baltics eServiss facade (Basic→Bearer, typed OpenAPI client) |
| `dpduk/` | UK / DPD Local hand-written REST client (JWT + Client-Id) |
| `catalog/` | 17TRACK key → package / market map |
| `internal/generated/eserviss/` | `oapi-codegen` output for Baltics |
| `scripts/` + `generate.sh` | Fetch Baltics OpenAPI → generate |

## Install

```bash
go get github.com/befship/dpd-sdk@latest
```

## Quick start — Baltics (LV/LT/EE)

```go
client := dpdlv.New(dpdlv.Config{
	Username:    os.Getenv("DPD_USERNAME"),
	Password:    os.Getenv("DPD_PASSWORD"),
	Environment: dpdlv.EnvSandbox,
	Market:      dpdlv.MarketLT, // or MarketLV / MarketEE
})
api, auth, err := client.API(context.Background())
```

## Quick start — UK / DPD Local

```go
c := dpduk.New()
tok, err := c.GetAccessToken(ctx, dpduk.Creds{
	APIKey: os.Getenv("DPD_UK_KEY"), APISecret: os.Getenv("DPD_UK_SECRET"),
	Environment: dpduk.EnvSandbox,
})
_ = tok
```

## Regenerate Baltics client

```bash
make generate
```

## License

Hand-written code: MIT. Generated clients: subject to DPD API terms.
