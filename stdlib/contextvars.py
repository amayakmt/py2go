"""contextvars: context variables."""


class _Missing:
    def __repr__(self):
        return "<Token.MISSING>"


_MISSING = _Missing()


class _State:
    def __init__(self):
        self.contexts = {}


_state = _State()


def _ident():
    return _native("thread.get_ident")


def _current():
    ident = _ident()
    ctx = _state.contexts.get(ident)
    if ctx is None:
        ctx = Context()
        _state.contexts[ident] = ctx
    return ctx


class Context:
    def __init__(self):
        self._data = {}
        self._entered = False

    def run(self, callable, *args, **kwargs):
        if self._entered:
            raise RuntimeError("cannot enter context: " + repr(self) + " is already entered")
        ident = _ident()
        prev = _state.contexts.get(ident)
        self._entered = True
        _state.contexts[ident] = self
        try:
            return callable(*args, **kwargs)
        finally:
            self._entered = False
            if prev is None:
                _state.contexts.pop(ident, None)
            else:
                _state.contexts[ident] = prev

    def copy(self):
        new = Context()
        new._data = dict(self._data)
        return new

    def __getitem__(self, var):
        if var not in self._data:
            raise KeyError(var)
        return self._data[var]

    def get(self, var, default=None):
        if var in self._data:
            return self._data[var]
        return default

    def __contains__(self, var):
        return var in self._data

    def __len__(self):
        return len(self._data)

    def __iter__(self):
        return iter(self._data)

    def keys(self):
        return self._data.keys()

    def values(self):
        return self._data.values()

    def items(self):
        return self._data.items()

    def __repr__(self):
        return "<Context object at " + hex(id(self)) + ">"


def copy_context():
    return _current().copy()


class Token:
    MISSING = _MISSING

    def __init__(self, ctx, var, old):
        self._ctx = ctx
        self._var = var
        self._old = old
        self._used = False

    @property
    def var(self):
        return self._var

    @property
    def old_value(self):
        return self._old

    def __repr__(self):
        used = " used" if self._used else ""
        return "<Token" + used + " var=" + repr(self._var) + " at " + hex(id(self)) + ">"


class ContextVar:
    def __init__(self, name, *, default=_MISSING):
        if not isinstance(name, str):
            raise TypeError("context variable name must be a str")
        self._name = name
        self._default = default

    @property
    def name(self):
        return self._name

    def get(self, *args):
        ctx = _current()
        if self in ctx._data:
            return ctx._data[self]
        if len(args) > 1:
            raise TypeError("get expected at most 1 argument, got " + str(len(args)))
        if len(args) == 1:
            return args[0]
        if self._default is not _MISSING:
            return self._default
        raise LookupError(self)

    def set(self, value):
        ctx = _current()
        if self in ctx._data:
            old = ctx._data[self]
        else:
            old = _MISSING
        token = Token(ctx, self, old)
        ctx._data[self] = value
        return token

    def reset(self, token):
        if not isinstance(token, Token):
            raise TypeError("expected an instance of Token, got " + repr(token))
        if token._used:
            raise RuntimeError(repr(token) + " has already been used once")
        if token._var is not self:
            raise ValueError(repr(token) + " was created by a different ContextVar")
        ctx = _current()
        if token._ctx is not ctx:
            raise ValueError(repr(token) + " was created in a different Context")
        if token._old is _MISSING:
            ctx._data.pop(self, None)
        else:
            ctx._data[self] = token._old
        token._used = True

    def __repr__(self):
        d = ""
        if self._default is not _MISSING:
            d = " default=" + repr(self._default)
        return "<ContextVar name=" + repr(self._name) + d + " at " + hex(id(self)) + ">"
