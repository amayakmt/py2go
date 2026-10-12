"""Checks the standard library coverage map in docs/roadmap.md.

Every module in sys.stdlib_module_names of the running Python, plus `test`
and the removed `msilib`, must be listed in exactly one "- Phase N, ...:" line.
Run it with the interpreter the project targets:

    uv run --python 3.13 python scripts/check_stdlib_map.py
"""
import re
import sys
from pathlib import Path

if sys.version_info[:2] != (3, 13):
    sys.exit(f"needs Python 3.13, got {sys.version_info.major}.{sys.version_info.minor}")

roadmap = Path(__file__).resolve().parent.parent / "docs" / "roadmap.md"
listed: dict[str, list[int]] = {}
for line in roadmap.read_text().splitlines():
    m = re.match(r"- Phase (\d+), [^:]+: (.+)$", line)
    if not m:
        continue
    for name in m.group(2).split(", "):
        listed.setdefault(name.strip(), []).append(int(m.group(1)))

expected = set(sys.stdlib_module_names) | {"test", "msilib"}
missing = sorted(expected - listed.keys())
unknown = sorted(listed.keys() - expected)
repeated = sorted(n for n, phases in listed.items() if len(phases) > 1)

for label, names in (("missing from the map", missing), ("not in Python 3.13", unknown), ("listed twice", repeated)):
    if names:
        print(f"{label}: {', '.join(names)}")
if missing or unknown or repeated:
    sys.exit(1)
print(f"ok: {len(listed)} modules mapped")
