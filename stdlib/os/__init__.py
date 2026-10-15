"""os: operating system interfaces (POSIX)."""
import sys
import stat as st
from . import path
from .path import _fspath as fspath

name = "posix"
sep = "/"
altsep = None
extsep = "."
pathsep = ":"
linesep = "\n"
curdir = "."
pardir = ".."
devnull = "/dev/null"
defpath = "/bin:/usr/bin"
error = OSError

supports_bytes_environ = True

F_OK = 0
R_OK = 4
W_OK = 2
X_OK = 1

SEEK_SET = 0
SEEK_CUR = 1
SEEK_END = 2
SEEK_DATA = 3
SEEK_HOLE = 4

O_RDONLY = 0
O_WRONLY = 1
O_RDWR = 2
O_APPEND = 1024
O_CREAT = 64
O_EXCL = 128
O_TRUNC = 512
O_NONBLOCK = 2048
O_NDELAY = 2048
O_SYNC = 1052672
O_DSYNC = 4096
O_NOCTTY = 256
O_CLOEXEC = 524288
O_DIRECTORY = 65536
O_NOFOLLOW = 131072
O_ASYNC = 8192
O_DIRECT = 16384
O_LARGEFILE = 0
O_NOATIME = 262144
O_PATH = 2097152
O_TMPFILE = 4259840
O_ACCMODE = 3

EX_OK = 0
EX_USAGE = 64
EX_DATAERR = 65
EX_NOINPUT = 66
EX_NOUSER = 67
EX_NOHOST = 68
EX_UNAVAILABLE = 69
EX_SOFTWARE = 70
EX_OSERR = 71
EX_OSFILE = 72
EX_CANTCREAT = 73
EX_IOERR = 74
EX_TEMPFAIL = 75
EX_PROTOCOL = 76
EX_NOPERM = 77
EX_CONFIG = 78

WNOHANG = 1
WUNTRACED = 2
P_NOWAIT = 1
P_NOWAITO = 1
P_WAIT = 0
NGROUPS_MAX = 65536
TMP_MAX = 238328
PRIO_PROCESS = 0
PRIO_PGRP = 1
PRIO_USER = 2

supports_dir_fd = set()
supports_fd = {"stat", "truncate", "chmod", "chown", "listdir", "scandir"}
supports_effective_ids = {"access"}
supports_follow_symlinks = {"stat", "utime"}


class PathLike:
    def __fspath__(self):
        raise NotImplementedError


def fsencode(filename):
    filename = fspath(filename)
    if isinstance(filename, str):
        return filename.encode("utf-8", "surrogateescape")
    return filename


def fsdecode(filename):
    filename = fspath(filename)
    if isinstance(filename, bytes):
        return filename.decode("utf-8", "surrogateescape")
    return filename


def fsencoding():
    return "utf-8"


def get_exec_path(env=None):
    if env is None:
        e = environ
    else:
        e = env
    p = e.get("PATH", defpath)
    return p.split(pathsep)


# ---- the environment ----

class _Environ:
    def __init__(self, data):
        self._data = data

    def __getitem__(self, key):
        try:
            return self._data[key]
        except KeyError:
            raise KeyError(key) from None

    def __setitem__(self, key, value):
        if not isinstance(key, str):
            raise TypeError("str expected, not " + type(key).__name__)
        if not isinstance(value, str):
            raise TypeError("str expected, not " + type(value).__name__)
        if "=" in key or key == "":
            raise OSError(22, "Invalid argument")
        _native("os.putenv", key, value)
        self._data[key] = value

    def __delitem__(self, key):
        _native("os.unsetenv", key)
        try:
            del self._data[key]
        except KeyError:
            raise KeyError(key) from None

    def __iter__(self):
        return iter(list(self._data))

    def __len__(self):
        return len(self._data)

    def __contains__(self, key):
        return key in self._data

    def __repr__(self):
        return "environ(" + repr(self._data) + ")"

    def __eq__(self, other):
        if isinstance(other, _Environ):
            return self._data == other._data
        return self._data == other

    def get(self, key, default=None):
        return self._data.get(key, default)

    def keys(self):
        return self._data.keys()

    def values(self):
        return self._data.values()

    def items(self):
        return self._data.items()

    def copy(self):
        return dict(self._data)

    def setdefault(self, key, value):
        if key not in self._data:
            self[key] = value
        return self[key]

    def pop(self, key, *default):
        if key in self._data:
            v = self._data[key]
            del self[key]
            return v
        if default:
            return default[0]
        raise KeyError(key)

    def update(self, other=(), **kwds):
        if hasattr(other, "keys"):
            for k in other.keys():
                self[k] = other[k]
        else:
            for k, v in other:
                self[k] = v
        for k in kwds:
            self[k] = kwds[k]

    def clear(self):
        for k in list(self._data):
            del self[k]

    def popitem(self):
        for k in self._data:
            v = self._data[k]
            del self[k]
            return (k, v)
        raise KeyError("popitem(): dictionary is empty")

    def __or__(self, other):
        new = dict(self._data)
        new.update(other)
        return new


environ = _Environ(_native("os.environ"))


def getenv(key, default=None):
    return environ.get(key, default)


def putenv(key, value):
    _native("os.putenv", key, value)


def unsetenv(key):
    _native("os.unsetenv", key)


# ---- process information ----

def getpid():
    return _native("os.getpid")


def getppid():
    return _native("os.getppid")


def getuid():
    return _native("posix.ids")[0]


def getgid():
    return _native("posix.ids")[1]


def geteuid():
    return _native("posix.ids")[2]


def getegid():
    return _native("posix.ids")[3]


def getlogin():
    return _native("posix.getlogin")


def cpu_count():
    return _native("os.cpu_count")


def process_cpu_count():
    return _native("os.cpu_count")


def getloadavg():
    return _native("posix.getloadavg")


def uname():
    return uname_result(*_native("posix.uname"))


class uname_result:
    def __init__(self, sysname, nodename, release, version, machine):
        self.sysname = sysname
        self.nodename = nodename
        self.release = release
        self.version = version
        self.machine = machine

    def _t(self):
        return (self.sysname, self.nodename, self.release, self.version, self.machine)

    def __getitem__(self, i):
        return self._t()[i]

    def __iter__(self):
        return iter(self._t())

    def __len__(self):
        return 5

    def __eq__(self, other):
        if isinstance(other, uname_result):
            return self._t() == other._t()
        return self._t() == other

    def __repr__(self):
        return ("posix.uname_result(sysname=%r, nodename=%r, release=%r, version=%r, machine=%r)"
                % self._t())


class terminal_size:
    def __init__(self, columns, lines):
        self.columns = columns
        self.lines = lines

    def __getitem__(self, i):
        return (self.columns, self.lines)[i]

    def __iter__(self):
        return iter((self.columns, self.lines))

    def __len__(self):
        return 2

    def __eq__(self, other):
        if isinstance(other, terminal_size):
            return (self.columns, self.lines) == (other.columns, other.lines)
        return (self.columns, self.lines) == other

    def __repr__(self):
        return "os.terminal_size(columns=%d, lines=%d)" % (self.columns, self.lines)


def get_terminal_size(fd=1):
    c, l = _native("posix.get_terminal_size", fd)
    return terminal_size(c, l)


def umask(mask):
    return _native("posix.umask", mask)


def urandom(size):
    return _native("secrets.token_bytes", size)


def strerror(code):
    return _native("posix.strerror", code)


def kill(pid, sig):
    _native("signal.kill", pid, int(sig))


def _exit(status):
    _native("os._exit", status)


def abort():
    _native("os._exit", 134)


def system(command):
    return _native("posix.system", command)


def times():
    t = _native("time.process_time")
    return (t, 0.0, 0.0, 0.0, _native("time.perf_counter"))


def device_encoding(fd):
    if isatty(fd):
        return "utf-8"
    return None


# ---- file descriptors ----

def open(path, flags, mode=511, *, dir_fd=None):
    return _native("posix.open", fspath(path), flags, mode)


def close(fd):
    _native("posix.close", fd)


def closerange(fd_low, fd_high):
    for fd in range(fd_low, fd_high):
        try:
            _native("posix.close", fd)
        except OSError:
            pass


def read(fd, n):
    return _native("posix.read", fd, n)


def write(fd, data):
    return _native("posix.write", fd, data)


def lseek(fd, pos, how):
    return _native("posix.lseek", fd, pos, how)


def fstat(fd):
    return stat_result(_native("posix.fstat", fd))


def ftruncate(fd, length):
    _native("posix.ftruncate", fd, length)


def fsync(fd):
    _native("posix.fsync", fd)


def fdatasync(fd):
    _native("posix.fsync", fd)


def dup(fd):
    return _native("posix.dup", fd)


def dup2(fd, fd2, inheritable=True):
    return _native("posix.dup2", fd, fd2)


def pipe():
    return _native("posix.pipe")


def isatty(fd):
    return _native("posix.isatty", fd)


def fchmod(fd, mode):
    _native("posix.fchmod", fd, mode)


def get_inheritable(fd):
    return False


def set_inheritable(fd, inheritable):
    pass


def get_blocking(fd):
    return True


def set_blocking(fd, blocking):
    pass


def fdopen(fd, mode="r", buffering=-1, encoding=None, errors=None, newline=None, closefd=True):
    if not isinstance(fd, int):
        raise TypeError("invalid fd type (" + type(fd).__name__ + ", expected integer)")
    from builtins import open as _open
    return _open(fd, mode, buffering, encoding, errors, newline, closefd)


# ---- files and directories ----

class stat_result:
    def __init__(self, fields):
        self._f = fields
        self.st_mode = fields[0]
        self.st_ino = fields[1]
        self.st_dev = fields[2]
        self.st_nlink = fields[3]
        self.st_uid = fields[4]
        self.st_gid = fields[5]
        self.st_size = fields[6]
        self.st_atime = fields[10]
        self.st_mtime = fields[11]
        self.st_ctime = fields[12]
        self.st_atime_ns = fields[13]
        self.st_mtime_ns = fields[14]
        self.st_ctime_ns = fields[15]
        self.st_blksize = fields[16]
        self.st_blocks = fields[17]
        self.st_rdev = fields[18]

    def _t(self):
        f = self._f
        return (f[0], f[1], f[2], f[3], f[4], f[5], f[6], f[7], f[8], f[9])

    def __getitem__(self, i):
        return self._t()[i]

    def __iter__(self):
        return iter(self._t())

    def __len__(self):
        return 10

    def __eq__(self, other):
        if isinstance(other, stat_result):
            return self._t() == other._t()
        return self._t() == other

    def __hash__(self):
        return hash(self._t())

    def __repr__(self):
        f = self._f
        return ("os.stat_result(st_mode=%d, st_ino=%d, st_dev=%d, st_nlink=%d, st_uid=%d, st_gid=%d, "
                "st_size=%d, st_atime=%d, st_mtime=%d, st_ctime=%d)" % self._t())


def stat(path, *, dir_fd=None, follow_symlinks=True):
    if dir_fd is not None:
        raise NotImplementedError("dir_fd unavailable on this platform")
    if isinstance(path, int):
        return fstat(path)
    return stat_result(_native("posix.stat", fspath(path), follow_symlinks))


def lstat(path, *, dir_fd=None):
    return stat_result(_native("posix.stat", fspath(path), False))


def listdir(path=None):
    if path is None:
        path = "."
    return _native("posix.listdir", fspath(path))


class DirEntry:
    def __init__(self, dirpath, name, kind):
        self.name = name
        self.path = dirpath + name if dirpath.endswith("/") else dirpath + "/" + name
        if dirpath == "":
            self.path = name
        self._kind = kind
        self._stat = None
        self._lstat = None

    def __fspath__(self):
        return self.path

    def __repr__(self):
        return "<DirEntry " + repr(self.name) + ">"

    def inode(self):
        return self.stat(follow_symlinks=False).st_ino

    def is_symlink(self):
        if self._kind != "?":
            return self._kind == "l"
        return st.S_ISLNK(self.stat(follow_symlinks=False).st_mode)

    def is_dir(self, *, follow_symlinks=True):
        if self._kind == "d":
            return True
        if self._kind == "f":
            return False
        if self._kind == "l" and not follow_symlinks:
            return False
        try:
            return st.S_ISDIR(self.stat(follow_symlinks=follow_symlinks).st_mode)
        except OSError:
            return False

    def is_file(self, *, follow_symlinks=True):
        if self._kind == "f":
            return True
        if self._kind == "d":
            return False
        if self._kind == "l" and not follow_symlinks:
            return False
        try:
            return st.S_ISREG(self.stat(follow_symlinks=follow_symlinks).st_mode)
        except OSError:
            return False

    def is_junction(self):
        return False

    def stat(self, *, follow_symlinks=True):
        if follow_symlinks:
            if self._stat is None:
                if self._kind == "l" or self._kind == "?":
                    self._stat = stat(self.path)
                else:
                    self._stat = self.stat(follow_symlinks=False)
            return self._stat
        if self._lstat is None:
            self._lstat = lstat(self.path)
        return self._lstat


class _ScandirIterator:
    def __init__(self, entries):
        self._entries = entries
        self._i = 0
        self._closed = False

    def __iter__(self):
        return self

    def __next__(self):
        if self._closed or self._i >= len(self._entries):
            raise StopIteration
        e = self._entries[self._i]
        self._i += 1
        return e

    def close(self):
        self._closed = True

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        self.close()


def scandir(path=None):
    if path is None:
        p = "."
    else:
        p = fspath(path)
    raw = _native("posix.scandir", p)
    prefix = "" if path is None else p
    entries = []
    for item in raw:
        entries.append(DirEntry(prefix, item[0], item[1]))
    return _ScandirIterator(entries)


def mkdir(path, mode=511, *, dir_fd=None):
    _native("posix.mkdir", fspath(path), mode)


def makedirs(name, mode=511, exist_ok=False):
    name = fspath(name)
    head, tail = path.split(name)
    if not tail:
        head, tail = path.split(head)
    if head and tail and not path.exists(head):
        try:
            makedirs(head, exist_ok=exist_ok)
        except FileExistsError:
            pass
        cdir = curdir
        if tail == cdir:
            return
    try:
        mkdir(name, mode)
    except OSError:
        if not exist_ok or not path.isdir(name):
            raise


def remove(path, *, dir_fd=None):
    _native("posix.unlink", fspath(path))


unlink = remove


def rmdir(path, *, dir_fd=None):
    _native("posix.rmdir", fspath(path))


def removedirs(name):
    rmdir(name)
    head, tail = path.split(name)
    if not tail:
        head, tail = path.split(head)
    while head and tail:
        try:
            rmdir(head)
        except OSError:
            break
        head, tail = path.split(head)


def rename(src, dst, *, src_dir_fd=None, dst_dir_fd=None):
    _native("posix.rename", fspath(src), fspath(dst))


def replace(src, dst, *, src_dir_fd=None, dst_dir_fd=None):
    _native("posix.rename", fspath(src), fspath(dst))


def renames(old, new):
    head, tail = path.split(new)
    if head and tail and not path.exists(head):
        makedirs(head)
    rename(old, new)
    head, tail = path.split(old)
    if head and tail:
        try:
            removedirs(head)
        except OSError:
            pass


def link(src, dst, *, src_dir_fd=None, dst_dir_fd=None, follow_symlinks=True):
    _native("posix.link", fspath(src), fspath(dst))


def symlink(src, dst, target_is_directory=False, *, dir_fd=None):
    _native("posix.symlink", fspath(src), fspath(dst))


def readlink(path, *, dir_fd=None):
    return _native("posix.readlink", fspath(path))


def chmod(path, mode, *, dir_fd=None, follow_symlinks=True):
    if isinstance(path, int):
        _native("posix.fchmod", path, mode)
    else:
        _native("posix.chmod", fspath(path), mode)


def chown(path, uid, gid, *, dir_fd=None, follow_symlinks=True):
    if follow_symlinks:
        _native("posix.chown", fspath(path), uid, gid)
    else:
        _native("posix.lchown", fspath(path), uid, gid)


def lchown(path, uid, gid):
    _native("posix.lchown", fspath(path), uid, gid)


def utime(path, times=None, *, ns=None, dir_fd=None, follow_symlinks=True):
    if times is not None and ns is not None:
        raise ValueError("utime: you may specify either 'times' or 'ns' but not both")
    if ns is not None:
        a, m = ns
    elif times is not None:
        if not isinstance(times, tuple) or len(times) != 2:
            raise TypeError("utime: 'times' must be either a tuple of two ints or None")
        a = int(times[0] * 1000000000)
        m = int(times[1] * 1000000000)
    else:
        a = m = int(_native("time.time") * 1000000000)
    _native("posix.utime", fspath(path), a, m)


def truncate(path, length):
    if isinstance(path, int):
        _native("posix.ftruncate", path, length)
    else:
        _native("posix.truncate", fspath(path), length)


def access(path, mode, *, dir_fd=None, effective_ids=False, follow_symlinks=True):
    return _native("posix.access", fspath(path), mode)


def chdir(path):
    _native("posix.chdir", fspath(path))


def getcwd():
    return _native("posix.getcwd")


def getcwdb():
    return _native("posix.getcwd").encode("utf-8", "surrogateescape")


def mkfifo(path, mode=438, *, dir_fd=None):
    _native("posix.mkfifo", fspath(path), mode)


def statvfs(path):
    return _native("posix.statvfs", fspath(path))


def sync():
    pass


def walk(top, topdown=True, onerror=None, followlinks=False):
    top = fspath(top)
    dirs = []
    nondirs = []
    walk_dirs = []
    try:
        scandir_it = scandir(top)
    except OSError as error:
        if onerror is not None:
            onerror(error)
        return
    with scandir_it:
        while True:
            try:
                try:
                    entry = next(scandir_it)
                except StopIteration:
                    break
            except OSError as error:
                if onerror is not None:
                    onerror(error)
                return
            try:
                is_dir = entry.is_dir()
            except OSError:
                is_dir = False
            if is_dir:
                dirs.append(entry.name)
            else:
                nondirs.append(entry.name)
            if not topdown and is_dir:
                if followlinks:
                    walk_into = True
                else:
                    try:
                        is_symlink = entry.is_symlink()
                    except OSError:
                        is_symlink = False
                    walk_into = not is_symlink
                if walk_into:
                    walk_dirs.append(entry.path)
    if topdown:
        yield (top, dirs, nondirs)
        for dirname in dirs:
            new_path = path.join(top, dirname)
            if followlinks or not path.islink(new_path):
                for item in walk(new_path, topdown, onerror, followlinks):
                    yield item
    else:
        for new_path in walk_dirs:
            for item in walk(new_path, topdown, onerror, followlinks):
                yield item
        yield (top, dirs, nondirs)


def popen(cmd, mode="r", buffering=-1):
    if not isinstance(cmd, str):
        raise TypeError("invalid cmd type (%s, expected string)" % type(cmd))
    if mode not in ("r", "w"):
        raise ValueError("invalid mode %r" % mode)
    import subprocess
    if mode == "r":
        proc = subprocess.Popen(cmd, shell=True, text=True, stdout=subprocess.PIPE, bufsize=buffering)
        return _wrap_close(proc.stdout, proc)
    proc = subprocess.Popen(cmd, shell=True, text=True, stdin=subprocess.PIPE, bufsize=buffering)
    return _wrap_close(proc.stdin, proc)


class _wrap_close:
    def __init__(self, stream, proc):
        self._stream = stream
        self._proc = proc

    def close(self):
        self._stream.close()
        returncode = self._proc.wait()
        if returncode == 0:
            return None
        return returncode << 8

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        self.close()

    def __iter__(self):
        return iter(self._stream)

    def read(self, *args):
        return self._stream.read(*args)

    def readline(self, *args):
        return self._stream.readline(*args)

    def readlines(self, *args):
        return self._stream.readlines(*args)

    def write(self, s):
        return self._stream.write(s)

    def flush(self):
        self._stream.flush()
