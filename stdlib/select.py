"""select: wait for I/O completion (select() only)."""
error = OSError


def select(rlist, wlist, xlist, timeout=None):
    return _native("select.select", rlist, wlist, xlist, timeout)

POLLIN = 1
POLLPRI = 2
POLLOUT = 4
POLLERR = 8
POLLHUP = 16
POLLNVAL = 32
POLLRDNORM = 64
POLLRDBAND = 128
POLLWRNORM = 256
POLLWRBAND = 512
POLLRDHUP = 8192

EPOLLIN = 1
EPOLLPRI = 2
EPOLLOUT = 4
EPOLLERR = 8
EPOLLHUP = 16
EPOLLRDHUP = 8192
EPOLLET = 2147483648
EPOLLONESHOT = 1073741824
EPOLLEXCLUSIVE = 268435456
EPOLL_CLOEXEC = 524288
PIPE_BUF = 4096


def _fileno(f):
    if isinstance(f, int):
        return f
    return f.fileno()


class poll:
    def __init__(self):
        self._fds = {}

    def register(self, fd, eventmask=POLLIN | POLLPRI | POLLOUT):
        self._fds[_fileno(fd)] = eventmask

    def modify(self, fd, eventmask):
        n = _fileno(fd)
        if n not in self._fds:
            raise OSError(2, "No such file or directory")
        self._fds[n] = eventmask

    def unregister(self, fd):
        n = _fileno(fd)
        if n not in self._fds:
            raise KeyError(n)
        del self._fds[n]

    def poll(self, timeout=None):
        if timeout is not None:
            if timeout < 0:
                timeout = None
            else:
                timeout = timeout / 1000.0
        rl = [fd for fd in self._fds if self._fds[fd] & (POLLIN | POLLPRI)]
        wl = [fd for fd in self._fds if self._fds[fd] & POLLOUT]
        r, w, x = select(rl, wl, [], timeout)
        out = []
        for fd in self._fds:
            ev = 0
            if fd in r:
                ev |= POLLIN
            if fd in w:
                ev |= POLLOUT
            if ev:
                out.append((fd, ev))
        return out


class epoll:
    def __init__(self, sizehint=-1, flags=0):
        self._poll = poll()
        self._closed = False

    def close(self):
        self._closed = True

    @property
    def closed(self):
        return self._closed

    def fileno(self):
        return -1

    def register(self, fd, eventmask=EPOLLIN | EPOLLPRI | EPOLLOUT):
        n = _fileno(fd)
        if n in self._poll._fds:
            raise FileExistsError(17, "File exists")
        self._poll.register(n, eventmask)

    def modify(self, fd, eventmask):
        self._poll.modify(fd, eventmask)

    def unregister(self, fd):
        n = _fileno(fd)
        if n in self._poll._fds:
            del self._poll._fds[n]

    def poll(self, timeout=None, maxevents=-1):
        t = None if timeout is None or timeout < 0 else timeout * 1000.0
        return self._poll.poll(t)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        self.close()
