# py2go

Transpiler from Python 3.13 to Go.

## Status

Phase 3: strings, builtins, sequences. On top of the phase 2 typed core
(functions, `int`/`float`/`bool`/`str`, arithmetic, `if`/`while`/`for`), it
handles `str` methods and slicing, f-strings, `list` and `tuple` with unpacking,
`for` over lists, strings, `enumerate` and `zip`, and the builtins `len`, `int`,
`float`, `str`, `abs`, `min`, `max`, `sum`, `sorted`, `round`, `divmod` and `**`.
`dict`, `set` and comprehensions are not supported yet.

## Usage

    py2go file.py        # writes file.go next to it

Generated code imports `github.com/amayakmt/py2go/runtime`, so build it inside a
module that requires py2go (or uses a `replace` directive pointing at a checkout).
Python 3.13 is resolved through `uv`.

## Tests

    go test -count=1 ./...

Each case under `testdata/phase2/` and `testdata/phase3/` runs under CPython 3.13 (committed
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