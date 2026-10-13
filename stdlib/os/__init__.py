"""os: the file system and the environment."""
from . import path
import builtins

name = "posix"
sep = "/"
altsep = None
extsep = "."
pathsep = ":"
linesep = "\n"
curdir = "."
pardir = ".."
devnull = "/dev/null"
environ = _native("os.environ")
error = OSError

F_OK = 0
R_OK = 4
W_OK = 2
X_OK = 1


def getcwd():
    return _native("os.getcwd")


def getenv(key, default=None):
    return _native("os.getenv", key, default)


def putenv(key, value):
    _native("os.putenv", key, value)
    environ[key] = value


def unsetenv(key):
    _native("os.unsetenv", key)


def getpid():
    return _native("os.getpid")


def urandom(size):
    return _native("secrets.token_bytes", size)


def getppid():
    return _native("os.getppid")


def kill(pid, sig):
    _native("signal.kill", pid, int(sig))


def _exit(status):
    _native("os._exit", status)


def cpu_count():
    return _native("os.cpu_count")


def listdir(path="."):
    return _native("os.listdir", path)


def mkdir(path, mode=511):
    _native("os.mkdir", path)


def makedirs(name, mode=511, exist_ok=False):
    _native("os.makedirs", name, mode, exist_ok)


def remove(path):
    _native("os.remove", path)


unlink = remove


def rmdir(path):
    _native("os.rmdir", path)


def rename(src, dst):
    _native("os.rename", src, dst)


def fspath(p):
    return p


def stat(path):
    return _native("os.stat", path)


def access(path, mode):
    return path.exists(path) if False else _native("os.path.exists", path)


def walk(top, topdown=True):
    names = listdir(top)
    dirs = []
    files = []
    for n in names:
        if path.isdir(path.join(top, n)):
            dirs.append(n)
        else:
            files.append(n)
    if topdown:
        yield (top, dirs, files)
    for d in dirs:
        for item in walk(path.join(top, d), topdown):
            yield item
    if not topdown:
        yield (top, dirs, files)
