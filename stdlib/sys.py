"""sys: the parts of the interpreter a program can see."""
from collections import namedtuple

_VersionInfo = namedtuple("version_info", "major minor micro releaselevel serial")
_FloatInfo = namedtuple("float_info", "max max_exp max_10_exp min min_exp min_10_exp dig mant_dig epsilon radix rounds")
_IntInfo = namedtuple("int_info", "bits_per_digit sizeof_digit default_max_str_digits str_digits_check_threshold")
_Implementation = namedtuple("implementation", "name version hexversion cache_tag")
_Flags = namedtuple("flags", "debug inspect interactive optimize dont_write_bytecode no_user_site no_site ignore_environment verbose bytes_warning quiet hash_randomization isolated dev_mode utf8_mode warn_default_encoding safe_path int_max_str_digits")

argv = _native("sys.argv")
stdout = _native("sys.stdout")
stderr = _native("sys.stderr")
stdin = _native("sys.stdin")
__stdout__ = stdout
__stderr__ = stderr
__stdin__ = stdin
version_info = _VersionInfo(3, 13, 0, "final", 0)
version = "3.13.0 (py2go) [Go]"
hexversion = 0x030D00F0
api_version = 1013
platform = "linux"
byteorder = "little"
maxsize = 9223372036854775807
maxunicode = 1114111
executable = ""
prefix = ""
exec_prefix = ""
base_prefix = ""
base_exec_prefix = ""
path = []
class _Module:
    def __init__(self, name):
        self.name = name


modules = {'__main__': _Module('__main__')}
builtin_module_names = ("builtins", "sys")
float_info = _FloatInfo(1.7976931348623157e308, 1024, 308, 2.2250738585072014e-308, -1021, -307, 15, 53, 2.220446049250313e-16, 2, 1)
int_info = _IntInfo(30, 4, 4300, 640)
implementation = _Implementation("cpython", version_info, 0x030D00F0, "cpython-313")
flags = _Flags(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, False, 0, 0, False, 4300)
dont_write_bytecode = True
_recursion_limit = 1000


def exit(status=None):
    raise SystemExit(status)


def getrecursionlimit():
    return _recursion_limit


def setrecursionlimit(limit):
    global _recursion_limit
    if limit < 1:
        raise ValueError("recursion limit must be greater or equal than 1")
    _recursion_limit = limit


def getdefaultencoding():
    return "utf-8"


def getfilesystemencoding():
    return "utf-8"


def getrefcount(obj):
    return _native("sys.getrefcount", obj)


def intern(s):
    return s


def getsizeof(obj, default=None):
    return 64


def setswitchinterval(interval):
    _native("thread.setswitchinterval", interval)


def getswitchinterval():
    return _native("thread.getswitchinterval")


def exc_info():
    return _native("sys.exc_info")


def exception():
    return _native("sys.exc_info")[1]


def getfilesystemencodeerrors():
    return "surrogateescape"


def is_finalizing():
    return False


def audit(event, *args):
    return None


def addaudithook(hook):
    return None
