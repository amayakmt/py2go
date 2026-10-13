#!/usr/bin/env bash
# Regenerates expected.txt (and expected.err) for differential test cases by
# running each input.py under CPython 3.13. Review the diff before committing:
# these files are the oracle the transpiled Go is compared against.
#
#   scripts/regen_expected.sh                       all of testdata/phase2 to phase10
#   scripts/regen_expected.sh testdata/phase2/foo   just one case
#
# expected.txt holds stdout. If the program exits with an error, expected.err
# holds the last line of its stderr (e.g. "ZeroDivisionError: ..."), and the
# harness then requires the Go program to fail with the same message.
set -euo pipefail
cd "$(dirname "$0")/.."

dirs=("$@")
if [ ${#dirs[@]} -eq 0 ]; then
	dirs=(testdata/phase[2-9]/*/ testdata/phase1[0-9]/*/)
fi

for dir in "${dirs[@]}"; do
	dir=${dir%/}
	rm -f "$dir/expected.err" "$dir/expected.exit"
	if uv run --python 3.13 python "$dir/input.py" >"$dir/expected.txt" 2>"$dir/.stderr"; then
		:
	else
		code=$?
		tail -n 1 "$dir/.stderr" >"$dir/expected.err"
		# A program that ends with sys.exit(n) has status n.
		if [ "$code" != 1 ]; then echo "$code" >"$dir/expected.exit"; fi
	fi
	rm -f "$dir/.stderr"
done
