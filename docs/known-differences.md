# Known differences from CPython 3.13

These are deliberate or unavoidable, as of the last phase.

## Toolchain

- Transpiling needs `uv` and Python 3.13 (the parser is CPython's `ast`), and
  building needs the Go toolchain. The `py2go` binary does not embed a Python
  parser.
- Windows is not a build target: the runtime uses Linux and macOS system calls
  (sockets, terminals, files). Linux and macOS on amd64 and arm64 build.
- No cgo: no `ssl`, `ctypes`, `curses`, `tkinter`, `sqlite3`, `bz2`, `lzma`, `xml`
  (`pyexpat`), `readline`, `xmlrpc`. Importing the ones the linker knows
  raises `ModuleNotFoundError` like a CPython build without them.

## Language

- An annotation is a contract, not a hint: a value that does not match is a
  compile error. An annotated `int` is 64 bits and raises `OverflowError` on `*`,
  `**` and `<<` beyond that; unannotated `int`s are arbitrary precision.
- No `eval`, `exec`, `compile`, `ast.parse` or `__import__` of arbitrary names
  at run time: the program is compiled before it runs. No code objects,
  bytecode, frames, `sys._getframe`, `sys.settrace`, `sys.setprofile`,
  `sys.monitoring`, `dis`, `inspect.getsource`, `pdb`, `doctest`.
- Metaclasses, `__init_subclass__`, `__set_name__`, `__setattr__`-style attribute
  hooks beyond `__getattr__`/`__getattribute__`/properties, `__slots__` and
  `__new__` are limited or rejected; `__del__` does not run on garbage collection
  (files are flushed at exit; `open(...).write(...)` closes its file).
- An exception class held in a variable cannot be used in `except cls:`; use
  `except BaseException as e` and `isinstance`. Function attributes, docstrings
  at run time (`__doc__` of a function is `None`) and `globals()` / `locals()`
  are not supported. Decorators on methods other than the built-in ones are
  limited. `del name` of a local is rejected.
- Class-body names read inside the body, instance attributes shadowing methods
  after the compiler has bound a call statically, and containers whose item
  types change after first use are supported only in the dynamic mode.

## Library

- `id()` is the address of the Go object (or a hash for values); address-
  dependent `repr` differs. Hash randomization follows `PYTHONHASHSEED` but the
  hash values are not CPython's for every type.
- Threads run under a global lock handed over in ticket order; programs whose
  output depends on thread scheduling can differ.
- Socket constants are plain ints (CPython uses `IntEnum`), `HTTPStatus` and a
  few other enums are classes with members assigned after the class, and
  `collections.OrderedDict`, `defaultdict` and `Counter` do not subclass `dict`.
- `unittest` reports exception lines without stack traces; `unittest.main()` finds
  the test classes of the main file; discovery and module fixtures are missing.
  `unittest.mock.patch` works by `patch.object`/`patch.dict` only.
- `email` headers are `str` (no `headerregistry`); `urllib.request` has no proxy,
  auth or `ftp:`; `http.client` has no `HTTPSConnection`.
- Wall-clock-dependent text (durations, timestamps) obviously differs; traceback
  text of uncaught errors in threads is shorter.

The exact list per module is in [decisions.md](decisions.md) under "Known limits"
and "Deviations".
