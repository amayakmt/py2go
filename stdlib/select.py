"""select: wait for I/O completion (select() only)."""
error = OSError


def select(rlist, wlist, xlist, timeout=None):
    return _native("select.select", rlist, wlist, xlist, timeout)
