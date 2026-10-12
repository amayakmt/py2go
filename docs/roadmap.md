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

### Phase 3: strings, builtins, sequences

> Phase 2 can only compute on scalars. Phase 3 adds the data real programs are made of: text and `list`/`tuple`. `dict` and `set` come later, because they need an insertion-ordered hash runtime. Exceptions stay in phase 4; index and conversion errors panic with CPython's message, the way `ZeroDivisionError` does now.

1. Add the scalar builtins. `int(x)`, `float(x)`, `str(x)`, `bool(x)`, `abs`, `min`/`max` with 2+ arguments, `divmod`, `round` with 1 argument. Each is a runtime function selected by the argument's static type. `int("abc")` panics with `ValueError: invalid literal for int() with base 10: 'abc'`.
2. Support `**`. `int ** int` with a literal non-negative exponent stays `int`; with a negative exponent it is `float`; a variable exponent is rejected, since the result type would depend on its sign. Float operands call `math.Pow` through the runtime, which raises on `0.0 ** -1`.
3. Make `str` real. Python strings index by code point and Go by byte, so the runtime owns every string operation: `len`, indexing (negative indices included), slicing with step, `+`, `*`, `in`, `==`/`<` (already correct). Methods: `upper`, `lower`, `strip`, `split`, `join`, `replace`, `find`, `startswith`, `endswith`, `count`. Out-of-range indexing panics with `IndexError: string index out of range`.
4. Support f-strings and `print(sep=, end=)`. `JoinedStr` becomes a runtime format call. Format specs are limited to `d`, `f`, `.Nf`, `s`, width and alignment. Anything else is rejected with a line number. These are the only keyword arguments in the language so far, so allow them for `print` and reject all others.
5. Extend the type table. Add `list[T]` and `tuple[A, B, ...]`, with `T` any supported type, nested allowed. Anything else is unsupported. A list is `*runtime.List[T]`, not a bare Go slice, because Python lists alias: `b = a` shares storage, and `append` through one name must show through the other.
6. Support list basics. Literals `[1, 2, 3]`, an empty list only when an annotation supplies its type, indexing and index assignment (negative indices included), `len`, `in`, `+`, `*` with an int, slicing, `append`, `pop`, `insert`, `extend`, `reverse`, `sort` with no key, `index`, `count`. Out-of-range access panics with `IndexError: list index out of range`.
7. Support tuples and unpacking. Tuples are immutable fixed-arity values, emitted as Go structs per distinct shape. Support tuple literals and indexing with a literal index. Support `a, b = b, a`, and functions that return tuples with `-> tuple[int, int]` unpacked at the call site. Unpacking arity mismatches are rejected at compile time when the sizes are static.
8. Iterate over sequences. `for x in list`, `for c in str`, `for i, x in enumerate(xs)`, `for a, b in zip(xs, ys)`. Mutating the list while iterating keeps Python's behavior: the iterator checks the index against the current length each step. `range()` keeps its fast path.
9. Give containers Python's `str()` and `repr()`. `print([1, 'a'])` shows `[1, 'a']`, not `[1 a]`; strings inside containers use `repr` quoting rules (single quotes unless the text has one, escapes for `\n` and non-printables). Floats inside containers use `FloatRepr`. Extend `runtime.Str` and add `runtime.Repr`.
10. Decide aliasing and copying, and log it. Assignment shares, slicing copies, `list + list` copies, function arguments share. Write the tests that expose it first: two names for one list, a function that mutates its parameter, `a = b[:]`. Default-less parameters make the mutable-default trap impossible, so note that.
11. Decide equality and ordering for containers. `==` is element-wise for lists and tuples of the same type. Ordering (`<`) is lexicographic. Comparing different container types is rejected, as in phase 2.
12. Write 40+ differential tests under `testdata/phase3/`, one feature each. Include the cases that break naive translations: negative indices, `s[::-1]`, `s[-3:]`, non-ASCII strings (`len("héllo")` is 5, not 6), aliasing, `sorted` stability, `print` of a nested list with strings and floats, and a loop that appends while iterating over `range(len(xs))`.
13. Write `testdata/unsupported/` cases for each new rejection: a list of an unsupported type, a mixed-type list literal, an untyped empty list, a variable exponent, an unsupported format spec, a keyword argument other than `sep`/`end`, `sort(key=...)`, `dict` and `set` annotations.
14. Run the exit test. All differential tests pass under `go test -count=1 ./...`, and every decision above is logged in `docs/decisions.md`. Known limits documented: string methods cover only the listed set, and every string operation that indexes by code point is O(n).


### Phase 4: exceptions

> Until now an error panicked with CPython's message and could not be caught. Phase 4 makes exceptions real: `raise`, `try`/`except`/`else`/`finally`, `assert`, and uncaught errors that exit the way CPython does. Only the built-in exception classes exist, because user-defined ones need classes. `with` and generators stay out.

1. Give the runtime real exceptions. `PyError` carries its class name and its arguments, and the runtime knows the built-in hierarchy (`ZeroDivisionError` is an `ArithmeticError` is an `Exception`). Add `NewError`, `Match` and `Catch`. `str(e)` and `repr(e)` follow CPython: no argument gives `''` and `ValueError()`, one argument gives its `str` and `ValueError('x')`, several give a tuple. Every panic that phases 2 and 3 raise keeps its message and class.
2. Report uncaught exceptions the way CPython does. The generated `main` defers `runtime.Report`, which prints `Traceback (most recent call last):` and `Type: message` to stderr and exits with status 1. The harness builds each program, then compares stdout, the exit status and the last line of stderr.
3. Decode `Try`, `ExceptHandler`, `Raise` and `Assert`. `TryStar` (`except*`) and `with` stay unsupported.
4. Add the exception class table and an `exception` type. A name bound by `except ... as e` has that type. It supports `str(e)`, `repr(e)`, `print(e)`, f-string fields, `type(e).__name__` and `raise e`. It cannot be annotated, stored or passed around, so it never outlives its handler.
5. Support `raise` in these forms: `raise Cls`, `raise Cls(args)` with any supported value as an argument, `raise e`, and a bare `raise` inside a handler. `raise X from Y` is accepted and the cause is ignored, since only stderr would show it. Raising anything else, or an unknown class, is rejected.
6. Support `assert test` and `assert test, message`, which raise `AssertionError`.
7. Support `try`/`except`. Handlers are tried in order. A handler can name one class, a tuple of classes, or nothing (bare `except:`), and can bind `as e`. An exception no handler matches continues up. Errors from the runtime, such as `IndexError` from `xs[9]` or `ValueError` from `int("x")`, are caught by the class CPython gives them.
8. Support `else` and `finally`. `else` runs only when the body finished normally. `finally` always runs, and a `return`, `break` or `continue` inside it replaces a pending exception or jump.
9. Make `break`, `continue` and `return` work inside a `try`. Go has no exceptions, so the body becomes a closure and these statements become a status code that the code after the closure acts on. Check nesting: a `try` in a loop, in a `finally`, in a `try`.
10. Declarations made at the top level of a `try` body are visible afterwards, because the usual pattern is `x: int = int(s)` in the body and `x = 0` in the handler. Hoist them out of the closure.
11. Model the recursion limit. Functions that can reach themselves through calls count their depth, and the 1000th nested call raises `RecursionError: maximum recursion depth exceeded`, so infinite recursion fails like CPython and can be caught instead of crashing Go.
12. Write 40+ differential tests under `testdata/phase4/`. Include the cases that break naive translations: `finally` order with `return`, a `return` in `finally` that swallows an exception, `else` skipped by `break`, re-raise from a nested handler, handlers matched by parent class, an exception escaping a loop, `continue` in `finally`, uncaught errors from helpers and from `raise` (exit status 1), and `RecursionError` caught and uncaught.
13. Write `testdata/unsupported/` cases for each new rejection: user-defined exception classes, `except*`, an unknown class in `except` or `raise`, raising a non-exception, a bare `raise` outside a handler, storing or returning an exception, `e.args`, `with`.
14. Run the exit test. All differential tests pass under `go test -count=1 ./...`, and every decision above is logged in `docs/decisions.md`. Known limits documented: tracebacks show no frames, exception chaining is not recorded, and the recursion limit counts only functions that can recurse.
