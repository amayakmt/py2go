#!/usr/bin/env bash
# Compares py2go's tokenize module (Lib/tokenize.py over the Go tokenizer in
# runtime/tokenizer.go) with CPython 3.13's on every .py file of CPython's
# Lib/, or on the files given. The same program prints the tokens of each
# file; it runs under CPython and compiled by py2go, and the outputs must be
# identical.
#
#   scripts/compare_tokenizer.sh            all of Lib/
#   scripts/compare_tokenizer.sh a.py b.py  just these files
set -euo pipefail
cd "$(dirname "$0")/.."

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cat > "$work/dump_tokens.py" <<'PY'
import io
import sys
import tokenize

for path in sys.argv[1:]:
    with open(path, encoding="utf-8") as f:
        src = f.read()
    print("FILE", path)
    try:
        for t in tokenize.generate_tokens(io.StringIO(src).readline):
            print(t.type, repr(t.string), t.start, t.end, repr(t.line))
    except Exception as e:
        print("ERROR", type(e).__name__, e)
PY

if [ $# -gt 0 ]; then
	files=("$@")
else
	lib=$(uv run -q --python 3.13 python -c 'import sysconfig; print(sysconfig.get_paths()["stdlib"])')
	mapfile -t files < <(find "$lib" -name '*.py' -not -path '*/site-packages/*' | sort)
fi

go run ./cmd/py2go build "$work/dump_tokens.py" -o "$work/dump_tokens" >/dev/null
uv run -q --python 3.13 python "$work/dump_tokens.py" "${files[@]}" > "$work/cpython.txt"
"$work/dump_tokens" "${files[@]}" > "$work/py2go.txt"
if diff -q "$work/cpython.txt" "$work/py2go.txt" >/dev/null; then
	echo "same tokens for ${#files[@]} files"
else
	diff "$work/cpython.txt" "$work/py2go.txt" | head -40
	exit 1
fi
