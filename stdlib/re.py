"""re: regular expressions. The engine is native."""

class RegexFlag:
    def __init__(self, value):
        self._value = value

    def __int__(self):
        return self._value

    def __index__(self):
        return self._value

    def __or__(self, other):
        return RegexFlag(self._value | _flagvalue(other))

    def __ror__(self, other):
        return RegexFlag(self._value | _flagvalue(other))

    def __and__(self, other):
        return RegexFlag(self._value & _flagvalue(other))

    def __xor__(self, other):
        return RegexFlag(self._value ^ _flagvalue(other))

    def __invert__(self):
        return RegexFlag(~self._value & 511)

    def __eq__(self, other):
        return self._value == _flagvalue(other)

    def __hash__(self):
        return hash(self._value)

    def __bool__(self):
        return self._value != 0

    def __repr__(self):
        if self._value == 0:
            return "re.NOFLAG"
        names = []
        for bit, name in _NAMES:
            if self._value & bit:
                names.append("re." + name)
        return "|".join(names)

    def __str__(self):
        return self.__repr__()


_NAMES = [(256, "ASCII"), (2, "IGNORECASE"), (4, "LOCALE"), (32, "UNICODE"), (8, "MULTILINE"), (16, "DOTALL"), (64, "VERBOSE"), (128, "DEBUG")]


def _flagvalue(flags):
    if isinstance(flags, RegexFlag):
        return flags._value
    return flags


NOFLAG = RegexFlag(0)
A = ASCII = RegexFlag(256)
I = IGNORECASE = RegexFlag(2)
L = LOCALE = RegexFlag(4)
M = MULTILINE = RegexFlag(8)
S = DOTALL = RegexFlag(16)
U = UNICODE = RegexFlag(32)
X = VERBOSE = RegexFlag(64)
DEBUG = RegexFlag(128)

from builtins import re_pm_PatternError as PatternError
from builtins import re_pm_PatternError as error


def compile(pattern, flags=0):
    return _native("re.compile", pattern, _flagvalue(flags))


def search(pattern, string, flags=0):
    return _native("re.compile", pattern, _flagvalue(flags)).search(string)


def match(pattern, string, flags=0):
    return _native("re.compile", pattern, _flagvalue(flags)).match(string)


def fullmatch(pattern, string, flags=0):
    return _native("re.compile", pattern, _flagvalue(flags)).fullmatch(string)


def split(pattern, string, maxsplit=0, flags=0):
    return _native("re.compile", pattern, _flagvalue(flags)).split(string, maxsplit)


def findall(pattern, string, flags=0):
    return _native("re.compile", pattern, _flagvalue(flags)).findall(string)


def finditer(pattern, string, flags=0):
    return _native("re.compile", pattern, _flagvalue(flags)).finditer(string)


def sub(pattern, repl, string, count=0, flags=0):
    return _native("re.compile", pattern, _flagvalue(flags)).sub(repl, string, count)


def subn(pattern, repl, string, count=0, flags=0):
    return _native("re.compile", pattern, _flagvalue(flags)).subn(repl, string, count)


def escape(pattern):
    return _native("re.escape", pattern)


def purge():
    _native("re.purge")
