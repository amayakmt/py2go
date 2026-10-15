"""faulthandler: dump Python tracebacks on fatal errors (minimal)."""
_enabled = False


def enable(file=None, all_threads=True):
    global _enabled
    _enabled = True


def disable():
    global _enabled
    _enabled = False


def is_enabled():
    return _enabled


def dump_traceback(file=None, all_threads=True):
    pass


def dump_traceback_later(timeout, repeat=False, file=None, exit=False):
    pass


def cancel_dump_traceback_later():
    pass


def register(signum, file=None, all_threads=True, chain=False):
    pass


def unregister(signum):
    return False
