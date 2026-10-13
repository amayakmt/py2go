"""_thread: the low-level thread interface."""

error = RuntimeError
TIMEOUT_MAX = 9223372036.0


def start_new_thread(function, args, kwargs=None):
    if kwargs is None:
        return _native("thread.start_new_thread", function, args)
    return _native("thread.start_new_thread", function, args, kwargs)


start_new = start_new_thread


def get_ident():
    return _native("thread.get_ident")


def get_native_id():
    return _native("thread.get_ident")


def allocate_lock():
    return _native("thread.allocate_lock")


allocate = allocate_lock


def RLock():
    return _native("thread.RLock")


def _count():
    return _native("thread.count")


def _local():
    return _native("thread.local")


def interrupt_main(signum=2):
    _native("thread.interrupt_main", signum)


def stack_size(size=0):
    if size != 0 and size < 32768:
        raise ValueError("size not valid: " + str(size) + " bytes")
    return 0


def daemon_threads_allowed():
    return True


def exit():
    raise SystemExit


exit_thread = exit
