# dpd-sdk

[English](./README.md) | 中文

BefShip 使用的 **DPD / Geopost 各国开放接口** 非官方 Go SDK。

> **与 DPD 无隶属关系。** 需自备账号凭证。见 [DISCLAIMER.md](./DISCLAIMER.md)。

## 包一览

| 包 | 17TRACK | 接口 |
|----|---------|------|
| `dpdlv` | LV / LT / EE | Baltics eServiss（`Market=lv\|lt\|ee`） |
| `dpduk` | UK / DPD Local / IE domestic | [UK developer portal](https://developers.api.dpd.co.uk/) |
| `catalog` | 全部 17TRACK DPD 品牌 key | 仅映射；标明是否已实现发货客户端 |

DE/FR/PL/NL 等国别门户 API 不同，目前只在 `catalog` 登记，未实现发货客户端。

## 安装

```bash
go get github.com/befship/dpd-sdk@latest
```
