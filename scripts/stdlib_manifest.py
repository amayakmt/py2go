"""Keeps stdlib/manifest.json in step with the coverage map in docs/roadmap.md.

One entry per module of sys.stdlib_module_names (plus `test` and `msilib`):
its phase, its kind (L: compiled from Lib/, N: native Go, P: platform
specific), its status (todo, partial, done, platform) and the tests that are
allowed to fail. Statuses and expected failures are kept when the file is
rewritten.

    uv run --python 3.13 python scripts/stdlib_manifest.py          # rewrite
    uv run --python 3.13 python scripts/stdlib_manifest.py --check  # verify
"""
import json
import re
import sys
from pathlib import Path

if sys.version_info[:2] != (3, 13):
    sys.exit(f"needs Python 3.13, got {sys.version_info.major}.{sys.version_info.minor}")

root = Path(__file__).resolve().parent.parent
manifest_path = root / "stdlib" / "manifest.json"
KINDS = {"from `Lib/`": "L", "native Go": "N", "platform-specific": "P"}

planned: dict[str, tuple[int, str]] = {}
for line in (root / "docs" / "roadmap.md").read_text().splitlines():
    m = re.match(r"- Phase (\d+), (from `Lib/`|native Go|platform-specific): (.+)$", line)
    if not m:
        continue
    for name in m.group(3).split(", "):
        planned[name.strip()] = (int(m.group(1)), KINDS[m.group(2)])

expected = set(sys.stdlib_module_names) | {"test", "msilib"}
old = {}
if manifest_path.exists():
    old = {e["name"]: e for e in json.loads(manifest_path.read_text())["modules"]}

entries = []
for name in sorted(expected):
    phase, kind = planned.get(name, (0, "?"))
    prev = old.get(name, {})
    entries.append({
        "name": name,
        "phase": phase,
        "kind": kind,
        "status": prev.get("status", "platform" if kind == "P" else "todo"),
        "expected_failures": prev.get("expected_failures", []),
    })

if "--check" in sys.argv:
    problems = []
    if not manifest_path.exists():
        sys.exit("stdlib/manifest.json is missing")
    names = set(old)
    if names != expected:
        problems.append(f"modules differ: missing {sorted(expected - names)}, unknown {sorted(names - expected)}")
    for e in entries:
        o = old.get(e["name"])
        if o and (o["phase"], o["kind"]) != (e["phase"], e["kind"]):
            problems.append(f"{e['name']}: manifest says phase {o['phase']} {o['kind']}, the map says phase {e['phase']} {e['kind']}")
        if o and o["status"] not in ("todo", "partial", "done", "platform"):
            problems.append(f"{e['name']}: unknown status {o['status']}")
    if problems:
        print("\n".join(problems))
        sys.exit(1)
    done = sum(1 for e in old.values() if e["status"] == "done")
    partial = sum(1 for e in old.values() if e["status"] == "partial")
    print(f"ok: {len(old)} modules, {done} done, {partial} partial")
else:
    manifest_path.write_text(json.dumps({"modules": entries}, indent=1) + "\n")
    print(f"wrote {manifest_path.relative_to(root)}: {len(entries)} modules")
