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

> The end goal: any Python 3.13 source file transpiles to Go and runs with the same observable behaviour as CPython 3.13, and every module of the 3.13 standard library can be imported and works. Phases 1 to 4 built a typed subset. Phases 5 to 8 finish the language inside that typed world. Phase 9 is the pivot, because the standard library is untyped, dynamic Python and a type-hint-only transpiler cannot compile it. Phases 10 to 19 add the data model and the standard library in dependency order. Phase 20 proves it.

How the standard library gets done, in three parts:

- **L, from `Lib/`.** Most of the standard library is pure Python. The transpiler compiles CPython's own `Lib/` sources. This is also the best test of the transpiler: if `Lib/` compiles and passes its own tests, the transpiler is right.
- **N, native Go.** Modules that CPython implements in C (`_io`, `_socket`, `math`, `itertools`, `_sre`, `zlib`, ...) are written in Go in `runtime/`, with the same behaviour, including error messages. Where Go's standard library gives different bytes or digits than the C library CPython links (zlib output, libm results, OpenSSL, SQLite, expat), the runtime links the same C library through cgo, with a pure Go fallback that documents where it can differ.
- **P, platform-specific.** Modules that exist only on one OS (`msvcrt`, `winreg`, `winsound`, `_winapi`, `nt`) behave exactly as CPython does on the platform being built for. On Linux that means `ModuleNotFoundError`. `msilib` was removed in 3.13, so it must fail to import the way CPython 3.13 does.

The oracle changes along the way. Until phase 11 each feature has differential tests under `testdata/` that compare stdout, exit status and the last line of stderr with CPython. From phase 12 on, CPython's own test suite (`Lib/test/test_<module>.py`) runs under CPython and under the transpiled build, and the two are compared test by test, with a committed list of expected failures that must shrink and never grow.

Some behaviour cannot match and is out of scope for all phases: `id()` values and `repr` of objects that show a memory address, finalizer timing that depends on reference counting (where it can be observed it is emulated in phase 9), hash randomization when `PYTHONHASHSEED` is not set, thread scheduling order, memory usage, and third-party C extensions.

### Phase 5: dict, set, comprehensions, module globals

> The next gap in real programs. Everything here stays inside the typed world.

1. Add an insertion-ordered dict. `dict[K, V]` is `*runtime.Dict[K, V]`, a pointer type so aliasing works like lists. It keeps a slice of entries plus an index, so order survives deletion and reinsertion as in CPython.
2. Decide what a key can be: `int`, `str`, `bool` and `tuple`s of those, with CPython's rules that `1 == True`. `float` keys and mixed numeric keys wait for phase 10, when the numeric tower is complete.
3. Support dict literals, `d[k]`, `d[k] = v`, `del d[k]`, `in`, `len`, `get`, `pop`, `popitem`, `setdefault`, `update`, `copy`, `clear`, `|` and `|=`, `fromkeys`, and equality. A missing key raises `KeyError`, and `str(KeyError('k'))` shows the quoted key.
4. Support the views `keys()`, `values()` and `items()` as live views with their own types, and iteration over a dict. Changing a dict's size during iteration raises `RuntimeError: dictionary changed size during iteration`.
5. Add `set[T]` with literals, `add`, `remove`, `discard`, `pop`, `union`, `intersection`, `difference`, `symmetric_difference`, `issubset`, `issuperset`, `isdisjoint`, the operators `| & - ^` and their in-place forms, comparisons, and `{1, 2}` / `set()` printing.
6. Reproduce CPython's set and dict hash layout for ints and tuples of ints, so that iterating a set of small ints prints in CPython's order. Strings hash with a fixed seed when `PYTHONHASHSEED=0` and tests sort their output otherwise.
7. Support list, set and dict comprehensions with several `for` clauses and `if` filters. A comprehension has its own scope, which maps to a Go closure, and its variables do not leak.
8. Support generator expressions as the argument of `sum`, `min`, `max`, `any`, `all`, `sorted`, `list`, `set`, `dict`, `str.join` and `for`. They are evaluated eagerly until phase 8 gives them real laziness.
9. Add `any`, `all`, `list(x)`, `set(x)`, `dict(x)`, `reversed`, and `sorted(d)`. Add `range` as a value with a type of its own (`len`, indexing, `in`, `list(range(n))`, `sum(range(n))`) and `zip` and `enumerate` as lazy values for `list`, `dict`, `set` and `for`. Generator expressions and `reversed` are lazy iterators, so `any(f(x) for x in xs)` stops at the first true value. `tuple(x)` waits for phase 6, which adds variable-length tuples.
10. Let functions read module-level variables and rebind them with `global`. Module variables become package-level Go variables, a read before the assignment has run raises `NameError`, and the "module-level variable used inside a function" error disappears.
11. Write 60+ differential tests under `testdata/phase5/`, including insertion order after deletes, `d[k] += 1` counting, nested dicts of lists, set algebra, comprehension scoping and closure capture in a comprehension, and the error cases (`KeyError`, mutation during iteration).
12. Write `testdata/unsupported/` cases for each new rejection, then run the exit test: all differential tests pass under `go test -count=1 ./...` and every decision is in `docs/decisions.md`.

### Phase 6: functions as values and the full call syntax

> Today a function can only be called by name with positional arguments. This phase adds everything else about functions.

1. Accept the names from `typing` and `collections.abc` that annotations need (`Callable`, `Optional`, `Any`, `Union`, `Iterable`, ...) as built-in annotation names through a fixed table. The real modules arrive in phase 12.
2. Make functions values. A function can be assigned, stored in a list, tuple or dict, passed and returned. Function types are written `Callable[[int, str], bool]` and become Go function types.
3. Support `lambda`, nested `def`, and closures. Closures share variables with the enclosing scope the way Python's cells do, including late binding in loops. Add `nonlocal` and keep `global` from phase 5.
4. Support defaults, keyword arguments at the call site, keyword-only and positional-only parameters, `*args`, `**kwargs`, and `f(*xs, **d)` at the call. `*args: int` has type `tuple[int, ...]`, so the type system gains variable-length tuples and `tuple(x)`, and `**kwargs: int` has type `dict[str, int]`. Defaults are evaluated once at definition time, so the mutable default trap behaves as in Python.
5. Report call errors the way CPython does, as catchable `TypeError`s with the exact messages: `f() missing 1 required positional argument: 'b'`, `f() takes 2 positional arguments but 3 were given`, `f() got an unexpected keyword argument 'c'`, `f() got multiple values for argument 'a'`. Calls that can be checked statically are rejected at compile time with the same text.
6. Support decorators on functions, including decorators with arguments and stacked decorators. A decorator is a function from function to function, so this follows from steps 2 to 4.
7. Allow docstrings. A docstring is stored as the function's `__doc__` and no longer an unsupported expression statement. Add `__name__`, `__qualname__`, `__doc__`, `__defaults__` and `__kwdefaults__`.
8. Add `map`, `filter`, `zip` and `enumerate` as values consumed by `for`, `list`, `tuple`, `sum` and the like. Add `key=` to `sorted`, `min`, `max` and `list.sort`, and `reverse=` stays. Real lazy iterator objects come in phase 8.
9. Extend the recursion guard to calls through function values: a function that calls a value it cannot resolve is treated as possibly recursive.
10. Write 60+ differential tests under `testdata/phase6/` including closure counters, the late-binding loop trap, the mutable default trap, decorators that wrap and count calls, every `TypeError` message, `*args` forwarding, and recursion through a lambda.
11. Write the unsupported cases and run the exit test.

### Phase 7: classes

> Classes are the largest language feature still missing. They also unlock user-defined exceptions, so `raise MyError` stops being rejected.

1. Compile `class` statements to Go structs. An instance is a pointer, attributes declared in the class body or assigned in `__init__` become typed fields, and class attributes become package-level state with class-level lookup.
2. Support methods, `self`, `__init__`, `@staticmethod`, `@classmethod`, `@property` with setter and deleter, `__slots__` (accepted, not required), and class-level attributes shared across instances with the correct shadowing on assignment.
3. Support inheritance: single and multiple, with C3 method resolution order, `super()` with and without arguments, `isinstance`, `issubclass`, and `type(x)`. Method calls resolve through a per-class table, so overriding works through a base-class reference.
4. Support the special methods that Python calls implicitly: `__repr__`, `__str__`, `__format__`, `__bool__`, `__len__`, `__contains__`, `__getitem__`, `__setitem__`, `__delitem__`, `__iter__`, `__next__`, `__call__`, `__enter__`, `__exit__`, `__hash__`, `__eq__` and all the ordering methods, and the arithmetic family with its reflected (`__radd__`) and in-place (`__iadd__`) forms and `NotImplemented`. `print`, `str`, `repr`, `sorted`, `in`, `len` and the operators route through them.
5. Support the attribute hook `__getattr__`. `__getattribute__`, `__setattr__`, `__delattr__` and the descriptor protocol (`__get__`, `__set__`, `__delete__`) move to phase 9, where attributes become dynamic; `property` and methods are handled by the compiler until then.
6. Support `@dataclass` natively for the common options (`frozen`, `order`, `eq`, `field(default=..., default_factory=...)`, `__post_init__`) until the real `dataclasses` module is compiled from `Lib/` in phase 12.
7. Support user-defined exceptions: classes deriving from the built-in hierarchy, with `args`, a custom `__init__` and `__str__`, `e.__cause__` and `e.__context__` filled in by `raise ... from` and by raising inside a handler, `e.args`, and `except` clauses that match user classes. `e.with_traceback` and `e.__traceback__` move to phase 9 with real frames.
8. Print the chain of `__cause__` and `__context__` of an uncaught exception ("The above exception was the direct cause of the following exception"). The `File "...", line N, in f` frame lines, the caret lines and `traceback` objects move to phase 9, where the generator emits `//line file.py:N` directives and frames become real.
9. Make the recursion guard and the call graph aware of methods and virtual dispatch.
10. Write 80+ differential tests under `testdata/phase7/`: inheritance diamonds and MRO, `super()` in cooperative multiple inheritance, properties, operator overloading both directions, container protocol classes, class attribute shadowing, user exceptions with chaining and the exact traceback text on stderr.
11. Write the unsupported cases and run the exit test.

### Phase 8: iterators, generators, context managers, and the rest of the syntax

> After this phase every statement and expression form of Python 3.13 parses and has an implementation. What is left is dynamism and the library.

1. Implement the iterator protocol for everything: `iter()`, `next()` with a default, `StopIteration`, `for` over any iterable including user classes, and real lazy objects for `range`, `map`, `filter`, `zip`, `enumerate`, `reversed`, `dict` views and `str` iteration. `print(map(...))` shows `<map object at 0x...>` like CPython, with the address unmatched.
2. Implement generators: `yield`, `yield from`, generator expressions with real laziness, `send`, `throw`, `close`, `return value` carried by `StopIteration.value`, and `GeneratorExit`. A generator without `send` becomes a Go range-over-func iterator. One that needs `send` or `throw` runs as a coroutine with a strict hand-off, which keeps a single thread of execution as in Python. Benchmark both against CPython before choosing the default.
3. Implement `with` for any number of items, with exact `__exit__` semantics (exception suppression, the exception argument triple, an exception raised in `__exit__`), and `del` for items and slices. `del` of names and attributes moves to phase 9, since a deleted name is unbound at run time.
4. Add `None` as a value: `Optional[T]` and `T | None` as types, `x is None`, `is`, `is not`, and truthiness of `None`. Functions without a `return` return `None`, which fixes the "falls off the end" panic.
5. Support the walrus operator `:=`, chained comparisons (`a < b < c`, evaluating the middle operand once), conditional expressions, star expressions (`a, *rest = xs`, `[*a, *b]`, `{**d}`, `f(*xs)`), and `del`/`global`/`nonlocal` edge cases.
6. Implement `match`: literal, capture, wildcard, value, sequence, mapping, class and or-patterns, `as`, guards, and the exact failure semantics.
7. Implement PEP 695 and PEP 696: generic functions and classes `def f[T](x: T)`, `class Box[T]`, bounds and constraints, `type X = ...` aliases and type parameter defaults. Generics compile to Go generics where the type is known and to the dynamic form from phase 9 otherwise.
8. Implement `except*`, `ExceptionGroup` and `BaseExceptionGroup` with `split`, `subgroup` and `derive`.
9. Complete the syntax: the full f-string grammar of PEP 701 (nested quotes, nested replacement fields in format specs, `=`), implicit concatenation, line continuations, positional-only parameters, `print` as an ordinary function, and loop `else` clauses. Bytes prefixes move to phase 10, `async`/`await` to phase 13, and decorators on classes to phase 9, which has the dynamic classes they need. The decoder turns every node type it does not know into an "unsupported" error with a line number, and `py2go check` counts them.
10. Add a corpus check: `py2go check` parses and generates every file of CPython's `Lib/`, and the number of rejected files is a tracked metric that must reach zero by phase 12.
11. Write 80+ differential tests under `testdata/phase8/` and unsupported cases, then run the exit test.

### Phase 9: the dynamic object model

> The pivot. CPython's `Lib/` has no type annotations, uses attributes added at run time, calls functions with whatever arguments arrive and builds classes with `type()`. To compile it, the runtime needs the universal Python object, and the type-hint requirement goes away. Annotated code keeps its fast typed path.

1. Record the decision. "Code without type hints is rejected" and "inference is a separate project" are reversed in `docs/decisions.md` and the README. The typed subset remains as an optimization: where every type is known the generated Go is the same as before.
2. Replace the dependency on a running CPython for parsing. Generate a Go parser from CPython's `Grammar/python.gram` (a PEG generator backend written in Go), checked against `ast.parse` on every file of `Lib/` and `Lib/test/`. The transpiler no longer shells out to `uv` and Python, and the same parser is used at run time by `eval`, `exec` and `compile`.
3. Define `runtime.Object`, the universal value: ints (small and big), floats, strs, bools, `None`, lists, tuples, dicts, sets, functions, bound methods, classes, instances and modules. Every object has a type object, a `__dict__` where Python gives one, and an identity. Classes get a real `type` with `__mro__`, `__bases__`, `__dict__` and metaclass support.
4. Compile unannotated code. A local type inference pass types what it can from literals, annotations, return types of known functions and flow, and gives everything else `Object`. Statically typed and dynamic values mix: boxing and unboxing are inserted at the boundaries.
5. Implement dynamic operations with CPython's rules and error messages: operators with reflected and in-place dispatch and `NotImplemented`, comparisons, truthiness, iteration, indexing and slicing, calls with any signature, attribute get, set and delete through the descriptor protocol and `__getattr__`, and `getattr`, `setattr`, `hasattr`, `delattr`, `vars`, `dir`, `id`, `hash`, `isinstance` with tuples and ABC hooks, `callable`.
6. Implement the type system's dynamic side: three-argument `type()`, metaclasses and `__prepare__`, `__init_subclass__`, `__set_name__`, `__class_getitem__`, `__instancecheck__`, `__subclasshook__`, `__slots__` with real slot semantics, `__new__` and `__init__` ordering.
7. Treat annotations as hints: `Any`, `Union`, `Literal` and friends are accepted, evaluated lazily where `from __future__ import annotations` applies, and stored in `__annotations__`.
8. Add `eval`, `exec` and `compile` with an AST interpreter over the same object model, with `globals()`, `locals()`, `__builtins__` and `__name__` handled properly. Phase 18 replaces the interpreter with a real compiler and virtual machine for bytecode-level features; until then it covers the uses in `Lib/` such as `namedtuple` and `dataclasses`.
9. Add real frames. `sys._getframe`, `sys.exc_info`, `sys.settrace` hooks at line level, `f_back`, `f_locals`, `f_lineno`, and `traceback` objects built from the `//line` mapping.
10. Emulate the observable parts of reference counting. Go's collector does not run `__del__` at the moment the last reference disappears. Objects of classes that define `__del__`, hold weak references, or are `with`-less file objects use explicit reference counting inferred by escape analysis, so `__del__` and weakref callbacks fire when CPython would run them in the common cases. Document the cases that differ.
11. Set a performance budget and measure it: annotated numeric code within 3x of hand-written Go, and dynamic code no slower than 2x CPython on a chosen subset of pyperformance. The numbers are tracked in `docs/benchmarks.md`.
12. The exit test is that every differential test from phases 2 to 8 passes again with all annotations stripped from its input, in addition to the original runs. A script removes annotations mechanically; the expected output does not change.

### Phase 10: numbers, text and bytes in full

> Phases 2 to 3 chose `int64`, a strict `bool`, and a subset of `str`. This phase removes those limits so that the data model matches CPython.

1. Make `int` arbitrary precision. Small values stay machine words and arithmetic switches to `big.Int` on overflow. Every "int64 wraps" limit in `docs/decisions.md` is removed, and `2 ** 100`, `10 ** 30` and `-9223372036854775808` work. Add `int.bit_length`, `bit_count`, `to_bytes`, `from_bytes`, `as_integer_ratio`, `is_integer`, `pow(a, b, m)` with negative exponents, `int(s, base)` for bases 2 to 36, `math.isqrt` support, and the exact limit `ValueError: Exceeds the limit (4300 digits)` for huge conversions.
2. Complete `float`: `hex`, `fromhex`, `as_integer_ratio`, `is_integer`, `round(x, n)` with CPython's correctly rounded algorithm, exact true division of big ints, and `float.__format__` for every type (`e`, `E`, `f`, `F`, `g`, `G`, `n`, `%`).
3. Add `complex`: literals, arithmetic, `abs`, `conjugate`, `repr`, and the cases phase 3 had to reject, such as a negative base with a fractional exponent.
4. Make `bool` a subclass of `int`: `True + 1`, `True == 1`, `sum([True, True])`, and `bool` keys that equal `1`. The strict comparison rules of phase 2 are removed.
5. Add `bytes`, `bytearray` and `memoryview` with all their methods, the buffer protocol, slicing and assignment, `%` formatting on bytes, and `hex`/`fromhex`.
6. Complete `str`: all the methods, `str.format` and `%` formatting with the full format-spec mini-language, `format()` and `__format__`, `translate`, `casefold`, `splitlines`, `partition`, `expandtabs`, `isidentifier`, `isprintable`, the full Unicode case mapping (`"ß".upper()` is `"SS"`), normalization via `unicodedata`, lone surrogates (strings are stored so that `"\ud800"` survives), and `repr` for every character class.
7. Add `str.encode` and `bytes.decode` with the core codecs (`utf-8`, `ascii`, `latin-1`, `utf-16`, `utf-32`, `unicode-escape`, `raw-unicode-escape`) and all error handlers (`strict`, `ignore`, `replace`, `backslashreplace`, `xmlcharrefreplace`, `namereplace`, `surrogateescape`, `surrogatepass`). The codec registry opens for the rest in phase 14.
8. Add `frozenset`, `slice`, `Ellipsis`, `NotImplemented`, the `range` object with `len`, indexing, slicing and `in`, and the number protocol for `hash()`: `hash(1) == hash(1.0) == hash(True)`, `hash` of `str` and `bytes` randomized as CPython does unless `PYTHONHASHSEED` is set, in which case the values match CPython's SipHash13 output.
9. Make `is` and `id` behave like CPython where programs depend on it: small integers from -5 to 256, interned string literals and identifier-like strings, `None`, `True`, `False`, the empty tuple.
10. Write 150+ differential tests under `testdata/phase10/`, including arithmetic on huge ints against CPython, float formatting sweeps like the ones used to check phase 3, codec round trips, and the bytes/str mixes that raise `TypeError`. Remove the `int64` limit entries from the README and rerun every earlier test.
11. Write the unsupported cases and run the exit test.

### Phase 11: modules, packages and the import system

> Every real program is several files, and the standard library is hundreds. This phase turns `py2go file.py` into a build driver for projects.

1. Compile `import x`, `import x.y as z`, `from x import a, b as c`, `from x import *` with `__all__`, relative imports, `if __name__ == "__main__":`, and the module attributes `__name__`, `__file__`, `__package__`, `__spec__`, `__doc__`, `__all__`, `__path__`.
2. Add a project mode. `py2go build <dir-or-file>` transpiles a package tree to one Go package per Python module and builds the binary. The import graph decides initialization order, and cycles behave as in CPython: a partially initialized module is visible in `sys.modules` while it runs. `py2go run` builds and runs.
3. Locate the standard library. The Python 3.13 `Lib/` tree is pinned through uv, and each module is compiled once and cached by content hash. A release build can ship the standard library as precompiled Go packages generated by `go generate`.
4. Implement `importlib` from `Lib/` on top of native `_frozen_importlib`, `_frozen_importlib_external` and `_imp`: finders, loaders, `sys.meta_path`, `sys.path_hooks`, `importlib.import_module`, `__import__`, `importlib.resources`, `importlib.metadata` for installed distributions, namespace packages, and `reload`.
5. Support imports that cannot be seen at build time (`importlib.import_module(name)` with a computed name, `exec` of source read from a file). Modules present at build time are compiled; anything else is loaded from source by the interpreter of phase 9 and runs with the same object model.
6. Make `__future__` real: `annotations`, `generator_stop` and the rest are accepted, and the ones that change semantics (`annotations`) do.
7. Add `pkgutil`, `runpy` (`python -m` style execution), `site` and `zipimport`, and a command line for generated programs that honours `sys.argv`, `-m`, `-c`, `-X` options and the `PYTHON*` environment variables that matter (`PYTHONHASHSEED`, `PYTHONPATH`, `PYTHONUTF8`, `PYTHONIOENCODING`, `PYTHONDONTWRITEBYTECODE`).
8. Write tests under `testdata/phase11/` as small multi-file projects: packages with relative imports, circular imports that work and ones that fail with the exact `ImportError` text, `__all__` filtering, a namespace package, an import inside a function, and `python -m` style entry points.
9. Write the unsupported cases and run the exit test.

### Phase 12: the standard library core, and CPython's test suite as the oracle

> From here on a module is done when CPython's own tests for it pass. This phase builds the minimum to run those tests, then the modules the language itself depends on.

1. Create `stdlib/manifest.json`, one entry per module of `sys.stdlib_module_names` plus `test` and `msilib`, with its phase, kind (L, N or P), status and the list of expected failures. A script checks the manifest against the roadmap map below and against the running CPython, so no module can be forgotten.
2. Write the native core in `runtime/`: `sys`, `builtins`, `gc`, `atexit`, `errno`, `_weakref`, `_warnings`, `_abc`, `_collections` (deque, defaultdict machinery), `_functools`, `_operator`, `itertools`, `math`, `marshal`, `_string`, `time`, `_signal`, `posix`, `_stat`, `_io` and `_locale`, `_sre`. Each matches CPython's behaviour and messages.
3. Port the regular expression engine. Go's `regexp` (RE2) lacks backreferences and lookaround, so `_sre` is a Go port of CPython's `_sre` and the `sre_parse` / `sre_compile` pair is compiled from `Lib/`. The `re` module passes `test_re`.
4. Compile the pure Python core from `Lib/`: `types`, `typing`, `abc`, `collections`, `functools`, `operator`, `copy`, `copyreg`, `enum`, `dataclasses`, `contextlib`, `numbers`, `keyword`, `reprlib`, `weakref`, `warnings`, `graphlib`, `string`, `textwrap`, `difflib`, `pprint`, `fnmatch`, `argparse`, `gettext`, `locale`, `shutil`, `os`, `posixpath`, `genericpath`, `stat`, `io`, `traceback`, `linecache`, `tokenize`, `token`, `platform`, `signal`, and `unittest` with `test.support`. Run `scripts/import_closure.py` to compute the transitive imports of `unittest` and `test.support` and move whatever else they need into this phase.
5. Burn down the corpus failures. The count of `Lib/` files that `py2go check` rejects is listed per run; every rejection is either fixed in the transpiler or recorded with a reason. The target at the end of this phase is zero rejections in the modules above.
6. Build the conformance harness. For each module it runs `python -m unittest test.test_<module>` under CPython 3.13 and the transpiled build, collects results by test id, and fails if any test passes in CPython and fails in the build and is not in the expected-failures file. The same harness runs `test_grammar`, `test_types`, `test_class`, `test_descr`, `test_exceptions`, `test_generators`, `test_int`, `test_float`, `test_str`, `test_list`, `test_dict`, `test_set` and the other language tests, which pins down phases 5 to 10 against the real suite.
7. Handle the places where `Lib/` uses `exec`, `sys._getframe` and `type()` to build code, namely `collections.namedtuple`, `dataclasses`, `enum` and `typing`. They run through the interpreter of phase 9 first, then get generated specializations where it matters for speed.
8. Verify startup: `import sys`, `site`, and the default `sys.path`, `sys.flags`, `sys.implementation`, `sys.version` (reporting 3.13), `sys.platform`, `sys.getrecursionlimit` and `setrecursionlimit`.
9. Run the exit test: all the tests of the modules in this phase pass except those in `expected_failures`, and the manifest check is green.

### Phase 13: concurrency and async

> Python code assumes a global interpreter lock, a single event loop thread, and `KeyboardInterrupt`. Go gives parallelism and none of those, so the semantics have to be rebuilt on top.

1. Emulate the GIL. Only one goroutine runs Python code at a time. The lock is released at blocking calls, I/O, `time.sleep`, and at a switch interval (`sys.setswitchinterval`, default 5 ms) checked at function calls and loop back-edges with a cheap flag. Operations CPython performs atomically under the GIL (`list.append`, `dict[k] = v`, `+=` on ints in one bytecode) stay atomic.
2. Implement `_thread`, `_threading_local`, `threading` (from `Lib/`), `queue`, `_queue`, locks, `RLock`, `Condition`, `Semaphore`, `Event`, `Barrier`, `Timer`, thread-local storage, daemon threads, `threading.excepthook`, and the exact behaviour of exceptions in threads.
3. Implement signals: `signal.signal` handlers run in the main thread between calls, `KeyboardInterrupt` on SIGINT, `signal.alarm`, `set_wakeup_fd`. This also gives `time.sleep` its interruptibility.
4. Compile `async def`, `await`, `async for`, `async with`, async generators and comprehensions, and coroutine objects with their `send`, `throw` and `close`. The runtime offers `_asyncio` natively (Future, Task, the event loop hooks), with `asyncio` compiled from `Lib/` on top of `select`, `selectors` (epoll, kqueue, poll) and `_socket`.
5. Implement `contextvars` and `_contextvars` natively, including context copying in tasks and callbacks.
6. Compile `concurrent.futures` (thread pool, process pool) and `sched`.
7. Implement `multiprocessing` and `_multiprocessing`, `_posixshmem`. Go cannot `fork` a running runtime safely, so the `spawn` and `forkserver` start methods re-execute the binary with a bootstrap argument and pickle state over pipes, and the `fork` start method is emulated with `spawn` and documented as the one deliberate difference.
8. Implement the experimental subinterpreter modules `_interpreters`, `_interpchannels`, `_interpqueues` at the level `Lib/test` expects.
9. Run the conformance harness for `test_threading`, `test_thread`, `test_queue`, `test_asyncio`, `test_contextvars`, `test_concurrent_futures`, `test_sched`, `test_signal`, `test_multiprocessing_*`, `test_generators` and `test_coroutines`, and the exit test.

### Phase 14: text, numeric and algorithm modules

> The libraries that processes numbers and text. They mix pure Python with C accelerators, and the accelerators decide whether the output matches digit for digit.

1. Match the C math library. `math` and `cmath` call the platform libm through cgo on Linux and macOS, because Go's `math` rounds some functions differently in the last bit and `repr` shows it. A pure Go fallback is available and documented as possibly differing. `math` keeps CPython's exceptions (`ValueError: math domain error`, `OverflowError`) and special-value handling.
2. Implement `_random` exactly: the Mersenne Twister with `init_by_array` seeding and `random.seed` of ints, strs and bytes, `getrandbits`, `random()`, and the `random` module from `Lib/`. Seeded sequences match CPython's.
3. Compile `decimal` from `_pydecimal` first, and write a native `_decimal` (a port of libmpdec) when speed requires it. Compile `fractions`, `statistics` (with `_statistics`), `numbers` follow-ups, and `_pylong`.
4. Implement `array`, `struct` and `_struct` natively with every format character, byte order and alignment rule, and `bisect` / `heapq` from `Lib/` with native accelerators.
5. Implement `unicodedata` with the Unicode 15.1 tables CPython 3.13 ships (generated from `Tools/unicode`), `stringprep`, and normalization.
6. Implement the codec machinery: `codecs`, `_codecs`, the CJK codecs (`_codecs_cn`, `_codecs_hk`, `_codecs_iso2022`, `_codecs_jp`, `_codecs_kr`, `_codecs_tw`, `_multibytecodec`) and the `encodings` package with all its tables.
7. Compile `glob`, `shlex`, `optparse`, `getopt`, `cmd`, `_markupbase`, `_strptime`, `colorsys`, `secrets`, `this` and `antigravity` from `Lib/`.
8. Run the conformance harness for `test_math`, `test_cmath`, `test_random`, `test_decimal`, `test_fractions`, `test_statistics`, `test_array`, `test_struct`, `test_bisect`, `test_heapq`, `test_unicodedata`, `test_codecs`, `test_codeccallbacks`, `test_multibytecodec`, `test_glob`, `test_shlex`, `test_optparse`, `test_getopt`, `test_cmd`, `test_strptime`, `test_secrets` and the rest of the phase, and the exit test.

### Phase 15: data formats, compression, persistence and hashing

> Modules that read and write bytes in a specified format. The outputs are compared byte for byte, so the native parts use the same C libraries as CPython.

1. Link the same C libraries as CPython through cgo for the parts where the exact bytes matter: zlib (`zlib`, `gzip`), bzip2 (`_bz2`), liblzma (`_lzma`), OpenSSL's digests (`_hashlib`), SQLite (`_sqlite3`), expat (`pyexpat`), and gdbm and ndbm (`_gdbm`, `_dbm`). Go's `compress/flate` produces different compressed bytes than zlib, so it is only the fallback for decompression.
2. Implement `_md5`, `_sha1`, `_sha2`, `_sha3`, `_blake2` and `hmac`, with the exact digest sizes, constructors and error behaviour of `hashlib`.
3. Implement `marshal` with the version 4 format CPython 3.13 writes, `_pickle` natively (protocols 0 to 5), and compile `pickle`, `pickletools`, `copyreg` follow-ups, `shelve`, `dbm` (with `dbm.dumb`).
4. Compile `json` with `_json` natively as an accelerator, `csv` with `_csv` (there is no pure Python fallback), `configparser`, `tomllib`, `plistlib`.
5. Compile `xml` (`dom`, `sax`, `etree`, `parsers`) with `pyexpat` and `_elementtree`, and `html` with `html.parser` and `html.entities`.
6. Compile `zipfile`, `tarfile`, `zipapp`, `base64`, `quopri`, `uuid` with `_uuid`, `mimetypes`, `netrc`, `wave`, `sqlite3` and its DB-API layer, and the `binascii` native module.
7. Run the conformance harness for `test_json`, `test_csv`, `test_configparser`, `test_tomllib`, `test_plistlib`, `test_xml_etree`, `test_minidom`, `test_sax`, `test_pyexpat`, `test_html`, `test_htmlparser`, `test_pickle`, `test_pickletools`, `test_marshal`, `test_shelve`, `test_dbm*`, `test_sqlite3`, `test_zlib`, `test_gzip`, `test_bz2`, `test_lzma`, `test_zipfile`, `test_tarfile`, `test_zipapp`, `test_base64`, `test_binascii`, `test_quopri`, `test_hashlib`, `test_hmac`, `test_uuid`, `test_mimetypes`, `test_netrc`, `test_wave`, and the exit test.

### Phase 16: operating system, files, time and processes

> Everything that touches the machine. The native parts are written against `syscall` and libc so that error numbers, flags and messages match CPython on Linux and macOS.

1. Implement `posix` natively: file descriptors, `open`, `read`, `write`, `stat`, `scandir`, `readlink`, environment, `fork`-free process control, `uname`, `urandom`, `getcwd`, `chdir`, and the full list of `os` functions, with `OSError` subclasses (`FileNotFoundError`, `PermissionError`, ...) chosen from `errno` exactly as CPython does. Compile `os`, `pathlib`, `shutil`, `tempfile`, `filecmp`, `fileinput`, `ntpath`, `nturl2path` from `Lib/`.
2. Implement `_io` natively and completely: `FileIO`, `BufferedReader`, `BufferedWriter`, `BufferedRandom`, `TextIOWrapper` (encodings, newline translation, line buffering, `errors`), `BytesIO`, `StringIO`, and `open()` with all modes. Keep `_pyio` from `Lib/` as the reference implementation tested against it.
3. Implement `time` completely (`strftime`, `strptime` support, `mktime`, `tzset`, `monotonic`, `perf_counter`, `process_time`, `clock_gettime`, `sleep` interruptible by signals), `datetime` with `_datetime` natively and `_pydatetime` as the reference, `zoneinfo` with `_zoneinfo` and the system tz database, `calendar`, `timeit`.
4. Implement `subprocess` with `_posixsubprocess.fork_exec` on `syscall.ForkExec` (with the documented limit that `preexec_fn` cannot run arbitrary Python in the child), `signal` follow-ups, `select` (`poll`, `epoll`, `kqueue`), `fcntl`, `termios`, `pty`, `tty`, `pwd`, `grp`, `resource`, `syslog`, `mmap`.
5. Compile `logging` (with `logging.config` and `logging.handlers`), `getpass`, `platform`, `sysconfig`, and implement `faulthandler`, which prints Python tracebacks on fatal signals using the frames from phase 9.
6. Compile the platform support modules (`_aix_support`, `_android_support`, `_apple_support`, `_ios_support`, `_osx_support`, `_scproxy`) so that each behaves as CPython does on its own platform and imports cleanly where it should.
7. Run the conformance harness for `test_os`, `test_posix`, `test_pathlib`, `test_shutil`, `test_tempfile`, `test_filecmp`, `test_fileinput`, `test_io`, `test_fileio`, `test_file`, `test_memoryio`, `test_bufio`, `test_time`, `test_datetime`, `test_zoneinfo`, `test_calendar`, `test_timeit`, `test_subprocess`, `test_signal`, `test_select`, `test_epoll`, `test_poll`, `test_fcntl`, `test_termios`, `test_pty`, `test_tty`, `test_pwd`, `test_grp`, `test_resource`, `test_syslog`, `test_mmap`, `test_logging`, `test_getpass`, `test_platform`, `test_sysconfig`, `test_faulthandler`, and the exit test.

### Phase 17: networking and internet protocols

> Sockets, TLS and the protocol modules built on them.

1. Implement `_socket` natively on raw `syscall` sockets, not on Go's `net`, so that the whole BSD socket API is available: every address family and option, `getaddrinfo`, `sendmsg` / `recvmsg` with ancillary data, `socketpair`, timeouts, and the exact `OSError` / `socket.timeout` / `gaierror` behaviour. Compile `socket`, `socketserver`, `select` users and `selectors` on top.
2. Implement `_ssl` against OpenSSL 3 through cgo, since Go's `crypto/tls` cannot reproduce `SSLContext`, certificate verification flags, `SSLObject`, session objects or the error messages. Compile `ssl` from `Lib/`.
3. Compile `http` (`client`, `server`, `cookies`, `cookiejar`), `urllib` (`request`, `parse`, `error`, `response`, `robotparser`), `email` (all of it, including `email.mime`, `email.policy`, `email.headerregistry`), `smtplib`, `ftplib`, `imaplib`, `poplib`, `xmlrpc`, `ipaddress`, `wsgiref`, `webbrowser` and `mailbox`.
4. Make the tests runnable offline: a loopback fixture replaces external hosts, and tests that need the network are in the expected-failures file with the reason.
5. Run the conformance harness for `test_socket`, `test_socketserver`, `test_ssl`, `test_httplib`, `test_httpservers`, `test_http_cookies`, `test_urllib*`, `test_email`, `test_smtplib`, `test_ftplib`, `test_imaplib`, `test_poplib`, `test_xmlrpc`, `test_ipaddress`, `test_wsgiref`, `test_webbrowser`, `test_mailbox`, and the exit test.

### Phase 18: introspection, the compiler and developer tools

> `dis`, `inspect`, `pdb`, `trace` and `py_compile` do not look at source text. They look at code objects and bytecode, which the runtime so far does not have. This phase adds them for real.

1. Port CPython 3.13's compiler to Go: the symbol table, the code generator, the flow graph optimizer and the assembler, so that `compile()` returns code objects holding genuine 3.13 bytecode, `co_*` attributes, line tables, exception tables and `co_consts`. The output is checked against CPython by comparing `marshal.dumps(compile(src, ...))` on all of `Lib/`.
2. Port CPython's evaluation loop as the bytecode virtual machine behind `exec`, `eval` and `compile`, replacing the AST interpreter from phase 9. Generated Go code and VM-run code share the object model, so they call each other freely. `sys.settrace`, `sys.setprofile`, `sys.monitoring` and frame `f_trace` work in VM-run code, and generated code hooks the same events at calls and line changes.
3. Implement `_ast` natively with the node classes and `ast` from `Lib/` on the parser of phase 9: `ast.parse`, `literal_eval`, `unparse`, `NodeVisitor`, `NodeTransformer`, `fix_missing_locations`, and compile from an AST.
4. Implement `_symtable`, `symtable`, `_opcode`, `_opcode_metadata`, `opcode`, `dis` (so `dis.dis(f)` prints CPython's listing), `_tokenize`, `tokenize`'s native tokenizer, `token`, `inspect` (signatures, source lookup, frames, `getmembers`, `getsource`), `code`, `codeop`, `traceback` extras.
5. Implement `bdb`, `pdb`, `trace`, `profile`, `cProfile` with `_lsprof`, `pstats`, `tracemalloc` with `_tracemalloc` (allocation tracking on top of the Go allocator hooks), `py_compile`, `compileall`, `modulefinder`, `pyclbr`, `tabnanny`.
6. Compile `unittest` extras (mock, async cases), `doctest`, `pydoc` with `pydoc_data`, `rlcompleter`, `readline` (cgo to libreadline or libedit), `_pyrepl`, `ensurepip`, `venv`, and the `test` package.
7. Run the conformance harness for `test_ast`, `test_compile`, `test_code`, `test_dis`, `test_inspect`, `test_symtable`, `test_tokenize`, `test_traceback`, `test_pdb`, `test_bdb`, `test_trace`, `test_profile`, `test_cprofile`, `test_pstats`, `test_tracemalloc`, `test_py_compile`, `test_compileall`, `test_modulefinder`, `test_pyclbr`, `test_tabnanny`, `test_unittest`, `test_doctest`, `test_pydoc`, `test_rlcompleter`, `test_readline`, `test_venv`, `test_ensurepip`, `test_sys_settrace`, `test_monitoring`, and the exit test.

### Phase 19: graphics, terminals and platform-specific modules

> The modules that need a display, a terminal library, a foreign function interface, or a particular operating system.

1. Implement `ctypes` and `_ctypes` with libffi through cgo: library loading, function prototypes, structures, unions, arrays, pointers, callbacks, `c_*` types, and the `errno` and `last_error` conventions.
2. Implement `curses`, `_curses` and `_curses_panel` against ncurses through cgo, with the exact module-level state rules.
3. Implement `_tkinter` against the system Tcl/Tk libraries through cgo, and compile `tkinter` (all submodules), `turtle`, `turtledemo` and `idlelib` from `Lib/`. The tests run under a virtual display.
4. Implement the Windows-only modules `msvcrt`, `winreg`, `winsound`, `_winapi`, `_overlapped`, `_wmi` and `nt` when the build target is Windows, using the Win32 API. On Linux and macOS they fail to import exactly as in CPython (`ModuleNotFoundError: No module named 'msvcrt'`), and `msilib`, which CPython 3.13 removed, fails to import everywhere.
5. Make Windows a real build target for the rest: path handling in `ntpath`, `os.name == "nt"`, console I/O, `subprocess` on `CreateProcess`, sockets on Winsock, signals as CPython maps them.
6. Run the conformance harness for `test_ctypes`, `test_curses`, `test_tkinter`, `test_ttk`, `test_turtle`, `test_idle`, `test_msvcrt`, `test_winreg`, `test_winsound`, `test_winapi`, and the exit test on each platform that applies.

### Phase 20: conformance, performance and release

> The exit criteria for the whole project.

1. The manifest has no entry in a state other than `done` or `platform`. Every module of `sys.stdlib_module_names` for Python 3.13, plus `test`, imports and works. The platform-specific ones behave on each platform as CPython does there.
2. `py2go check Lib/` accepts every file of CPython's `Lib/`, including `Lib/test/`, with zero rejections.
3. The conformance harness covers every `Lib/test/test_*.py`. Every test that passes under CPython 3.13 passes in the transpiled build, except those in `stdlib/expected_failures.json`. Each entry in that file has a reason from a short list (address-dependent output, refcount timing, thread scheduling, unavailable C library, platform), and the file is reviewed for every release.
4. The language conformance suites pass: `test_grammar`, `test_syntax`, `test_exceptions`, `test_descr`, `test_class`, `test_scope`, `test_generators`, `test_coroutines`, `test_with`, `test_unpack`, `test_fstring`, `test_patma`, `test_typing`, `test_type_params`, and the others in `Lib/test` that test the language rather than a module.
5. Differential fuzzing runs in CI: a generator of random programs over the supported language (the ones used during phases 3 and 4 grew into a standing tool) is run under CPython and the build, and any difference is a bug.
6. Performance is measured and published: startup time of a hello-world binary, the pyperformance suite compared to CPython, binary size, and transpile-plus-build time with and without the module cache. Regressions against the committed baseline fail CI.
7. Release the toolchain: a single `py2go` binary that embeds the parser and the standard library, builds for Linux, macOS and Windows on amd64 and arm64, reproducible builds, and a build flag for the cgo-free variant with its documented differences.
8. Write the documentation: a user guide, the data model and runtime internals, how to add a native module, and the final list of known differences from CPython (`id()`, address-dependent `repr`, hash randomization, thread timing, finalizer timing outside the emulated cases, C extensions).
9. Run the exit test: a clean checkout runs `go test -count=1 ./...` and the full conformance harness, and every check passes. Any valid Python 3.13 file that imports any standard library module transpiles, builds, and behaves like CPython, within the documented differences.

#### Standard library coverage map

Every module in `sys.stdlib_module_names` of Python 3.13, plus `test` and the removed `msilib`, appears exactly once. `scripts/check_stdlib_map.py` checks this list against the running CPython. A phase lists the modules that are first made to work in it. The transitive import closure computed in phase 12 can pull a module forward; when that happens the module moves to phase 12 in this list.

- Phase 11, from `Lib/`: importlib, zipimport, pkgutil, runpy, site, _sitebuiltins
- Phase 11, native Go: _frozen_importlib, _frozen_importlib_external, _imp
- Phase 12, from `Lib/`: types, typing, abc, _py_abc, collections, _collections_abc, functools, operator, copy, copyreg, enum, dataclasses, contextlib, numbers, keyword, reprlib, weakref, _weakrefset, warnings, graphlib, _colorize, __future__, re, sre_compile, sre_constants, sre_parse, string, textwrap, difflib, pprint, fnmatch, argparse, gettext, locale, shutil, os, posixpath, genericpath, stat, io, _pyio, traceback, linecache, tokenize, token, unittest, signal, _compression, platform
- Phase 12, native Go: builtins, sys, _typing, _abc, _collections, _functools, itertools, _operator, _weakref, gc, atexit, _warnings, _string, errno, marshal, math, _suggestions, _sre, _locale, posix, _stat, _io, time, _tokenize, _signal
- Phase 13, from `Lib/`: threading, _threading_local, queue, contextvars, asyncio, concurrent, multiprocessing, sched, selectors
- Phase 13, native Go: _thread, _queue, _contextvars, _asyncio, _multiprocessing, _posixshmem, _interpchannels, _interpqueues, _interpreters, select
- Phase 14, from `Lib/`: random, statistics, decimal, _pydecimal, _pylong, fractions, struct, bisect, heapq, stringprep, shlex, codecs, encodings, glob, optparse, getopt, cmd, _markupbase, _strptime, colorsys, secrets, this, antigravity
- Phase 14, native Go: cmath, _random, _statistics, _decimal, array, _struct, _bisect, _heapq, unicodedata, _codecs, _codecs_cn, _codecs_hk, _codecs_iso2022, _codecs_jp, _codecs_kr, _codecs_tw, _multibytecodec
- Phase 15, from `Lib/`: json, csv, configparser, tomllib, plistlib, xml, html, pickle, pickletools, _compat_pickle, shelve, dbm, sqlite3, gzip, bz2, lzma, zipfile, tarfile, zipapp, base64, quopri, hashlib, hmac, uuid, mimetypes, netrc, wave
- Phase 15, native Go: _json, _csv, pyexpat, _elementtree, _pickle, _dbm, _gdbm, _sqlite3, zlib, _bz2, _lzma, binascii, _hashlib, _md5, _sha1, _sha2, _sha3, _blake2, _uuid
- Phase 16, from `Lib/`: ntpath, nturl2path, pathlib, tempfile, filecmp, fileinput, subprocess, pty, tty, getpass, datetime, _pydatetime, zoneinfo, calendar, timeit, logging, sysconfig
- Phase 16, native Go: fcntl, pwd, grp, resource, mmap, _posixsubprocess, termios, syslog, faulthandler, _datetime, _zoneinfo, _sysconfig
- Phase 16, platform-specific: _aix_support, _android_support, _apple_support, _ios_support, _osx_support, _scproxy
- Phase 17, from `Lib/`: socket, socketserver, ssl, http, urllib, email, smtplib, ftplib, imaplib, poplib, xmlrpc, ipaddress, wsgiref, webbrowser, mailbox
- Phase 17, native Go: _socket, _ssl
- Phase 18, from `Lib/`: ast, symtable, opcode, _opcode_metadata, dis, inspect, tracemalloc, trace, pdb, bdb, code, codeop, profile, cProfile, pstats, doctest, pydoc, pydoc_data, py_compile, compileall, modulefinder, pyclbr, tabnanny, rlcompleter, ensurepip, venv, test, _pyrepl
- Phase 18, native Go: _ast, _symtable, _opcode, _tracemalloc, _lsprof, readline
- Phase 19, from `Lib/`: tkinter, turtle, turtledemo, idlelib, curses, ctypes
- Phase 19, native Go: _tkinter, _curses, _curses_panel, _ctypes
- Phase 19, platform-specific: msvcrt, winreg, winsound, msilib, _winapi, _overlapped, _wmi, nt
