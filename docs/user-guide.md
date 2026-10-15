# User guide

py2go turns a Python 3.13 program into a Go program and builds it. The result
is a native executable; the standard library the program uses is joined into
it.

## What you need

- Go 1.27 or later on `PATH` (the generated code is compiled with it),
- `uv` and Python 3.13: the parser is CPython's own `ast` module, run by `uv run
  --python 3.13` to dump the tree (see [known differences](known-differences.md)),
- a checkout of this repository, or a `py2go` binary built from it
  (`scripts/release.sh`) with `$PY2GO_ROOT` pointing at the checkout: the generated
  program imports `github.com/amayakmt/py2go/runtime`.

No C compiler is needed or used (`CGO_ENABLED=0`).

## Commands

    py2go file.py                 # write file.go next to it
    py2go build file.py [-o out]  # join the imported modules, generate Go, build
    py2go build dir [-o out]      # a directory with __main__.py
    py2go run file.py [args...]   # build, then run with the arguments; same exit status
    py2go check path...           # transpile without writing; list what is rejected and why

`py2go check` is the measure of coverage: it prints every file that is rejected
with the line and the reason, then a count and the most frequent reasons.

## Writing programs

Plain Python 3.13 works. Annotations are optional:

- **Unannotated code** runs through the dynamic operations of the runtime: any
  size of `int`, values of any type in a list or dict, attributes added at will
  to instances.
- **Annotated code** compiles to typed Go and is much faster. An annotation is
  a contract: assigning a `str` to a variable annotated `int` is a compile
  error, and an annotated `int` is 64 bits (`OverflowError` beyond that, on `*`,
  `**` and `<<`).

Mixing is allowed; the compiler infers the types of unannotated code by joining
the types of everything that is assigned. If inference cannot settle a type, the
variable is a dynamic object.

Unsupported syntax or library use fails at transpile time with the file, the
line and the reason (`unsupported: ... at line N`), never by changing what the
program means.

## Standard library

`stdlib/manifest.json` lists every module of Python 3.13 with its state
(`done`, `partial`, `todo`, `platform`). Most modules are Python files in
`stdlib/` joined into the program, over natives in `runtime/`. A module that
needs a C library (`ssl`, `ctypes`, `curses`, `tkinter`, `sqlite3`, `bz2`, `lzma`,
`xml`) or Windows is not available: importing it raises
`ModuleNotFoundError`, as in a CPython built without it, where the compiler can
tell; otherwise the transpiler rejects the import.

## Tests of your own

`unittest` and `unittest.mock` are there. Call `unittest.main()` in the main
file; the linker hands it the test classes of that file. Decorators `skip`,
`skipIf`, `skipUnless` and `expectedFailure` on test methods are supported.

## Performance

Annotated numeric code is typically several times faster than CPython; code
without annotations that loops over numbers can be slower, since each operation
is dynamic. [benchmarks.md](benchmarks.md) has a recorded run.
