# Python to Go Transpiler Roadmap

> Rule that runs through every phase: differential testing.
> Each testdata/ file gets run under Python 3.13 and as transpiled Go, and the harness diffs stdout.

### Phase 1: pipeline

1. Write the AST dumper. `scripts/dump_ast.py` reads a `.py` file, parses it with `ast.parse`, and prints the tree as JSON. Each node gets a `_type` and a `lineno`.
2. Define Go structs. One struct per node type to support: FunctionDef, Expr, Call, Name, Constant. All satisfy a Node interface.
3. Write the Go decoder. Read the JSON, peek at `_type`, then unmarshal into the matching struct. Use `json.RawMessage` for the peek.
4. Handle unknown nodes. Anything not supported decodes into `Unsupported{Type, Lineno}`
5. Write the code generator. One `emit(node)` function with a type switch. `FunctionDef` becomes a Go `func`. `print(x)` becomes `fmt.Println(x)`. Unsupported nodes return `unsupported: Match at line 12`.
6. Format the output. Run the generated source through `go/format`. If it errors, the generator produced invalid Go.
7. Wire the CLI. `py2go file.py` calls the dumper, decodes, generates, and writes `file.go`.
8. Run the exit test. Take `testdata/hello.py`. Run it with Python, transpile it, `go run` the result, and compare stdout. Byte-identical means phase 1 is done.

### Phase 2: typed core

1. Extend the dumper. Emit a `pytype` tag on every `Constant` (`int`, `float`, `str`, `bool`, `none`). Go's JSON decoding turns both `1` and `1.0` into `float64`, so without the tag you can't tell them apart.
2. Define the new node structs: `arguments`, `arg`, `Return`, `AnnAssign`, `Assign`, `AugAssign`, `BinOp`, `UnaryOp`, `Compare`, `BoolOp`, `If`, `While`, `For`, `Break`, `Continue`, `Pass`. Operators (`Add`, `Sub`, `FloorDiv`, ...) are nodes too, so decode them into a plain string field.
3. Add a type table. Map `int` to `int64`, `float` to `float64`, `bool` to `bool`, `str` to `string`. Any other annotation returns `unsupported`.
4. Support parameters and return types. `def add(a: int, b: int) -> int` becomes `func add(a int64, b int64) int64`. Reject defaults, `*args`, `**kwargs`, and keyword-only parameters.
5. Add a symbol table. Track each name's declared type per scope. `x: int = 5` declares. A later `x = 6` or `x += 1` checks that `x` exists. A plain `Assign` to an undeclared name fails as unsupported, since no hint means no type.
6. Give every expression a type. Codegen computes it bottom-up, so it knows which Go operator to emit. Mixed `int` and `float` operands get an explicit `float64(x)` conversion, because Go won't promote them.
7. Support the operators.
   - `/` always returns float, so convert both operands first.
   - `//` and `%` floor toward negative infinity. Emit calls to `runtime.FloorDiv` and `runtime.Mod`.
   - `and`/`or` are supported on `bool` operands only. Python returns an operand, Go returns a bool.
   - Chained comparisons (`a < b < c`) are rejected for now.
   - `**` is rejected for now.
8. Support control flow: `if`/`elif`/`else`, `while`, `break`, `continue`, `pass`, and `for` over `range()` with 1 to 3 arguments. Loop `else` clauses are rejected.
9. Create `runtime/`. Add `runtime.Print(args ...any)` with Python's formatting: `True`/`False`, `1.0` not `1`, `1e+16`, `1e-05`, space-separated. Change codegen so `print` emits `runtime.Print` instead of `fmt.Println`.
10. Decide division by zero. Python raises `ZeroDivisionError`, and Go panics on integer division and returns `Inf` on floats. For now the runtime helpers panic with a Python-style message. Real exceptions arrive in phase 4.
11. Write 30+ differential tests under `testdata/phase2/`, one feature each. Include the cases that break naive translations: `-7 // 2`, `-7 % 3`, `7 / 2`, `print(1.0)`, `print(True)`, `range(10, 0, -2)`, and nested loops with `break`.
12. Write `testdata/unsupported/` cases that assert the exact error and line number for each rejected feature.
13. Run the exit test. All differential tests pass under `go test -count=1 ./...`, and every decision above is logged in `docs/decisions.md`. Overflow is documented as a known limit: `int64` wraps where Python's `int` does not.