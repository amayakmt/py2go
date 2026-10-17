# Known differences from CPython 3.13

These are deliberate or unavoidable, as of the last phase.

## Toolchain

- Transpiling needs `uv` and Python 3.13 (the parser is CPython's `ast`), and
  building needs the Go toolchain. The `py2go` binary does not embed a Python
  parser.
- Windows is not a build target: the runtime uses Linux and macOS system calls
  (sockets, terminals, files). Linux and macOS on amd64 and arm64 build.
- No cgo: no `ctypes`, `curses`, `tkinter`, `sqlite3`, `bz2`, `lzma`,
  `readline`. Importing the ones the linker knows raises `ModuleNotFoundError`
  like a CPython build without them. `pyexpat` is a parser written in Go, not
  expat: well-formed documents give expat's events, and errors carry expat's
  codes and messages, but the column of an error can differ.

## Language

- An annotation is a contract, not a hint: a value that does not match is a
  compile error. An annotated `int` is 64 bits and raises `OverflowError` on `*`,
  `**` and `<<` beyond that; unannotated `int`s are arbitrary precision.
- No `eval`, `exec`, `compile`, `ast.parse` or `__import__` of arbitrary names
  at run time: the program is compiled before it runs.
  `importlib.import_module` finds the modules the program links (any module
  it imports, and any named by a literal string in a call of
  `import_module`); other names raise `ModuleNotFoundError`. No code objects,
  bytecode, frames, `sys._getframe`, `sys.settrace`, `sys.setprofile`,
  `sys.monitoring`, `dis`, `inspect.getsource`, `pdb`, `doctest`.
- Metaclasses other than `ABCMeta`, `__set_name__`, `__slots__` and `__new__`
  (outside subclasses of `int`, `str`, `float`, `tuple`, `bytes` and
  `frozenset`) are limited or rejected. `__del__` runs when the collector frees
  the object, which is later than CPython's reference counting frees it, and at
  exit for what is left.
- A subclass of a built-in type (`dict`, `list`, `set`, `str`, `int`, `float`,
  `tuple`, `bytes`, `frozenset`) holds the built-in value instead of being one:
  `isinstance` and the built-in functions see through it, but a native function
  that requires the exact type receives the inner value.
- Function attributes are not supported. Decorators on methods other than the
  built-in ones are limited. `del name` and `del self.attr` clear the variable
  or attribute to its zero value: a later read gives that value where CPython
  raises `NameError` or `AttributeError`. `globals()` is a copy for reading;
  assigning into it does not create variables.
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
  auth or `ftp:`.
- `ssl` runs on Go's `crypto/tls`, not OpenSSL: the errors carry OpenSSL's
  reasons, verify codes and messages, but Go picks the cipher (the name is
  OpenSSL's spelling of Go's choice), cipher strings and security levels are
  not applied, and there is no memory BIO (`SSLObject`, `MemoryBIO`), so
  `asyncio` has no TLS; no session reuse, no post-handshake authentication.
- `tokenize` runs on a tokenizer written in Go that gives the C tokenizer's
  tokens; its `_tokenize.TokenizerIter` reads all the lines before the first
  token, and only with `extra_tokens=True`. No `pydoc`, so no
  `xmlrpc.server.DocXMLRPCServer`. `optparse`'s `Values.read_module` and
  `read_file` raise `NotImplementedError`. `cmath` can differ from CPython in
  the last digit, as `math` does.
- Wall-clock-dependent text (durations, timestamps) obviously differs; traceback
  text of uncaught errors in threads is shorter.

The exact list per module is in [decisions.md](decisions.md) under "Known limits"
and "Deviations".
