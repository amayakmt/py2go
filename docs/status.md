# Where the project stands against the phase 20 criteria

Measured on 2026-10-17. The end goal of the roadmap is not met; this page says by how
much, so that the next piece of work can be chosen from numbers.

| # | criterion | state |
|---|---|---|
| 1 | every module of `sys.stdlib_module_names` is `done` or `platform` | **not met.** 292 modules: 0 done, 138 partial, 14 platform, 140 todo. Of the 140, 83 are private C accelerators (`_bisect`, `_json`, ...) whose public module exists and works, 57 are public modules. |
| 2 | `py2go check Lib/` accepts every file | **not met.** The 153 top-level files of `Lib/`, each checked alone: 33 accepted, 120 rejected (15 and 138 on 2026-10-16). `Lib/test` is not attempted. |
| 3 | conformance harness over `Lib/test/test_*.py` | **not built.** `stdlib/expected_failures.json` exists with no entries. The tests that exist are the differential programs under `testdata/` (recorded from CPython 3.13) and the unit tests. |
| 4 | the language conformance suites pass | **not run** (they need `unittest` over the module-level machinery of `Lib/test`). |
| 5 | differential fuzzing in CI | **not done.** There is no CI and the random-program generator of phases 3 and 4 was not kept. `scripts/compare_tokenizer.sh` compares the tokenizer with CPython's on all of `Lib/`. |
| 6 | performance measured and published | **done in part**: `scripts/bench.sh`, [benchmarks.md](benchmarks.md). Annotated code is 5 to 8 times faster than CPython; code without annotations is slower on loops and floats. No pyperformance run, no committed baseline. |
| 7 | release the toolchain | **partly**: `scripts/release.sh` builds linux and darwin on amd64 and arm64 with reproducible bytes (checked on linux/amd64). Not met: Windows builds, a parser embedded in the binary (it needs `uv` and Python 3.13), a cgo-free variant flag (the build is cgo-free). |
| 8 | documentation | **done**: [user guide](user-guide.md), [internals](internals.md), [known differences](known-differences.md). |
| 9 | exit test: `go test -count=1 ./...` and the harness pass | `go test -count=1 -timeout 60m ./...` passes (about 15 minutes; one thread-timing test can fail on a loaded machine). The harness does not exist. |

## The 120 rejected files of `Lib/`

First reasons, by number of files (a file is counted under the first thing that stops it):

| files | reason |
|---|---|
| 36 | imports a module that is not there: C accelerators that the public module does without (`_decimal`, `_struct`, `_heapq`, `_csv`, `_datetime`, ...), and `termios`, `tty`, `sysconfig`, `marshal`, `dis`, `bz2`, `webbrowser`, `py_compile`, ... |
| 15 | a base class that is not a plain name (`class C(mod.Base)`, `Generic[T]`, a call) |
| 6 | `__new__` outside subclasses of immutable built-in types |
| 6 | a call to a function that does not exist (`exec`, `compile`, `sys._getframe`, ...) |
| 4 | an unknown base class |
| 3 | metaclasses other than `ABCMeta` |
| 3 | a class defined inside a function or an `if` |
| 3 | an `if` in a class body |
| 3 | attributes of functions (`f.attr = v`) |
| 3 | decorators on methods other than the built-in ones |
| 3 | a module used as a value in a form that is not supported |
| 3 | an undefined name (names made by `globals()` updates and the like) |
| 32 | other single causes (namedtuple calls that are not literal, statements in class bodies, assignment shapes, type mismatches) |

The files are CPython's own sources, unedited; the modules py2go ships are ported from
them with edits (each marked "py2go"), which is why the library works where its
sources are rejected.

## What would move the numbers most

1. A run-time parser and bytecode compiler (`eval`, `exec`, `compile`, `ast`, `dis`,
   `inspect.getsource`, `doctest`, `pdb`): a second implementation of the language.
   The tokenizer exists (`runtime/tokenizer.go`).
2. Metaclasses beyond `ABCMeta`, `__new__` in general, and function attributes.
3. Classes in `if` blocks and functions, and base classes that are expressions.
4. The C libraries without cgo: `sqlite3`, `bz2`, `lzma` would each need a pure Go
   implementation.
5. Type inference that keeps more code out of the dynamic runtime, for the speed of
   code without annotations.
6. A Windows target and an embedded parser, for the release.
