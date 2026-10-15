"""fcntl: file descriptor control and file locking."""
LOCK_SH = 1
LOCK_EX = 2
LOCK_NB = 4
LOCK_UN = 8
F_DUPFD = 0
F_GETFD = 1
F_SETFD = 2
F_GETFL = 3
F_SETFL = 4
F_GETLK = 5
F_SETLK = 6
F_SETLKW = 7
F_RDLCK = 0
F_WRLCK = 1
F_UNLCK = 2
FD_CLOEXEC = 1
F_DUPFD_CLOEXEC = 1030


def _fd(fd):
    if isinstance(fd, int):
        return fd
    return fd.fileno()


def flock(fd, operation):
    _native("fcntl.flock", _fd(fd), operation)


def fcntl(fd, cmd, arg=0):
    return _native("fcntl.fcntl", _fd(fd), cmd, arg)


def lockf(fd, cmd, len=0, start=0, whence=0):
    if cmd == LOCK_UN:
        flock(fd, LOCK_UN)
    else:
        flock(fd, cmd)
