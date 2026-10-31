# dpdlv-sdk

English | [中文](./README.zh-CN.md)

Unofficial Go SDK for [DPD Baltics eServiss APIs](https://eserviss.dpd.lv/api/) (LV/LT/EE), generated from the official OpenAPI spec and wrapped with token auth, retries, and a stable facade.

> **Not affiliated with DPD.** You need your own eServiss credentials. See [DISCLAIMER.md](./DISCLAIMER.md).

## Why this layout

| Path | Purpose |
|------|---------|
| `dpdlv/` | Hand-written facade (Basic→Bearer, env URLs, typed API access, errors) |
| `internal/generated/eserviss/` | `oapi-codegen` output (committed so `go get` works) |
| `scripts/` + `generate.sh` | Fetch official JSON → preprocess → generate |
| `.cache/specs/` | Downloaded OpenAPI (**gitignored**, never commit) |

Official OpenAPI files are **not** stored in git. `make generate` pulls them (or accepts `SPEC_LOCAL` / `SPEC_URL`).

## Install

```bash
go get github.com/befship/dpdlv-sdk@latest
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/befship/dpdlv-sdk/dpdlv"
)

func main() {
	client := dpdlv.New(dpdlv.Config{
		Username:    os.Getenv("DPD_USERNAME"),
		Password:    os.Getenv("DPD_PASSWORD"),
		Environment: dpdlv.EnvSandbox,
	})
	api, auth, err := client.API(context.Background())
	if err != nil {
		panic(err)
	}
	_ = api
	_ = auth
	fmt.Println("ready", dpdlv.BaseURL(client.Environment()))
}
```

## Endpoints

| Environment | Base URL |
|---|---|
| Sandbox | `https://sandbox-eserviss.dpd.lv/api/v1` |
| Production | `https://eserviss.dpd.lv/api/v1` |

OAS JSON: `https://eserviss.dpd.lv/docs?api=json`

## Regenerate

```bash
make generate

# Force re-download
FORCE_FETCH=1 make generate

# Local / alternate URL
SPEC_LOCAL=/path/to/openapi.json make generate
SPEC_URL='https://eserviss.dpd.lv/docs?api=json' make generate
```

## License

Hand-written code: MIT. Generated clients: subject to DPD API terms. See `LICENSE` and `DISCLAIMER.md`.
