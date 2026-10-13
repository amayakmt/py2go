"""contextlib: helpers for with statements."""


class _GeneratorContextManager:
    def __init__(self, func, args, kwds):
        self.gen = func(*args, **kwds)

    def __enter__(self):
        try:
            return next(self.gen)
        except StopIteration:
            raise RuntimeError("generator didn't yield") from None

    def __exit__(self, typ, value, traceback):
        if typ is None:
            try:
                next(self.gen)
            except StopIteration:
                return False
            raise RuntimeError("generator didn't stop")
        try:
            self.gen.throw(value)
        except StopIteration as exc:
            return exc is not value
        except BaseException as exc:
            if exc is value:
                return False
            raise
        raise RuntimeError("generator didn't stop after throw()")


def contextmanager(func):
    def helper(*args, **kwds):
        return _GeneratorContextManager(func, args, kwds)
    return helper


class closing:
    def __init__(self, thing):
        self.thing = thing

    def __enter__(self):
        return self.thing

    def __exit__(self, exc_type, exc, tb):
        self.thing.close()
        return False


class suppress:
    def __init__(self, *exceptions):
        self._exceptions = exceptions

    def __enter__(self):
        return None

    def __exit__(self, exctype, excinst, exctb):
        if exctype is None:
            return False
        return isinstance(excinst, self._exceptions)


class nullcontext:
    def __init__(self, enter_result=None):
        self.enter_result = enter_result

    def __enter__(self):
        return self.enter_result

    def __exit__(self, exc_type, exc, tb):
        return False


class redirect_stdout:
    def __init__(self, new_target):
        self._new_target = new_target
        self._old = None

    def __enter__(self):
        return self._new_target

    def __exit__(self, exctype, excinst, exctb):
        return False


class ExitStack:
    def __init__(self):
        self._callbacks = []

    def enter_context(self, cm):
        result = cm.__enter__()
        self._callbacks.append(cm)
        return result

    def callback(self, fn, *args, **kwds):
        self._callbacks.append((fn, args, kwds))
        return fn

    def push(self, exit_fn):
        self._callbacks.append(exit_fn)
        return exit_fn

    def pop_all(self):
        other = ExitStack()
        other._callbacks = self._callbacks
        self._callbacks = []
        return other

    def close(self):
        self.__exit__(None, None, None)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        while self._callbacks:
            cb = self._callbacks.pop()
            if isinstance(cb, tuple):
                cb[0](*cb[1], **cb[2])
            else:
                cb.__exit__(None, None, None)
        return False
