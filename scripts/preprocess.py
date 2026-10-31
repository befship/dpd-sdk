#!/usr/bin/env python3
"""Normalize DPD eServiss OpenAPI before oapi-codegen.

- Rewrites MD5 / dotted operationIds to stable {Method}{PathPascal} names
- Sets sandbox + production (+ LT/EE) servers from manifest values
"""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any

METHOD_VERB = {
    "get": "Get",
    "post": "Create",
    "put": "Update",
    "patch": "Patch",
    "delete": "Delete",
    "head": "Head",
    "options": "Options",
}

PATH_PARAM_RE = re.compile(r"\{([^}]+)\}")
NON_ALNUM_RE = re.compile(r"[^A-Za-z0-9]+")


def path_to_pascal(path: str) -> str:
    parts: list[str] = []
    for seg in path.strip("/").split("/"):
        if not seg:
            continue
        m = PATH_PARAM_RE.fullmatch(seg)
        if m:
            name = m.group(1)
            parts.append("By" + "".join(p.capitalize() for p in NON_ALNUM_RE.split(name) if p))
            continue
        for p in NON_ALNUM_RE.split(seg):
            if not p:
                continue
            parts.append(p[:1].upper() + p[1:])
    return "".join(parts) or "Root"


def op_name(method: str, path: str) -> str:
    verb = METHOD_VERB.get(method.lower(), method.capitalize())
    return verb + path_to_pascal(path)


def needs_rewrite(oid: str | None) -> bool:
    if not oid:
        return True
    if "." in oid:
        return True
    # MD5 hex (32 chars) — with or without leading N from oapi-codegen
    bare = oid[1:] if oid.startswith("N") and len(oid) == 33 else oid
    if re.fullmatch(r"[0-9a-fA-F]{32}", bare):
        return True
    return False


def set_servers(
    doc: dict[str, Any],
    sandbox: str,
    prod: str,
    lt: str | None,
    ee: str | None,
) -> None:
    servers: list[dict[str, str]] = [
        {"url": sandbox, "description": "DPD LV sandbox"},
        {"url": prod, "description": "DPD LV production"},
    ]
    if lt:
        servers.append({"url": lt, "description": "DPD LT Api"})
    if ee:
        servers.append({"url": ee, "description": "DPD EE Api"})
    doc["servers"] = servers


def declared_path_params(op: dict[str, Any], path_item: dict[str, Any]) -> set[str]:
    names: set[str] = set()
    for src in (op.get("parameters") or [], path_item.get("parameters") or []):
        for p in src:
            if isinstance(p, dict) and p.get("in") == "path" and p.get("name"):
                names.add(str(p["name"]))
    return names


def ensure_param_schemas(doc: dict[str, Any]) -> int:
    """Add a default string schema when a parameter has neither schema nor content."""
    fixed = 0
    for item in (doc.get("paths") or {}).values():
        if not isinstance(item, dict):
            continue
        for key, op in item.items():
            if key == "parameters" and isinstance(op, list):
                targets = op
            elif isinstance(op, dict):
                targets = op.get("parameters") or []
            else:
                continue
            for p in targets:
                if not isinstance(p, dict) or "$ref" in p:
                    continue
                if "schema" not in p and "content" not in p:
                    p["schema"] = {"type": "string"}
                    fixed += 1
    return fixed


# OpenAPI formats that oapi-codegen accepts per primitive type.
_VALID_FORMATS: dict[str, set[str]] = {
    "string": {"date", "date-time", "password", "byte", "binary", "email", "uuid", "uri", "hostname", "ipv4", "ipv6"},
    "integer": {"int32", "int64"},
    "number": {"float", "double"},
    "boolean": set(),
}

# Non-standard DPD formats → canonical OpenAPI format (or None to drop).
_FORMAT_REMAP: dict[tuple[str | None, str], str | None] = {
    ("integer", "int"): "int64",
    ("number", "int"): "int64",  # also flip type to integer below
    ("string", "int"): None,
    ("string", "datetime"): "date-time",
    ("string", "guid"): "uuid",
    ("string", "base64"): "byte",
    ("string", "enum"): None,
    ("string", "string"): None,
    ("boolean", "boolean"): None,
    ("boolean", "bool"): None,
    ("integer", "integer"): None,
    ("integer", "string"): None,
}


def simplify_oneof_null(node: Any) -> int:
    """oapi-codegen rejects type:null inside oneOf; collapse to open object like prior patch."""
    fixed = 0
    if isinstance(node, dict):
        one = node.get("oneOf")
        if isinstance(one, list) and any(
            isinstance(x, dict) and x.get("type") == "null" for x in one
        ):
            # Drop oneOf entirely — cells are heterogeneous JSON values.
            for k in ("oneOf", "anyOf", "allOf", "type", "format", "items", "properties"):
                node.pop(k, None)
            fixed += 1
        for v in list(node.values()):
            fixed += simplify_oneof_null(v)
    elif isinstance(node, list):
        for v in node:
            fixed += simplify_oneof_null(v)
    return fixed


def sanitize_types(node: Any) -> int:
    """Map non-standard JSON Schema types (e.g. type: int) to OpenAPI primitives."""
    fixed = 0
    if isinstance(node, dict):
        t = node.get("type")
        if t == "int":
            node["type"] = "integer"
            if "format" not in node:
                node["format"] = "int64"
            fixed += 1
        elif t == "float":
            node["type"] = "number"
            node["format"] = "float"
            fixed += 1
        elif t == "double":
            node["type"] = "number"
            node["format"] = "double"
            fixed += 1
        elif t == "bool":
            node["type"] = "boolean"
            fixed += 1
        for v in node.values():
            fixed += sanitize_types(v)
    elif isinstance(node, list):
        for v in node:
            fixed += sanitize_types(v)
    return fixed


def sanitize_formats(node: Any) -> int:
    """Normalize DPD's non-standard schema formats for oapi-codegen."""
    fixed = 0
    if isinstance(node, dict):
        t = node.get("type")
        f = node.get("format")
        if isinstance(f, str):
            key = (t if isinstance(t, str) else None, f)
            if key in _FORMAT_REMAP:
                new_f = _FORMAT_REMAP[key]
                if key == ("number", "int"):
                    node["type"] = "integer"
                if new_f is None:
                    del node["format"]
                else:
                    node["format"] = new_f
                fixed += 1
            elif isinstance(t, str) and t in _VALID_FORMATS and f not in _VALID_FORMATS[t]:
                del node["format"]
                fixed += 1
        for v in node.values():
            fixed += sanitize_formats(v)
    elif isinstance(node, list):
        for v in node:
            fixed += sanitize_formats(v)
    return fixed


def fix_missing_path_params(doc: dict[str, Any]) -> int:
    """Official DPD OAS omits / misnames some path parameters; oapi-codegen is strict."""
    fixed = 0
    paths = doc.get("paths") or {}
    for path, item in paths.items():
        if not isinstance(item, dict):
            continue
        needed = PATH_PARAM_RE.findall(path)
        if not needed:
            continue
        for method, op in item.items():
            if method.startswith("x-") or not isinstance(op, dict):
                continue
            if method.lower() not in METHOD_VERB and method.lower() not in {
                "get",
                "post",
                "put",
                "patch",
                "delete",
                "head",
                "options",
                "trace",
            }:
                continue
            params = list(op.get("parameters") or [])
            path_params = [p for p in params if isinstance(p, dict) and p.get("in") == "path"]
            have = {str(p.get("name")) for p in path_params if p.get("name")}

            # Single template vs single wrong-named path param (e.g. ids vs {id})
            if len(needed) == 1 and len(path_params) == 1:
                want = needed[0]
                got = str(path_params[0].get("name") or "")
                if got != want:
                    path_params[0]["name"] = want
                    path_params[0]["required"] = True
                    fixed += 1
                    have = {want}

            for name in needed:
                if name in have:
                    continue
                params.append(
                    {
                        "name": name,
                        "in": "path",
                        "required": True,
                        "schema": {"type": "string"},
                    }
                )
                fixed += 1
                have.add(name)

            # Drop extra path params that are not in the path template
            if any(str(p.get("name")) not in needed for p in path_params):
                params = [
                    p
                    for p in params
                    if not (
                        isinstance(p, dict)
                        and p.get("in") == "path"
                        and str(p.get("name")) not in needed
                    )
                ]
                fixed += 1
            op["parameters"] = params
    return fixed


def normalize_operation_ids(doc: dict[str, Any]) -> int:
    paths = doc.get("paths") or {}
    used: set[str] = set()
    rewritten = 0
    for path, item in paths.items():
        if not isinstance(item, dict):
            continue
        for method, op in item.items():
            if method.startswith("x-") or not isinstance(op, dict):
                continue
            if method.lower() not in METHOD_VERB and method.lower() not in {
                "get",
                "post",
                "put",
                "patch",
                "delete",
                "head",
                "options",
                "trace",
            }:
                continue
            oid = op.get("operationId")
            if not needs_rewrite(oid if isinstance(oid, str) else None):
                name = str(oid)
            else:
                name = op_name(method, path)
                rewritten += 1
            base = name
            n = 2
            while name in used:
                name = f"{base}{n}"
                n += 1
            used.add(name)
            op["operationId"] = name
    return rewritten


def load_doc(path: Path) -> dict[str, Any]:
    text = path.read_text(encoding="utf-8")
    # JSON preferred; YAML fallback if PyYAML available
    try:
        doc = json.loads(text)
    except json.JSONDecodeError:
        try:
            import yaml  # type: ignore
        except ImportError as e:
            raise SystemExit(f"spec is not JSON and PyYAML missing: {e}") from e
        doc = yaml.safe_load(text)
    if not isinstance(doc, dict):
        raise SystemExit("invalid OpenAPI document")
    return doc


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--in", dest="inp", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--server-sandbox", required=True)
    ap.add_argument("--server-prod", required=True)
    ap.add_argument("--server-lt", default="")
    ap.add_argument("--server-ee", default="")
    args = ap.parse_args()

    doc = load_doc(Path(args.inp))
    set_servers(
        doc,
        args.server_sandbox,
        args.server_prod,
        args.server_lt or None,
        args.server_ee or None,
    )
    n_one = simplify_oneof_null(doc)
    n_type = sanitize_types(doc)
    n_fmt = sanitize_formats(doc)
    n_schema = ensure_param_schemas(doc)
    n_params = fix_missing_path_params(doc)
    n_ops = normalize_operation_ids(doc)
    out = Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print(
        f"fixed {n_one} oneOf-null, {n_type} type(s), {n_fmt} format(s), "
        f"{n_schema} param schema(s), {n_params} path param(s), "
        f"rewrote {n_ops} operationId(s) → {out}",
        file=sys.stderr,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
