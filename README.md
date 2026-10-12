# py2go

Transpiler from Python 3.13 to Go.

## Status

Phase 7: classes. Methods, `@staticmethod`, `@classmethod`, `@property` with
setters, class attributes with instance shadowing, single and multiple
inheritance with C3 and `super()`, `isinstance`, `type`, the special methods
(`__repr__`, `__eq__`, `__lt__`, the arithmetic family with reflected and
in-place forms, `__len__`, `__getitem__`, `__iter__`, `__call__`, ...),
`@dataclass`, and user-defined exceptions with `raise ... from` and the chain
printed for an uncaught one. `None` as a value, `with` and generators arrive in
phase 8.

Phase 6: functions as values and the full call syntax. Default values, keyword
and keyword-only and positional-only parameters, `*args`, `**kwargs`, `*xs` and
`**d` at the call site, lambdas, nested functions with `nonlocal`, closures,
decorators, `Callable[[A], R]`, `tuple[int, ...]`, `tuple(x)`, and `map`,
`filter` and `key=` for `sorted`, `min`, `max` and `sort`. Argument errors carry
CPython's `TypeError` text, at compile time where the call is static.

Phase 5: `dict`, `set`, list/set/dict comprehensions, generator expressions,
`range`, `zip`, `enumerate` and `reversed` as values, and module-level variables
that functions can read and rebind with `global`. Sets iterate in CPython's order.

Phase 4: exceptions. `raise`, `try`/`except`/`else`/`finally` and `assert` work
with the built-in exception classes, uncaught errors exit with status 1 like
CPython, and runaway recursion raises `RecursionError`. User-defined exception
classes arrived in phase 7.

Phase 3: strings, builtins, sequences. On top of the phase 2 typed core
(functions, `int`/`float`/`bool`/`str`, arithmetic, `if`/`while`/`for`), it
handles `str` methods and slicing, f-strings, `list` and `tuple` with unpacking,
`for` over lists, strings, `enumerate` and `zip`, and the builtins `len`, `int`,
`float`, `str`, `abs`, `min`, `max`, `sum`, `sorted`, `round`, `divmod` and `**`.
`None` as a value is not supported yet.

## Usage

    py2go file.py        # writes file.go next to it

Generated code imports `github.com/amayakmt/py2go/runtime`, so build it inside a
module that requires py2go (or uses a `replace` directive pointing at a checkout).
Python 3.13 is resolved through `uv`.

## Tests

    go test -count=1 ./...

Each case under `testdata/phase2/` to `phase7/` runs under CPython 3.13 (committed
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