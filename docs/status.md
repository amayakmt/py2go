# Where the project stands against the phase 20 criteria

Measured on 2026-10-16. The end goal of the roadmap is not met; this page says by how
much, so that the next piece of work can be chosen from numbers.

| # | criterion | state |
|---|---|---|
| 1 | every module of `sys.stdlib_module_names` is `done` or `platform` | **not met.** 292 modules: 0 done, 122 partial, 14 platform, 156 todo. Of the 156, 83 are private C accelerators (`_bisect`, `_json`, ...) whose public module exists and works, 73 are public modules. |
| 2 | `py2go check Lib/` accepts every file | **not met.** The 153 top-level files of `Lib/`, each checked alone: 15 accepted, 138 rejected. `Lib/test` is not attempted. |
| 3 | conformance harness over `Lib/test/test_*.py` | **not built.** `stdlib/expected_failures.json` exists with no entries. The tests that exist are the 1 000 differential programs under `testdata/` (recorded from CPython 3.13) and the unit tests. |
| 4 | the language conformance suites pass | **not run** (they need `unittest` over the module-level machinery of `Lib/test`). |
| 5 | differential fuzzing in CI | **not done.** There is no CI and the random-program generator of phases 3 and 4 was not kept. |
| 6 | performance measured and published | **done in part**: `scripts/bench.sh`, [benchmarks.md](benchmarks.md). Annotated code is 5 to 8 times faster than CPython; code without annotations is slower on loops and floats. No pyperformance run, no committed baseline. |
| 7 | release the toolchain | **partly**: `scripts/release.sh` builds linux and darwin on amd64 and arm64 with reproducible bytes (checked on linux/amd64). Not met: Windows builds, a parser embedded in the binary (it needs `uv` and Python 3.13), a cgo-free variant flag (the build is cgo-free). |
| 8 | documentation | **done**: [user guide](user-guide.md), [internals](internals.md), [known differences](known-differences.md). |
| 9 | exit test: `go test -count=1 ./...` and the harness pass | `go test -count=1 -timeout 40m ./...` passes (about 12 minutes; one thread-timing test can fail on a loaded machine). The harness does not exist. |

## The 138 rejected files of `Lib/`

First reasons, by number of files (a file is counted under the first thing that stops it):

| files | reason |
|---|---|
| 22 | imports a standard library module that is not there (`weakref`, `numbers`, `tokenize`, `linecache`, `locale`, `importlib`, `ssl`, `termios`, ...) |
| 13 | a base class that is not a plain name (`class C(mod.Base)`, `Generic[T]`) |
| 10 | `from m import x` of a missing module |
| 10 | a call to a built-in or function that does not exist (`globals`, `locals`, `vars` forms, `exec`, ...) |
| 9 | deriving from a built-in type (`dict`, `list`, `tuple`, `str`, `Exception` with `__new__`) |
| 8 | metaclasses (`metaclass=ABCMeta`, `EnumMeta`) |
| 8 | `__class_getitem__` |
| 4 | `__del__` |
| 4 | class decorators other than `@dataclass` |
| 3 | module used as a value in a form that is not supported |
| 4 | an unknown base class (`class C(SomeLibraryClass)`) |
| 3 | a class defined inside a class |

These are language-level gaps, not missing library code: each one needs a compiler
feature (subclassing of built-in types, metaclasses and `__class_getitem__`,
`weakref`, class decorators), and most of `Lib/` needs several of them at once.

## What would move the numbers most

1. Subclassing `dict`, `list`, `tuple`, `str`, `int` and `object.__new__`: it unblocks
   `collections`, `enum`, `numbers`, `argparse`'s unported pieces, `configparser`,
   `email` and a third of the rejected files.
2. Metaclasses and class decorators, enough for `abc.ABCMeta`, `enum`, `typing.Generic`.
3. `weakref` and `__del__` (finalizers on garbage collection).
4. A run-time parser and bytecode compiler (`eval`, `exec`, `compile`, `ast`, `dis`,
   `inspect.getsource`, `doctest`, `pdb`): a second implementation of the language.
5. Type inference that keeps more code out of the dynamic runtime, for the speed of
   code without annotations.
6. A Windows target and an embedded parser, for the release.
