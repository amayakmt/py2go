#!/usr/bin/env bash
# Measures py2go against CPython 3.13 on the programs of bench/: the time to
# transpile and build, the binary size, the time to start (hello world) and the
# run time. Prints a markdown table; docs/benchmarks.md holds a recorded run.
#
#   scripts/bench.sh            all of bench/*.py
set -euo pipefail
cd "$(dirname "$0")/.."
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
go build -o "$tmp/py2go" ./cmd/py2go
now() { date +%s.%N; }
echo "| program | transpile+build (s) | binary (KiB) | py2go run (s) | CPython run (s) |"
echo "|---|---|---|---|---|"
for src in bench/*.py; do
	name=$(basename "$src" .py)
	t0=$(now)
	"$tmp/py2go" build "$src" -o "$tmp/$name" >/dev/null
	t1=$(now)
	size=$(( $(stat -c %s "$tmp/$name") / 1024 ))
	"$tmp/$name" >/dev/null
	t2=$(now)
	uv run --python 3.13 python "$src" >/dev/null
	t3=$(now)
	printf "| %s | %.2f | %d | %.3f | %.3f |\n" "$name" "$(echo "$t1 - $t0" | bc)" "$size" "$(echo "$t2 - $t1" | bc)" "$(echo "$t3 - $t2" | bc)"
done
