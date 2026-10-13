"""collections: deque, defaultdict, Counter, OrderedDict, ChainMap.

They are classes that hold a list or a dict; none of them is a subclass of
list or dict, so isinstance(x, dict) is False for them.
"""
import builtins


class deque:
    def __init__(self, iterable=(), maxlen=None):
        if maxlen is not None and maxlen < 0:
            raise ValueError("maxlen must be non-negative")
        self.maxlen = maxlen
        self._items = []
        for x in iterable:
            self.append(x)

    def append(self, x):
        self._items.append(x)
        if self.maxlen is not None and len(self._items) > self.maxlen:
            del self._items[0]

    def appendleft(self, x):
        self._items.insert(0, x)
        if self.maxlen is not None and len(self._items) > self.maxlen:
            self._items.pop()

    def pop(self):
        if not self._items:
            raise IndexError("pop from an empty deque")
        return self._items.pop()

    def popleft(self):
        if not self._items:
            raise IndexError("pop from an empty deque")
        return self._items.pop(0)

    def extend(self, iterable):
        for x in list(iterable):
            self.append(x)

    def extendleft(self, iterable):
        for x in list(iterable):
            self.appendleft(x)

    def clear(self):
        self._items.clear()

    def copy(self):
        return deque(self._items, self.maxlen)

    def count(self, x):
        return self._items.count(x)

    def index(self, x, start=0, stop=None):
        if stop is None:
            stop = len(self._items)
        for i in range(start, min(stop, len(self._items))):
            if self._items[i] == x:
                return i
        raise ValueError(repr(x) + " is not in deque")

    def insert(self, i, x):
        if self.maxlen is not None and len(self._items) >= self.maxlen:
            raise IndexError("deque already at its maximum size")
        self._items.insert(i, x)

    def remove(self, x):
        for i in range(len(self._items)):
            if self._items[i] == x:
                del self._items[i]
                return
        raise ValueError(repr(x) + " is not in deque")

    def reverse(self):
        self._items.reverse()

    def rotate(self, n=1):
        size = len(self._items)
        if size == 0:
            return
        n = n % size
        if n:
            self._items = self._items[-n:] + self._items[:-n]

    def __len__(self):
        return len(self._items)

    def __iter__(self):
        return iter(list(self._items))

    def __reversed__(self):
        return iter(self._items[::-1])

    def __getitem__(self, i):
        if i >= len(self._items) or i < -len(self._items):
            raise IndexError("deque index out of range")
        return self._items[i]

    def __setitem__(self, i, x):
        if i >= len(self._items) or i < -len(self._items):
            raise IndexError("deque index out of range")
        self._items[i] = x

    def __delitem__(self, i):
        if i >= len(self._items) or i < -len(self._items):
            raise IndexError("deque index out of range")
        del self._items[i]

    def __contains__(self, x):
        return x in self._items

    def __bool__(self):
        return len(self._items) > 0

    def __eq__(self, other):
        return isinstance(other, deque) and self._items == other._items

    def __add__(self, other):
        if not isinstance(other, deque):
            raise TypeError("can only concatenate deque (not \"" + type(other).__name__ + "\") to deque")
        return deque(self._items + other._items, self.maxlen)

    def __mul__(self, n):
        return deque(self._items * n, self.maxlen)

    def __repr__(self):
        if self.maxlen is None:
            return "deque(" + repr(self._items) + ")"
        return "deque(" + repr(self._items) + ", maxlen=" + str(self.maxlen) + ")"


class _Mapping:
    """What defaultdict, OrderedDict and Counter share."""

    def __init__(self):
        self._d = {}

    def __len__(self):
        return len(self._d)

    def __iter__(self):
        return iter(list(self._d))

    def __contains__(self, key):
        return key in self._d

    def __setitem__(self, key, value):
        self._d[key] = value

    def __delitem__(self, key):
        del self._d[key]

    def __bool__(self):
        return len(self._d) > 0

    def __eq__(self, other):
        if isinstance(other, _Mapping):
            return self._d == other._d
        return self._d == other

    def keys(self):
        return self._d.keys()

    def values(self):
        return self._d.values()

    def items(self):
        return self._d.items()

    def get(self, key, default=None):
        return self._d.get(key, default)

    def pop(self, key, *default):
        return self._d.pop(key, *default)

    def popitem(self, last=True):
        if not self._d:
            raise KeyError("dictionary is empty")
        key = list(self._d)[-1 if last else 0]
        return (key, self._d.pop(key))

    def clear(self):
        self._d.clear()

    def setdefault(self, key, default=None):
        return self._d.setdefault(key, default)

    def update(self, other=(), **kwds):
        source = other._d if isinstance(other, _Mapping) else other
        self._d.update(source, **kwds)

    def __getitem__(self, key):
        return self._d[key]


class defaultdict(_Mapping):
    def __init__(self, default_factory=None, *args, **kwargs):
        self._d = {}
        self.default_factory = default_factory
        if args or kwargs:
            self.update(*args, **kwargs)

    def __getitem__(self, key):
        if key in self._d:
            return self._d[key]
        return self.__missing__(key)

    def __missing__(self, key):
        if self.default_factory is None:
            raise KeyError(key)
        value = self.default_factory()
        self._d[key] = value
        return value

    def copy(self):
        c = defaultdict(self.default_factory)
        c._d = dict(self._d)
        return c

    def __repr__(self):
        return "defaultdict(" + repr(self.default_factory) + ", " + repr(self._d) + ")"


class OrderedDict(_Mapping):
    def __init__(self, *args, **kwargs):
        self._d = {}
        if args or kwargs:
            self.update(*args, **kwargs)

    def move_to_end(self, key, last=True):
        value = self._d.pop(key)
        if last:
            self._d[key] = value
        else:
            rest = dict(self._d)
            self._d.clear()
            self._d[key] = value
            self._d.update(rest)

    def copy(self):
        return OrderedDict(self._d)

    def __eq__(self, other):
        if isinstance(other, OrderedDict):
            return list(self._d.items()) == list(other._d.items())
        if isinstance(other, _Mapping):
            return self._d == other._d
        return self._d == other

    def __repr__(self):
        if not self._d:
            return "OrderedDict()"
        return "OrderedDict(" + repr(self._d) + ")"


class Counter(_Mapping):
    def __init__(self, iterable=None, **kwds):
        self._d = {}
        self.update(iterable, **kwds)

    def __getitem__(self, key):
        return self._d.get(key, 0)

    def __missing__(self, key):
        return 0

    def update(self, iterable=None, **kwds):
        if iterable is not None:
            source = iterable._d if isinstance(iterable, _Mapping) else iterable
            if isinstance(source, dict):
                for key, n in source.items():
                    self._d[key] = self._d.get(key, 0) + n
            else:
                for key in source:
                    self._d[key] = self._d.get(key, 0) + 1
        for key, n in kwds.items():
            self._d[key] = self._d.get(key, 0) + n

    def subtract(self, iterable=None, **kwds):
        if iterable is not None:
            source = iterable._d if isinstance(iterable, _Mapping) else iterable
            if isinstance(source, dict):
                for key, n in source.items():
                    self._d[key] = self._d.get(key, 0) - n
            else:
                for key in source:
                    self._d[key] = self._d.get(key, 0) - 1
        for key, n in kwds.items():
            self._d[key] = self._d.get(key, 0) - n

    def most_common(self, n=None):
        items = sorted(self._d.items(), key=lambda kv: kv[1], reverse=True)
        if n is None:
            return items
        return items[:n]

    def elements(self):
        for key, n in self._d.items():
            for _ in range(n):
                yield key

    def total(self):
        return sum(self._d.values())

    def copy(self):
        return Counter(self._d)

    def __delitem__(self, key):
        if key in self._d:
            del self._d[key]

    def __repr__(self):
        if not self._d:
            return "Counter()"
        items = sorted(self._d.items(), key=lambda kv: kv[1], reverse=True)
        return "Counter({" + ", ".join(repr(k) + ": " + repr(v) for k, v in items) + "})"

    def __add__(self, other):
        result = Counter()
        for key, n in self._d.items():
            total = n + other[key]
            if total > 0:
                result._d[key] = total
        for key, n in other.items():
            if key not in self._d and n > 0:
                result._d[key] = n
        return result

    def __sub__(self, other):
        result = Counter()
        for key, n in self._d.items():
            total = n - other[key]
            if total > 0:
                result._d[key] = total
        for key, n in other.items():
            if key not in self._d and n < 0:
                result._d[key] = -n
        return result

    def __or__(self, other):
        result = Counter()
        for key, n in self._d.items():
            m = other[key]
            top = m if m > n else n
            if top > 0:
                result._d[key] = top
        for key, n in other.items():
            if key not in self._d and n > 0:
                result._d[key] = n
        return result

    def __and__(self, other):
        result = Counter()
        for key, n in self._d.items():
            m = other[key]
            low = n if n < m else m
            if low > 0:
                result._d[key] = low
        return result


class ChainMap(_Mapping):
    def __init__(self, *maps):
        self.maps = list(maps) or [{}]

    def __getitem__(self, key):
        for m in self.maps:
            if key in m:
                return m[key]
        raise KeyError(key)

    def get(self, key, default=None):
        if key in self:
            return self[key]
        return default

    def __contains__(self, key):
        for m in self.maps:
            if key in m:
                return True
        return False

    def __len__(self):
        return len(self._merged())

    def __iter__(self):
        return iter(self._merged())

    def _merged(self):
        merged = {}
        for m in reversed(self.maps):
            merged.update(m)
        return merged

    def keys(self):
        return self._merged().keys()

    def values(self):
        return self._merged().values()

    def items(self):
        return self._merged().items()

    def __setitem__(self, key, value):
        self.maps[0][key] = value

    def __delitem__(self, key):
        if key in self.maps[0]:
            del self.maps[0][key]
        else:
            raise KeyError("Key not found in the first mapping: " + repr(key))

    def new_child(self, m=None):
        if m is None:
            m = {}
        return ChainMap(m, *self.maps)

    @property
    def parents(self):
        return ChainMap(*self.maps[1:])

    def __repr__(self):
        return "ChainMap(" + ", ".join(repr(m) for m in self.maps) + ")"


def namedtuple(typename, field_names, *, rename=False, defaults=None, module=None):
    raise NotImplementedError("namedtuple needs literal names")
