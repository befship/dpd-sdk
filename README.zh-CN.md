# dpdlv-sdk

[English](./README.md) | 中文

非官方 Go SDK，面向 [DPD Baltics eServiss](https://eserviss.dpd.lv/api/)（LV/LT/EE）。从官方 OpenAPI 生成客户端，并提供 token 鉴权、重试与稳定 facade。

> **与 DPD 无隶属关系。** 需自备 eServiss 账号。见 [DISCLAIMER.md](./DISCLAIMER.md)。

## 布局

| 路径 | 作用 |
|------|------|
| `dpdlv/` | 手写 facade |
| `internal/generated/eserviss/` | `oapi-codegen` 产物（已提交） |
| `scripts/` + `generate.sh` | 拉取官方 JSON → 预处理 → 生成 |
| `.cache/specs/` | 下载的 OpenAPI（**不入库**） |

## 安装

```bash
go get github.com/befship/dpdlv-sdk@latest
```

## 重新生成

```bash
make generate
FORCE_FETCH=1 make generate
SPEC_LOCAL=/path/to/openapi.json make generate
```
