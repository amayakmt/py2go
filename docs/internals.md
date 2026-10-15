# Internals

## Pipeline

    file.py ──pyrun──▶ JSON AST ──pyast──▶ Go AST of the program
                                     │
              internal/link: joins the project and stdlib modules into one module
                                     │
              internal/codegen: infers types, generates one Go file
                                     │
              go build, against runtime/  ──▶ executable

- `internal/pyrun` runs `dump_ast.py` under CPython 3.13 to get the syntax tree
  as JSON. This is the only use of CPython at transpile time.
- `internal/pyast` decodes that JSON into Go node types (`node.go`).
- `internal/link` replaces `import` statements by the bodies of the imported
  modules, with every top-level name renamed `<module>_pm_<name>` (a package
  `a.b` is `a__b_pm_`). It also does the rewrites that need to see whole
  modules: `namedtuple` and `enum` expansion, `@functools.total_ordering`,
  `unittest.main` test discovery, `sys.stdout = f`, modules that cannot be
  imported. Standard library modules are found in the embedded `stdlib/` tree
  exactly like project modules.
- `internal/codegen` is the compiler. A prepass collects classes, functions,
  globals; type inference joins the types assigned to each name (`joins.go`) and
  falls back to a dynamic `object` (Go `any`) when they differ. Expressions
  compile to typed Go when the types are known and to `runtime.Dyn*` calls
  otherwise. Methods are compiled once per concrete class, so `isinstance(self,
  Sub)` is settled statically.
- `runtime/` is the support library: the object model (`object.go`), integers
  (`bigint.go`, with a 64-bit fast path), strings, containers (`List`, `Dict`,
  `Set`, `VarTuple`), exceptions (`errors.go`), generators and coroutines as
  goroutines with a handoff (`gen.go`, `coro.go`), the GIL (`thread.go`),
  signals, files, codecs, regular expressions (`regex.go`: a backtracking engine
  with CPython's semantics) and the natives of the stdlib modules.

## Object model

- A Python object is a Go `any`. `int` is `int64` or `*big.Int`, `str` is a Go
  `string` (code points are indexed on demand; surrogates are carried in a
  WTF-8 form), `float` is `float64`, `bool`, `None` is `runtime.NoneV`.
- A user class is a Go struct plus an interface (`_pyc_<Class>`): fields for the
  attributes the compiler found, `_pyd` for the rest. Dynamic access goes
  through `PyGetAttr`, `PySetAttr`, `PyMethodNames` that the compiler generates
  per class (`dynhooks.go`).
- Class values (`C` held in a variable) are `*runtime.ClassRef`; the compiler
  registers class attributes, members and classmethods with the runtime
  (`RegisterClassAttr`, `RegisterClassMembers`, `RegisterClassCallable`) so that
  `getattr(cls, name)` and `dir(cls)` work on them.
- Exceptions are `*runtime.PyError{Type, Args, Attrs}`; the class tree is the
  `parents` map in `errors.go`; user exception classes register their chain.
  `try` compiles to `runtime.Catch(func(){...})`.

## Standard library modules

A module is `stdlib/<name>.py` (or a package) plus, if needed, natives in Go:

1. Write the module in the Python subset py2go compiles. `scripts/port_stdlib.py
   SRC DST` copies a module of CPython's `Lib/` without annotations, which is a
   good start; then fix what the compiler rejects (the error message names the
   line). Typical edits: replace class decorators and metaclass tricks, subclass
   of `dict`/`tuple` by composition, `except cls:` with a variable by
   `except BaseException as e: if not isinstance(e, cls): raise`, and
   attributes that start as `None` by `_obj(None)` (an identity function returning
   `object`) so that inference does not fix their type to `None`.
2. Natives: `RegisterNative("mod.fn", func(args []any, kw []KW) any {...})` in a
   `runtime/*native.go` file; call it from Python as `_native("mod.fn", ...)`.
   A native that blocks must release the GIL with `runtime.Blocking(func(){...})`.
3. List the module in `stdlib/manifest.json`; `python3 scripts/stdlib_manifest.py
   --check` verifies the list against `sys.stdlib_module_names`.
4. Add a differential test: `testdata/phaseN/<case>/input.py`, record
   `expected.txt` with `scripts/regen_expected.sh testdata/phaseN/<case>`, and the
   test in `internal/difftest` picks it up.

## Tests

- `internal/difftest`: each case runs under CPython 3.13 (recorded
  `expected.txt`, plus `expected.err`/`expected.exit` for programs that fail) and
  as transpiled Go; stdout, the last stderr line and the exit status must match.
  `TestUnsupported` checks the exact message of what is rejected.
- `internal/codegen`, `internal/link`, `runtime`: unit tests for the generated
  code, the linking and the natives.
- `go test -count=1 ./...` runs everything (about 10 minutes; thread-timing
  tests can fail once on a loaded machine).
