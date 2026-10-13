#!/usr/bin/env python3
"""Writes runtime/unicodedata.json.gz, the data of the unicodedata module, from
CPython 3.13's own unicodedata (Unicode 15.1.0).

    uv run --python 3.13 python scripts/gen_unicodedata.py
"""
import gzip
import json
import sys
import unicodedata
from pathlib import Path

OUT = Path(__file__).resolve().parent.parent / "runtime" / "unicodedata.json.gz"
CPS = [c for c in range(0x110000) if not 0xD800 <= c <= 0xDFFF]


def runs(fn, skip=None):
    out = []
    start = prev = val = None
    for c in range(0x110000):
        v = fn(c)
        if v == skip:
            v = None
        if v is not None and v == val and prev == c - 1:
            prev = c
            continue
        if val is not None:
            out.append([start, prev, val])
        start, prev, val = c, c, v
    if val is not None:
        out.append([start, prev, val])
    return out


def cat(c):
    return unicodedata.category(chr(c)) if not 0xD800 <= c <= 0xDFFF else "Cs"


data = {
    "version": unicodedata.unidata_version,
    "cat": runs(cat, "Cn"),
    "bidi": runs(lambda c: unicodedata.bidirectional(chr(c)) or None),
    "east": runs(lambda c: unicodedata.east_asian_width(chr(c))),
    "comb": runs(lambda c: unicodedata.combining(chr(c)) or None),
    "mirrored": runs(lambda c: 1 if unicodedata.mirrored(chr(c)) else None),
    "decomp": {},
    "decimal": {},
    "digit": {},
    "numeric": {},
    "name": {},
    "nameranges": [],
    "pairs": [],
}
for c in CPS:
    ch = chr(c)
    d = unicodedata.decomposition(ch)
    if d:
        data["decomp"][c] = d
    v = unicodedata.decimal(ch, None)
    if v is not None:
        data["decimal"][c] = v
    v = unicodedata.digit(ch, None)
    if v is not None:
        data["digit"][c] = v
    v = unicodedata.numeric(ch, None)
    if v is not None:
        data["numeric"][c] = repr(v)
    n = unicodedata.name(ch, "")
    if n:
        data["name"][c] = n
# names that are a prefix and the code point: stored as ranges
ranges_ = []
rest = {}
cur = None
for c in sorted(data["name"]):
    n = data["name"][c]
    suffix = format(c, "X")
    if n.endswith("-" + suffix):
        prefix = n[: -len(suffix)]
        if cur and cur[2] == prefix and cur[1] == c - 1:
            cur[1] = c
        else:
            if cur:
                ranges_.append(cur)
            cur = [c, c, prefix]
    else:
        rest[c] = n
if cur:
    ranges_.append(cur)
big = [r for r in ranges_ if r[1] - r[0] >= 8]
small = [r for r in ranges_ if r[1] - r[0] < 8]
for r in small:
    for c in range(r[0], r[1] + 1):
        rest[c] = data["name"][c]
data["name"] = rest
data["nameranges"] = big
# canonical pairs that compose
for c, d in data["decomp"].items():
    parts = d.split()
    if parts[0].startswith("<") or len(parts) != 2:
        continue
    a, b = int(parts[0], 16), int(parts[1], 16)
    if unicodedata.normalize("NFC", chr(a) + chr(b)) == chr(c):
        data["pairs"].append([a, b, c])
OUT.write_bytes(gzip.compress(json.dumps(data, separators=(",", ":")).encode(), 9, mtime=0))
print("wrote", OUT, OUT.stat().st_size, "bytes", file=sys.stderr)
