"""operator: the operators as functions."""
import builtins


def lt(a, b):
    return a < b


def le(a, b):
    return a <= b


def eq(a, b):
    return a == b


def ne(a, b):
    return a != b


def ge(a, b):
    return a >= b


def gt(a, b):
    return a > b


def not_(a):
    return not a


def truth(a):
    return True if a else False


def is_(a, b):
    return a is b


def is_not(a, b):
    return a is not b


def abs(a):
    return builtins.abs(a)


def add(a, b):
    return a + b


def and_(a, b):
    return a & b


def floordiv(a, b):
    return a // b


def index(a):
    return a.__index__()


def inv(a):
    return ~a


invert = inv


def lshift(a, b):
    return a << b


def mod(a, b):
    return a % b


def mul(a, b):
    return a * b


def neg(a):
    return -a


def or_(a, b):
    return a | b


def pos(a):
    return +a


def pow(a, b):
    return a ** b


def rshift(a, b):
    return a >> b


def sub(a, b):
    return a - b


def truediv(a, b):
    return a / b


def xor(a, b):
    return a ^ b


def concat(a, b):
    return a + b


def contains(a, b):
    return b in a


def countOf(a, b):
    n = 0
    for x in a:
        if x == b:
            n += 1
    return n


def indexOf(a, b):
    for i, x in enumerate(a):
        if x == b:
            return i
    raise ValueError("sequence.index(x): x not in sequence")


def getitem(a, b):
    return a[b]


def setitem(a, b, c):
    a[b] = c


def delitem(a, b):
    del a[b]


def itemgetter(*items):
    if len(items) == 1:
        item = items[0]
        return lambda obj: obj[item]
    return lambda obj: tuple(obj[i] for i in items)


def attrgetter(*attrs):
    def get(obj, name):
        for part in name.split("."):
            obj = getattr(obj, part)
        return obj
    if len(attrs) == 1:
        attr = attrs[0]
        return lambda obj: get(obj, attr)
    return lambda obj: tuple(get(obj, a) for a in attrs)


def methodcaller(name, /, *args, **kwargs):
    return lambda obj: getattr(obj, name)(*args, **kwargs)


def iadd(a, b):
    a += b
    return a


def isub(a, b):
    a -= b
    return a


def imul(a, b):
    a *= b
    return a


def itruediv(a, b):
    a /= b
    return a


def ifloordiv(a, b):
    a //= b
    return a


def imod(a, b):
    a %= b
    return a


def ipow(a, b):
    a **= b
    return a


def iand(a, b):
    a &= b
    return a


def ior(a, b):
    a |= b
    return a


def ixor(a, b):
    a ^= b
    return a


def ilshift(a, b):
    a <<= b
    return a


def irshift(a, b):
    a >>= b
    return a


def iconcat(a, b):
    a += b
    return a


def length_hint(obj, default=0):
    try:
        return len(obj)
    except TypeError:
        return default
