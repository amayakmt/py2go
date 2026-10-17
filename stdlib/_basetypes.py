"""The bases of classes that derive from dict, list and set.

`class C(dict)` is compiled as `class C(_PyDict)`: an object that holds a real
dict in `_data` and answers every method of dict through it, the way
collections.UserDict does. isinstance(c, dict) is true for it (the runtime
knows these three classes), and the methods behave as the C ones do for a
subclass: update() and the constructor do not call an overridden
__setitem__, __getitem__ calls __missing__, copy() gives a plain dict, the
operators give plain containers."""


def _obj(x: object) -> object:
    return x


_unset = object()


def _plain(x):
    """The real value behind x, or x itself."""
    if isinstance(x, (_PyDict, _PyList, _PySet, _PyStr, _PyInt, _PyFloat, _PyTuple, _PyBytes, _PyFrozenset)):
        return x._data
    return x


def _unhashable(self):
    raise TypeError("unhashable type: '%s'" % type(self).__name__)


class _PyDict:

    def __init__(self, *args, **kwargs):
        self._data = _obj({})
        if len(args) > 1:
            raise TypeError('dict expected at most 1 argument, got %d' % len(args))
        if args:
            self._fill(args[0])
        for k in kwargs:
            self._data[k] = kwargs[k]

    def _fill(self, other):
        other = _plain(other)
        if hasattr(other, 'keys'):
            for k in other.keys():
                self._data[k] = other[k]
        else:
            for pair in other:
                k, v = pair
                self._data[k] = v

    def __missing__(self, key):
        if key is not _unset:
            raise KeyError(key)
        return _obj(None)

    def __getitem__(self, key):
        if key in self._data:
            return self._data[key]
        return self.__missing__(key)

    def __setitem__(self, key, value):
        self._data[key] = value

    def __delitem__(self, key):
        del self._data[key]

    def __contains__(self, key):
        return key in self._data

    def __len__(self):
        return len(self._data)

    def __iter__(self):
        return iter(self._data)

    def __reversed__(self):
        return reversed(list(self._data))

    def __eq__(self, other):
        return self._data == _plain(other)

    def __ne__(self, other):
        return self._data != _plain(other)

    def __or__(self, other):
        new = dict(self._data)
        new.update(_plain(other))
        return new

    def __ior__(self, other):
        self._fill(other)
        return self

    def __repr__(self):
        return repr(self._data)

    def __str__(self):
        return self.__repr__()

    def __hash__(self):
        raise TypeError("unhashable type: '%s'" % type(self).__name__)

    def keys(self):
        return self._data.keys()

    def values(self):
        return self._data.values()

    def items(self):
        return self._data.items()

    def get(self, key, default=None):
        return self._data.get(key, default)

    def setdefault(self, key, default=None):
        return self._data.setdefault(key, default)

    def pop(self, key, *args):
        return self._data.pop(key, *args)

    def popitem(self):
        return self._data.popitem()

    def clear(self):
        self._data.clear()

    def update(self, *args, **kwargs):
        if len(args) > 1:
            raise TypeError('update expected at most 1 argument, got %d' % len(args))
        if args:
            self._fill(args[0])
        for k in kwargs:
            self._data[k] = kwargs[k]

    def copy(self):
        return dict(self._data)


class _PyList:

    def __init__(self, iterable=()):
        self._data = _obj(list(_plain(iterable)))

    def __getitem__(self, index):
        return self._data[index]

    def __setitem__(self, index, value):
        self._data[index] = value

    def __delitem__(self, index):
        del self._data[index]

    def __contains__(self, item):
        return item in self._data

    def __len__(self):
        return len(self._data)

    def __iter__(self):
        return iter(self._data)

    def __reversed__(self):
        return reversed(self._data)

    def __eq__(self, other):
        return self._data == _plain(other)

    def __ne__(self, other):
        return self._data != _plain(other)

    def __lt__(self, other):
        return self._data < _plain(other)

    def __le__(self, other):
        return self._data <= _plain(other)

    def __gt__(self, other):
        return self._data > _plain(other)

    def __ge__(self, other):
        return self._data >= _plain(other)

    def __add__(self, other):
        return self._data + list(_plain(other))

    def __radd__(self, other):
        return list(_plain(other)) + self._data

    def __iadd__(self, other):
        self._data.extend(_plain(other))
        return self

    def __mul__(self, n):
        return self._data * n

    def __rmul__(self, n):
        return self._data * n

    def __imul__(self, n):
        self._data[:] = self._data * n
        return self

    def __repr__(self):
        return repr(self._data)

    def __str__(self):
        return self.__repr__()

    def __hash__(self):
        raise TypeError("unhashable type: '%s'" % type(self).__name__)

    def append(self, item):
        self._data.append(item)

    def extend(self, other):
        self._data.extend(_plain(other))

    def insert(self, i, item):
        self._data.insert(i, item)

    def pop(self, i=-1):
        return self._data.pop(i)

    def remove(self, item):
        self._data.remove(item)

    def clear(self):
        self._data.clear()

    def index(self, item, *args):
        return self._data.index(item, *args)

    def count(self, item):
        return self._data.count(item)

    def reverse(self):
        self._data.reverse()

    def sort(self, *, key=None, reverse=False):
        self._data.sort(key=key, reverse=reverse)

    def copy(self):
        return list(self._data)


class _PySet:

    def __init__(self, iterable=()):
        self._data = _obj(set(_plain(iterable)))

    def __contains__(self, item):
        return item in self._data

    def __len__(self):
        return len(self._data)

    def __iter__(self):
        return iter(self._data)

    def __eq__(self, other):
        return self._data == _plain(other)

    def __ne__(self, other):
        return self._data != _plain(other)

    def __le__(self, other):
        return self._data <= _plain(other)

    def __lt__(self, other):
        return self._data < _plain(other)

    def __ge__(self, other):
        return self._data >= _plain(other)

    def __gt__(self, other):
        return self._data > _plain(other)

    def __or__(self, other):
        return self._data | set(_plain(other))

    def __and__(self, other):
        return self._data & set(_plain(other))

    def __sub__(self, other):
        return self._data - set(_plain(other))

    def __xor__(self, other):
        return self._data ^ set(_plain(other))

    def __ior__(self, other):
        self._data |= set(_plain(other))
        return self

    def __iand__(self, other):
        self._data &= set(_plain(other))
        return self

    def __isub__(self, other):
        self._data -= set(_plain(other))
        return self

    def __repr__(self):
        if not self._data:
            return '%s()' % type(self).__name__
        return '%s(%r)' % (type(self).__name__, self._data)

    def __str__(self):
        return self.__repr__()

    def __hash__(self):
        raise TypeError("unhashable type: '%s'" % type(self).__name__)

    def add(self, item):
        self._data.add(item)

    def discard(self, item):
        self._data.discard(item)

    def remove(self, item):
        self._data.remove(item)

    def pop(self):
        return self._data.pop()

    def clear(self):
        self._data.clear()

    def update(self, *others):
        for o in others:
            self._data.update(_plain(o))

    def union(self, *others):
        out = set(self._data)
        for o in others:
            out |= set(_plain(o))
        return out

    def intersection(self, *others):
        out = set(self._data)
        for o in others:
            out &= set(_plain(o))
        return out

    def difference(self, *others):
        out = set(self._data)
        for o in others:
            out -= set(_plain(o))
        return out

    def symmetric_difference(self, other):
        return self._data ^ set(_plain(other))

    def issubset(self, other):
        return self._data <= set(_plain(other))

    def issuperset(self, other):
        return self._data >= set(_plain(other))

    def isdisjoint(self, other):
        return self._data.isdisjoint(set(_plain(other)))

    def copy(self):
        return set(self._data)


class _PyStr:

    def __init__(self, value='', *args):
        if args:
            if len(args) == 1:
                self._data = _obj(str(_plain(value), args[0]))
            else:
                self._data = _obj(str(_plain(value), args[0], args[1]))
        else:
            self._data = _obj(str(_plain(value)))

    def _set(self, value=''):
        self._data = _obj(str(_plain(value)))
        return self

    def __str__(self):
        return self._data

    def __repr__(self):
        return repr(self._data)

    def __format__(self, spec):
        return format(self._data, spec)

    def __hash__(self):
        return hash(self._data)

    def __len__(self):
        return len(self._data)

    def __iter__(self):
        return iter(self._data)

    def __contains__(self, item):
        return _plain(item) in self._data

    def __getitem__(self, index):
        return self._data[index]

    def __add__(self, other):
        return self._data + _plain(other)

    def __radd__(self, other):
        return _plain(other) + self._data

    def __mul__(self, n):
        return self._data * n

    def __rmul__(self, n):
        return self._data * n

    def __mod__(self, args):
        return self._data % args

    def __getnewargs__(self):
        return (self._data,)

    def __eq__(self, other):
        return self._data == _plain(other)

    def __ne__(self, other):
        return self._data != _plain(other)

    def __lt__(self, other):
        return self._data < _plain(other)

    def __le__(self, other):
        return self._data <= _plain(other)

    def __gt__(self, other):
        return self._data > _plain(other)

    def __ge__(self, other):
        return self._data >= _plain(other)

    def capitalize(self, *args, **kwargs):
        return self._data.capitalize(*args, **kwargs)

    def casefold(self, *args, **kwargs):
        return self._data.casefold(*args, **kwargs)

    def center(self, *args, **kwargs):
        return self._data.center(*args, **kwargs)

    def count(self, *args, **kwargs):
        return self._data.count(*args, **kwargs)

    def encode(self, *args, **kwargs):
        return self._data.encode(*args, **kwargs)

    def endswith(self, *args, **kwargs):
        return self._data.endswith(*args, **kwargs)

    def expandtabs(self, *args, **kwargs):
        return self._data.expandtabs(*args, **kwargs)

    def find(self, *args, **kwargs):
        return self._data.find(*args, **kwargs)

    def format(self, *args, **kwargs):
        return self._data.format(*args, **kwargs)

    def format_map(self, *args, **kwargs):
        return self._data.format_map(*args, **kwargs)

    def index(self, *args, **kwargs):
        return self._data.index(*args, **kwargs)

    def isalnum(self, *args, **kwargs):
        return self._data.isalnum(*args, **kwargs)

    def isalpha(self, *args, **kwargs):
        return self._data.isalpha(*args, **kwargs)

    def isascii(self, *args, **kwargs):
        return self._data.isascii(*args, **kwargs)

    def isdecimal(self, *args, **kwargs):
        return self._data.isdecimal(*args, **kwargs)

    def isdigit(self, *args, **kwargs):
        return self._data.isdigit(*args, **kwargs)

    def isidentifier(self, *args, **kwargs):
        return self._data.isidentifier(*args, **kwargs)

    def islower(self, *args, **kwargs):
        return self._data.islower(*args, **kwargs)

    def isnumeric(self, *args, **kwargs):
        return self._data.isnumeric(*args, **kwargs)

    def isprintable(self, *args, **kwargs):
        return self._data.isprintable(*args, **kwargs)

    def isspace(self, *args, **kwargs):
        return self._data.isspace(*args, **kwargs)

    def istitle(self, *args, **kwargs):
        return self._data.istitle(*args, **kwargs)

    def isupper(self, *args, **kwargs):
        return self._data.isupper(*args, **kwargs)

    def join(self, *args, **kwargs):
        return self._data.join(*args, **kwargs)

    def ljust(self, *args, **kwargs):
        return self._data.ljust(*args, **kwargs)

    def lower(self, *args, **kwargs):
        return self._data.lower(*args, **kwargs)

    def lstrip(self, *args, **kwargs):
        return self._data.lstrip(*args, **kwargs)

    def partition(self, *args, **kwargs):
        return self._data.partition(*args, **kwargs)

    def removeprefix(self, *args, **kwargs):
        return self._data.removeprefix(*args, **kwargs)

    def removesuffix(self, *args, **kwargs):
        return self._data.removesuffix(*args, **kwargs)

    def replace(self, *args, **kwargs):
        return self._data.replace(*args, **kwargs)

    def rfind(self, *args, **kwargs):
        return self._data.rfind(*args, **kwargs)

    def rindex(self, *args, **kwargs):
        return self._data.rindex(*args, **kwargs)

    def rjust(self, *args, **kwargs):
        return self._data.rjust(*args, **kwargs)

    def rpartition(self, *args, **kwargs):
        return self._data.rpartition(*args, **kwargs)

    def rsplit(self, *args, **kwargs):
        return self._data.rsplit(*args, **kwargs)

    def rstrip(self, *args, **kwargs):
        return self._data.rstrip(*args, **kwargs)

    def split(self, *args, **kwargs):
        return self._data.split(*args, **kwargs)

    def splitlines(self, *args, **kwargs):
        return self._data.splitlines(*args, **kwargs)

    def startswith(self, *args, **kwargs):
        return self._data.startswith(*args, **kwargs)

    def strip(self, *args, **kwargs):
        return self._data.strip(*args, **kwargs)

    def swapcase(self, *args, **kwargs):
        return self._data.swapcase(*args, **kwargs)

    def title(self, *args, **kwargs):
        return self._data.title(*args, **kwargs)

    def translate(self, *args, **kwargs):
        return self._data.translate(*args, **kwargs)

    def upper(self, *args, **kwargs):
        return self._data.upper(*args, **kwargs)

    def zfill(self, *args, **kwargs):
        return self._data.zfill(*args, **kwargs)


class _PyBytes:

    def __init__(self, value=b'', *args):
        if args:
            if len(args) == 1:
                self._data = _obj(bytes(_plain(value), args[0]))
            else:
                self._data = _obj(bytes(_plain(value), args[0], args[1]))
        else:
            self._data = _obj(bytes(_plain(value)))

    def _set(self, value=b'', *args):
        if args:
            if len(args) == 1:
                self._data = _obj(bytes(_plain(value), args[0]))
            else:
                self._data = _obj(bytes(_plain(value), args[0], args[1]))
        else:
            self._data = _obj(bytes(_plain(value)))
        return self

    def __bytes__(self):
        return self._data

    def __repr__(self):
        return repr(self._data)

    def __str__(self):
        return self.__repr__()

    def __hash__(self):
        return hash(self._data)

    def __len__(self):
        return len(self._data)

    def __iter__(self):
        return iter(self._data)

    def __contains__(self, item):
        return _plain(item) in self._data

    def __getitem__(self, index):
        return self._data[index]

    def __add__(self, other):
        return self._data + _plain(other)

    def __radd__(self, other):
        return _plain(other) + self._data

    def __mul__(self, n):
        return self._data * n

    def __rmul__(self, n):
        return self._data * n

    def __mod__(self, args):
        return self._data % args

    def __eq__(self, other):
        return self._data == _plain(other)

    def __ne__(self, other):
        return self._data != _plain(other)

    def __lt__(self, other):
        return self._data < _plain(other)

    def __le__(self, other):
        return self._data <= _plain(other)

    def __gt__(self, other):
        return self._data > _plain(other)

    def __ge__(self, other):
        return self._data >= _plain(other)

    def capitalize(self, *args, **kwargs):
        return self._data.capitalize(*args, **kwargs)

    def center(self, *args, **kwargs):
        return self._data.center(*args, **kwargs)

    def count(self, *args, **kwargs):
        return self._data.count(*args, **kwargs)

    def decode(self, *args, **kwargs):
        return self._data.decode(*args, **kwargs)

    def endswith(self, *args, **kwargs):
        return self._data.endswith(*args, **kwargs)

    def expandtabs(self, *args, **kwargs):
        return self._data.expandtabs(*args, **kwargs)

    def find(self, *args, **kwargs):
        return self._data.find(*args, **kwargs)

    def hex(self, *args, **kwargs):
        return self._data.hex(*args, **kwargs)

    def index(self, *args, **kwargs):
        return self._data.index(*args, **kwargs)

    def isalnum(self, *args, **kwargs):
        return self._data.isalnum(*args, **kwargs)

    def isalpha(self, *args, **kwargs):
        return self._data.isalpha(*args, **kwargs)

    def isascii(self, *args, **kwargs):
        return self._data.isascii(*args, **kwargs)

    def isdigit(self, *args, **kwargs):
        return self._data.isdigit(*args, **kwargs)

    def islower(self, *args, **kwargs):
        return self._data.islower(*args, **kwargs)

    def isspace(self, *args, **kwargs):
        return self._data.isspace(*args, **kwargs)

    def istitle(self, *args, **kwargs):
        return self._data.istitle(*args, **kwargs)

    def isupper(self, *args, **kwargs):
        return self._data.isupper(*args, **kwargs)

    def join(self, *args, **kwargs):
        return self._data.join(*args, **kwargs)

    def ljust(self, *args, **kwargs):
        return self._data.ljust(*args, **kwargs)

    def lower(self, *args, **kwargs):
        return self._data.lower(*args, **kwargs)

    def lstrip(self, *args, **kwargs):
        return self._data.lstrip(*args, **kwargs)

    def partition(self, *args, **kwargs):
        return self._data.partition(*args, **kwargs)

    def removeprefix(self, *args, **kwargs):
        return self._data.removeprefix(*args, **kwargs)

    def removesuffix(self, *args, **kwargs):
        return self._data.removesuffix(*args, **kwargs)

    def replace(self, *args, **kwargs):
        return self._data.replace(*args, **kwargs)

    def rfind(self, *args, **kwargs):
        return self._data.rfind(*args, **kwargs)

    def rindex(self, *args, **kwargs):
        return self._data.rindex(*args, **kwargs)

    def rjust(self, *args, **kwargs):
        return self._data.rjust(*args, **kwargs)

    def rpartition(self, *args, **kwargs):
        return self._data.rpartition(*args, **kwargs)

    def rsplit(self, *args, **kwargs):
        return self._data.rsplit(*args, **kwargs)

    def rstrip(self, *args, **kwargs):
        return self._data.rstrip(*args, **kwargs)

    def split(self, *args, **kwargs):
        return self._data.split(*args, **kwargs)

    def splitlines(self, *args, **kwargs):
        return self._data.splitlines(*args, **kwargs)

    def startswith(self, *args, **kwargs):
        return self._data.startswith(*args, **kwargs)

    def strip(self, *args, **kwargs):
        return self._data.strip(*args, **kwargs)

    def swapcase(self, *args, **kwargs):
        return self._data.swapcase(*args, **kwargs)

    def title(self, *args, **kwargs):
        return self._data.title(*args, **kwargs)

    def translate(self, *args, **kwargs):
        return self._data.translate(*args, **kwargs)

    def upper(self, *args, **kwargs):
        return self._data.upper(*args, **kwargs)

    def zfill(self, *args, **kwargs):
        return self._data.zfill(*args, **kwargs)


class _PyInt:

    def __init__(self, value=0, *args):
        if args:
            if len(args) == 1:
                self._data = _obj(int(_plain(value), args[0]))
            else:
                raise TypeError('int() takes at most 2 arguments (%d given)' % (len(args) + 1))
        else:
            self._data = _obj(int(_plain(value)))

    def _set(self, value=0, *args):
        if args:
            if len(args) == 1:
                self._data = _obj(int(_plain(value), args[0]))
            else:
                raise TypeError('int() takes at most 2 arguments (%d given)' % (len(args) + 1))
        else:
            self._data = _obj(int(_plain(value)))
        return self

    def __repr__(self):
        return repr(self._data)

    def __str__(self):
        return self.__repr__()

    def __format__(self, spec):
        if spec == '':
            return self.__str__()
        return format(self._data, spec)

    def __hash__(self):
        return hash(self._data)

    def __bool__(self):
        return bool(self._data)

    def __int__(self):
        return int(self._data)

    def __float__(self):
        return float(self._data)

    def __neg__(self):
        return -self._data

    def __pos__(self):
        return +self._data

    def __abs__(self):
        return abs(self._data)

    def __round__(self, ndigits=None):
        if ndigits is None:
            return round(self._data)
        return round(self._data, ndigits)

    def __trunc__(self):
        return int(self._data)

    def __getnewargs__(self):
        return (self._data,)

    @property
    def real(self):
        return self._data.real

    @property
    def imag(self):
        return self._data.imag

    def conjugate(self):
        return self._data

    def __add__(self, other):
        return self._data + _plain(other)

    def __radd__(self, other):
        return _plain(other) + self._data

    def __sub__(self, other):
        return self._data - _plain(other)

    def __rsub__(self, other):
        return _plain(other) - self._data

    def __mul__(self, other):
        return self._data * _plain(other)

    def __rmul__(self, other):
        return _plain(other) * self._data

    def __truediv__(self, other):
        return self._data / _plain(other)

    def __rtruediv__(self, other):
        return _plain(other) / self._data

    def __floordiv__(self, other):
        return self._data // _plain(other)

    def __rfloordiv__(self, other):
        return _plain(other) // self._data

    def __mod__(self, other):
        return self._data % _plain(other)

    def __rmod__(self, other):
        return _plain(other) % self._data

    def __pow__(self, other):
        return self._data ** _plain(other)

    def __rpow__(self, other):
        return _plain(other) ** self._data

    def __divmod__(self, other):
        return divmod(self._data, _plain(other))

    def __rdivmod__(self, other):
        return divmod(_plain(other), self._data)

    def __lshift__(self, other):
        return self._data << _plain(other)

    def __rlshift__(self, other):
        return _plain(other) << self._data

    def __rshift__(self, other):
        return self._data >> _plain(other)

    def __rrshift__(self, other):
        return _plain(other) >> self._data

    def __and__(self, other):
        return self._data & _plain(other)

    def __rand__(self, other):
        return _plain(other) & self._data

    def __or__(self, other):
        return self._data | _plain(other)

    def __ror__(self, other):
        return _plain(other) | self._data

    def __xor__(self, other):
        return self._data ^ _plain(other)

    def __rxor__(self, other):
        return _plain(other) ^ self._data

    def __eq__(self, other):
        return self._data == _plain(other)

    def __ne__(self, other):
        return self._data != _plain(other)

    def __lt__(self, other):
        return self._data < _plain(other)

    def __le__(self, other):
        return self._data <= _plain(other)

    def __gt__(self, other):
        return self._data > _plain(other)

    def __ge__(self, other):
        return self._data >= _plain(other)

    def __index__(self):
        return self._data

    def __invert__(self):
        return ~self._data

    def __floor__(self):
        return self._data

    def __ceil__(self):
        return self._data

    @property
    def numerator(self):
        return self._data

    @property
    def denominator(self):
        return 1

    def bit_length(self):
        return self._data.bit_length()

    def bit_count(self):
        return self._data.bit_count()

    def to_bytes(self, *args, **kwargs):
        return self._data.to_bytes(*args, **kwargs)

    def as_integer_ratio(self):
        return (self._data, 1)

    def is_integer(self):
        return True


class _PyFloat:

    def __init__(self, value=0, *args):
        if args:
            raise TypeError('float expected at most 1 argument, got %d' % (len(args) + 1))
        else:
            self._data = _obj(float(_plain(value)))

    def _set(self, value=0, *args):
        if args:
            raise TypeError('float expected at most 1 argument, got %d' % (len(args) + 1))
        else:
            self._data = _obj(float(_plain(value)))
        return self

    def __repr__(self):
        return repr(self._data)

    def __str__(self):
        return self.__repr__()

    def __format__(self, spec):
        if spec == '':
            return self.__str__()
        return format(self._data, spec)

    def __hash__(self):
        return hash(self._data)

    def __bool__(self):
        return bool(self._data)

    def __int__(self):
        return int(self._data)

    def __float__(self):
        return float(self._data)

    def __neg__(self):
        return -self._data

    def __pos__(self):
        return +self._data

    def __abs__(self):
        return abs(self._data)

    def __round__(self, ndigits=None):
        if ndigits is None:
            return round(self._data)
        return round(self._data, ndigits)

    def __trunc__(self):
        return int(self._data)

    def __getnewargs__(self):
        return (self._data,)

    @property
    def real(self):
        return self._data.real

    @property
    def imag(self):
        return self._data.imag

    def conjugate(self):
        return self._data

    def __add__(self, other):
        return self._data + _plain(other)

    def __radd__(self, other):
        return _plain(other) + self._data

    def __sub__(self, other):
        return self._data - _plain(other)

    def __rsub__(self, other):
        return _plain(other) - self._data

    def __mul__(self, other):
        return self._data * _plain(other)

    def __rmul__(self, other):
        return _plain(other) * self._data

    def __truediv__(self, other):
        return self._data / _plain(other)

    def __rtruediv__(self, other):
        return _plain(other) / self._data

    def __floordiv__(self, other):
        return self._data // _plain(other)

    def __rfloordiv__(self, other):
        return _plain(other) // self._data

    def __mod__(self, other):
        return self._data % _plain(other)

    def __rmod__(self, other):
        return _plain(other) % self._data

    def __pow__(self, other):
        return self._data ** _plain(other)

    def __rpow__(self, other):
        return _plain(other) ** self._data

    def __divmod__(self, other):
        return divmod(self._data, _plain(other))

    def __rdivmod__(self, other):
        return divmod(_plain(other), self._data)

    def __eq__(self, other):
        return self._data == _plain(other)

    def __ne__(self, other):
        return self._data != _plain(other)

    def __lt__(self, other):
        return self._data < _plain(other)

    def __le__(self, other):
        return self._data <= _plain(other)

    def __gt__(self, other):
        return self._data > _plain(other)

    def __ge__(self, other):
        return self._data >= _plain(other)

    def is_integer(self):
        return self._data.is_integer()

    def as_integer_ratio(self):
        return self._data.as_integer_ratio()

    def hex(self):
        return self._data.hex()


class _PyTuple:

    def __init__(self, iterable=()):
        self._data = _obj(tuple(_plain(iterable)))

    def _set(self, iterable=()):
        self._data = _obj(tuple(_plain(iterable)))
        return self

    def __repr__(self):
        return repr(self._data)

    def __str__(self):
        return self.__repr__()

    def __hash__(self):
        return hash(self._data)

    def __len__(self):
        return len(self._data)

    def __iter__(self):
        return iter(self._data)

    def __reversed__(self):
        return reversed(self._data)

    def __contains__(self, item):
        return item in self._data

    def __getitem__(self, index):
        return self._data[index]

    def __add__(self, other):
        return self._data + tuple(_plain(other))

    def __radd__(self, other):
        return tuple(_plain(other)) + self._data

    def __mul__(self, n):
        return self._data * n

    def __rmul__(self, n):
        return self._data * n

    def __getnewargs__(self):
        return (self._data,)

    def count(self, item):
        return self._data.count(item)

    def index(self, item, *args):
        return self._data.index(item, *args)

    def __eq__(self, other):
        return self._data == _plain(other)

    def __ne__(self, other):
        return self._data != _plain(other)

    def __lt__(self, other):
        return self._data < _plain(other)

    def __le__(self, other):
        return self._data <= _plain(other)

    def __gt__(self, other):
        return self._data > _plain(other)

    def __ge__(self, other):
        return self._data >= _plain(other)


class _PyFrozenset:

    def __init__(self, iterable=()):
        self._data = _obj(frozenset(_plain(iterable)))

    def _set(self, iterable=()):
        self._data = _obj(frozenset(_plain(iterable)))
        return self

    def __repr__(self):
        if not self._data:
            return '%s()' % type(self).__name__
        return '%s(%r)' % (type(self).__name__, set(self._data))

    def __hash__(self):
        return hash(self._data)

    def __len__(self):
        return len(self._data)

    def __iter__(self):
        return iter(self._data)

    def __contains__(self, item):
        return item in self._data

    def __or__(self, other):
        return self._data | frozenset(_plain(other))

    def __and__(self, other):
        return self._data & frozenset(_plain(other))

    def __sub__(self, other):
        return self._data - frozenset(_plain(other))

    def __xor__(self, other):
        return self._data ^ frozenset(_plain(other))

    def union(self, *others):
        out = set(self._data)
        for o in others:
            out |= set(_plain(o))
        return frozenset(out)

    def intersection(self, *others):
        out = set(self._data)
        for o in others:
            out &= set(_plain(o))
        return frozenset(out)

    def difference(self, *others):
        out = set(self._data)
        for o in others:
            out -= set(_plain(o))
        return frozenset(out)

    def issubset(self, other):
        return self._data <= frozenset(_plain(other))

    def issuperset(self, other):
        return self._data >= frozenset(_plain(other))

    def isdisjoint(self, other):
        return self._data.isdisjoint(frozenset(_plain(other)))

    def copy(self):
        return self._data

    def __eq__(self, other):
        return self._data == _plain(other)

    def __ne__(self, other):
        return self._data != _plain(other)

    def __lt__(self, other):
        return self._data < _plain(other)

    def __le__(self, other):
        return self._data <= _plain(other)

    def __gt__(self, other):
        return self._data > _plain(other)

    def __ge__(self, other):
        return self._data >= _plain(other)
