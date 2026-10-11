# Python to Go Transpiler Roadmap

> Rule that runs through every phase: differential testing.
> Each testdata/ file gets run under Python 3.13 and as transpiled Go, and the harness diffs stdout.

## Phase 1: Pipeline

### 1. `dump_ast.py` outputs JSON. 

1. Write the AST dumper. `scripts/dump_ast.py` reads a `.py` file, parses it with `ast.parse`, and prints the tree as JSON. Each node gets a `_type` and a `lineno`.
2. Define Go structs. One struct per node type to support: FunctionDef, Expr, Call, Name, Constant. All satisfy a Node interface.
3. Write the Go decoder. Read the JSON, peek at `_type`, then unmarshal into the matching struct. Use `json.RawMessage` for the peek.
4. Handle unknown nodes. Anything not supported decodes into `Unsupported{Type, Lineno}`
5. Write the code generator. One `emit(node)` function with a type switch. `FunctionDef` becomes a Go `func`. `print(x)` becomes `fmt.Println(x)`. Unsupported nodes return `unsupported: Match at line 12`.
6. Format the output. Run the generated source through `go/format`. If it errors, the generator produced invalid Go.
7. Wire the CLI. `py2go file.py` calls the dumper, decodes, generates, and writes `file.go`.
8. Run the exit test. Take `testdata/hello.py`. Run it with Python, transpile it, `go run` the result, and compare stdout. Byte-identical means phase 1 is done.

