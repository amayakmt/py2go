import errno
import grp
import os
import platform
import pwd
import stat
import timeit
import getpass
import warnings
import sys

print(errno.ENOENT, errno.errorcode[errno.EEXIST], errno.EACCES, errno.EPIPE, os.strerror(errno.ENOENT))
print(stat.S_ISDIR(0o40755), stat.S_ISREG(0o100644), stat.S_IMODE(0o100755), stat.filemode(0o100644), stat.filemode(0o40755), stat.filemode(0o120777), stat.S_IRWXU, stat.ST_MODE)
pw = pwd.getpwuid(os.getuid())
print(pw.pw_name == getpass.getuser() or True, pw.pw_uid == os.getuid(), isinstance(pw.pw_dir, str), pwd.getpwnam(pw.pw_name).pw_uid == pw[2], len(pw))
try:
    pwd.getpwnam("no-such-user-xyz")
except KeyError as e:
    print(e)
try:
    pwd.getpwuid(99999)
except KeyError as e:
    print(e)
g = grp.getgrgid(os.getgid())
print(g.gr_gid == os.getgid(), grp.getgrnam(g.gr_name).gr_gid == g.gr_gid)
try:
    grp.getgrnam("no-such-group-xyz")
except KeyError as e:
    print(e)
print(platform.system(), platform.machine() != "", platform.python_implementation(), platform.python_version_tuple()[0], platform.uname().system, len(platform.uname()))
print(platform.platform().startswith("Linux"), isinstance(platform.node(), str), platform.architecture()[0])
calls = []
def work():
    calls.append(1)
print(len(timeit.repeat(work, number=10, repeat=3)), len(calls), timeit.timeit(work, number=5) >= 0, len(calls))
t = timeit.Timer(work, setup=lambda: calls.append("s"))
print(t.timeit(2) >= 0, calls[-3:])
with warnings.catch_warnings(record=True) as caught:
    warnings.simplefilter("always")
    warnings.warn("careful", UserWarning)
    warnings.warn("old", DeprecationWarning)
print([(str(w.message), w.category.__name__) for w in caught])
warnings.warn("to stderr")
print(sys.exc_info())
try:
    raise KeyError("k")
except KeyError:
    et, ev, tb = sys.exc_info()
    print(et.__name__, ev)
    try:
        raise ValueError("inner")
    except ValueError:
        print(sys.exc_info()[0].__name__)
    print(sys.exc_info()[0].__name__)
print(sys.exc_info())
print(os.get_terminal_size.__name__, os.cpu_count() > 0, os.getloadavg()[0] >= 0, os.sep, os.pathsep, os.linesep == "\n", os.name)
print(os.O_RDONLY, os.O_CREAT, os.SEEK_END, os.EX_OK, os.WNOHANG)
print(os.environ.get("PATH") is not None, "HOME" in os.environ, isinstance(os.environ.copy(), dict), repr(os.environ)[:8])
print(os.system("exit 3") >> 8, os.system("true"))
print(os.popen("echo popen").read(), end="")
with os.popen("printf 'a\\nb\\n'") as p:
    print(p.readlines())
