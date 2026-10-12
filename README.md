# py2go

Transpiler from Python 3.13 to Go.

## Status

[![phases](docs/progress/badge-phases.svg)](docs/roadmap.md) [![differential tests](docs/progress/badge-tests.svg)](testdata) [![rejection tests](docs/progress/badge-rejected.svg)](testdata/unsupported)

[![progress map](docs/progress/progress.svg)](docs/roadmap.md)

<sub>A phase is done when its exit test passes (see [decisions](docs/decisions.md)). The pictures are drawn by `go run ./cmd/progress` from the roadmap, the decision log and the test directories; run it after finishing a phase.</sub>

## Usage

    py2go file.py        # writes file.go next to it

Generated code imports `github.com/amayakmt/py2go/runtime`, so build it inside a
module that requires py2go (or uses a `replace` directive pointing at a checkout).
Python 3.13 is resolved through `uv`.

## Tests

    go test -count=1 ./...

Each case under `testdata/phase2/` to `phase8/` runs under CPython 3.13 (committed
`expected.txt`) and as transpiled Go; stdout must match. Cases under
`testdata/unsupported/` assert the exact error and line number.
After adding a case, `scripts/regen_expected.sh testdata/phase2/<case>` records
CPython's output.

## Source language

Python 3.13 only. Pinned in .python-version.

## Scope

- Standard library only
- Type hints required
- Unsupported syntax fails with a line number