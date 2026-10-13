"""atexit: functions to call when the program ends."""


def register(func, *args, **kwargs):
    _native("atexit.register", func, *args, **kwargs)
    return func


def unregister(func):
    _native("atexit.unregister", func)


def _clear():
    _native("atexit._clear")


def _ncallbacks():
    return _native("atexit._ncallbacks")


def _run_exitfuncs():
    _native("atexit._run_exitfuncs")
