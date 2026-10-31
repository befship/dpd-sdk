#!/usr/bin/env bash
# Fetch DPD eServiss OpenAPI, preprocess, generate Go client into internal/generated/eserviss.
# Official JSON is NEVER committed — only downloaded into .cache/specs.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

OAPI_CODEGEN_VERSION="${OAPI_CODEGEN_VERSION:-v2.5.0}"
export PATH="$(go env GOPATH)/bin:${PATH}"

if ! command -v oapi-codegen >/dev/null 2>&1; then
  echo "installing oapi-codegen@${OAPI_CODEGEN_VERSION} ..."
  go install "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}"
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 required" >&2
  exit 1
fi

MANIFEST="${ROOT}/scripts/manifest.yaml"
PREPROCESS="${ROOT}/scripts/preprocess.py"
SPEC_CACHE="${ROOT}/.cache/specs"
OUT_DIR="${ROOT}/internal/generated/eserviss"
CACHED="${SPEC_CACHE}/openapi.json"
TMPDIR="$(mktemp -d "${TMPDIR:-/tmp}/dpdlv-sdk-gen.XXXXXX")"
trap 'rm -rf "$TMPDIR"' EXIT
UA="${CURL_UA:-Mozilla/5.0 (compatible; BefShip-dpdlv-sdk/1.0)}"

mkdir -p "$SPEC_CACHE" "$OUT_DIR"

read_manifest() {
  local key="$1"
  python3 - "$MANIFEST" "$key" <<'PY'
import sys
try:
    import yaml
except ImportError:
    key = sys.argv[2]
    text = open(sys.argv[1], encoding="utf-8").read().splitlines()
    if key == "spec_urls":
        collecting = False
        for line in text:
            if line.startswith("spec_urls:"):
                collecting = True
                continue
            if collecting:
                if line.startswith("  - "):
                    print(line[4:].strip())
                elif line and not line.startswith(" ") and not line.startswith("#"):
                    break
        raise SystemExit(0)
    for line in text:
        line = line.split("#", 1)[0].rstrip()
        if ":" in line and not line.startswith(" "):
            k, _, v = line.partition(":")
            if k.strip() == key:
                print(v.strip())
                raise SystemExit(0)
    raise SystemExit(f"missing manifest key: {key}")
doc = yaml.safe_load(open(sys.argv[1], encoding="utf-8"))
key = sys.argv[2]
if key == "spec_urls":
    for u in doc.get("spec_urls") or []:
        print(u)
else:
    print(doc[key])
PY
}

SERVER_SANDBOX="$(read_manifest server_sandbox)"
SERVER_PROD="$(read_manifest server_prod)"
SERVER_LT="$(read_manifest server_lt)"
SERVER_EE="$(read_manifest server_ee)"
PKG="$(read_manifest package)"

URLS=()
while IFS= read -r u; do
  [[ -n "$u" ]] && URLS+=("$u")
done < <(read_manifest spec_urls)

fetch_spec() {
  if [[ -n "${SPEC_LOCAL:-}" && -f "${SPEC_LOCAL}" ]]; then
    echo "· local ${SPEC_LOCAL}"
    cp "${SPEC_LOCAL}" "$CACHED"
    return 0
  fi
  if [[ -n "${SPEC_URL:-}" ]]; then
    echo "↓ fetch ${SPEC_URL}"
    curl -fsSL -A "$UA" --connect-timeout 30 --max-time 180 --retry 3 -o "${CACHED}.tmp" "$SPEC_URL"
    mv "${CACHED}.tmp" "$CACHED"
    return 0
  fi
  if [[ -f "$CACHED" && "${FORCE_FETCH:-}" != "1" ]]; then
    echo "· cache hit ${CACHED}"
    return 0
  fi
  for url in "${URLS[@]}"; do
    echo "↓ try ${url}"
    if curl -fsSL -A "$UA" --connect-timeout 20 --max-time 120 --retry 2 -o "${CACHED}.tmp" "$url"; then
      mv "${CACHED}.tmp" "$CACHED"
      return 0
    fi
    rm -f "${CACHED}.tmp"
  done
  cat >&2 <<'MSG'
failed to download DPD OpenAPI.

Download JSON from https://eserviss.dpd.lv/docs?api=json then run:
  SPEC_LOCAL=/path/to/openapi.json make generate

Or set SPEC_URL to a reachable OpenAPI document.
MSG
  return 1
}

fetch_spec

CLEANED="${TMPDIR}/eserviss.json"
python3 "$PREPROCESS" \
  --in "$CACHED" \
  --out "$CLEANED" \
  --server-sandbox "$SERVER_SANDBOX" \
  --server-prod "$SERVER_PROD" \
  --server-lt "$SERVER_LT" \
  --server-ee "$SERVER_EE"

echo "→ generate ${PKG}"
SIZE=$(wc -c <"$CLEANED" | tr -d ' ')
if [[ "$SIZE" -gt 800000 ]]; then
  oapi-codegen -generate types -package "$PKG" -o "${OUT_DIR}/types.gen.go" "$CLEANED"
  oapi-codegen -generate client -package "$PKG" -o "${OUT_DIR}/client.gen.go" "$CLEANED"
else
  rm -f "${OUT_DIR}/types.gen.go"
  oapi-codegen -generate types,client -package "$PKG" -o "${OUT_DIR}/client.gen.go" "$CLEANED"
fi

if command -v gofmt >/dev/null 2>&1; then
  gofmt -w "${OUT_DIR}"/*.go
fi

echo "generate done."
