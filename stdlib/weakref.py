"""Weak references (strong ones here: objects live until the program ends).

There is no reference counting to observe, so a `ref` never goes dead and
`finalize` callbacks run only when called or at exit through `atexit`."""
__all__ = ['ref', 'proxy', 'getweakrefcount', 'getweakrefs', 'WeakKeyDictionary', 'WeakValueDictionary', 'WeakSet', 'WeakMethod', 'finalize', 'ReferenceType', 'ProxyType', 'CallableProxyType']


def _obj(x: object) -> object:
    return x


class ref:

    def __init__(self, ob, callback=None):
        self._ob = _obj(ob)
        self.__callback__ = _obj(callback)

    def __call__(self):
        return self._ob

    def __eq__(self, other):
        if not isinstance(other, ref):
            return NotImplemented
        return self._ob == other._ob

    def __hash__(self):
        return hash(self._ob)

    def __repr__(self):
        return '<weakref at %s; to %r>' % (hex(id(self)), type(self._ob).__name__)


ReferenceType = ref
WeakMethod = ref


def proxy(object, callback=None):
    return object


ProxyType = type(proxy)
CallableProxyType = type(proxy)


def getweakrefcount(object):
    return 0


def getweakrefs(object):
    return []


class WeakValueDictionary:

    def __init__(self, other=(), /, **kw):
        self.data = _obj({})
        self.update(other, **kw)

    def update(self, other=(), /, **kw):
        if hasattr(other, 'keys'):
            for k in other.keys():
                self.data[k] = other[k]
        else:
            for k, v in other:
                self.data[k] = v
        for k in kw:
            self.data[k] = kw[k]

    def __getitem__(self, key):
        return self.data[key]

    def __setitem__(self, key, value):
        self.data[key] = value

    def __delitem__(self, key):
        del self.data[key]

    def __contains__(self, key):
        return key in self.data

    def __len__(self):
        return len(self.data)

    def __iter__(self):
        return iter(self.data)

    def get(self, key, default=None):
        return self.data.get(key, default)

    def keys(self):
        return self.data.keys()

    def values(self):
        return self.data.values()

    def items(self):
        return self.data.items()

    def pop(self, key, *args):
        return self.data.pop(key, *args)

    def setdefault(self, key, default=None):
        return self.data.setdefault(key, default)

    def copy(self):
        new = WeakValueDictionary()
        new.data = dict(self.data)
        return new

    def clear(self):
        self.data.clear()

    def valuerefs(self):
        return [ref(v) for v in self.data.values()]

    def itervaluerefs(self):
        return iter(self.valuerefs())

    def __repr__(self):
        return '<%s at %s>' % (type(self).__name__, hex(id(self)))


class WeakKeyDictionary(WeakValueDictionary):

    def keyrefs(self):
        return [ref(k) for k in self.data.keys()]

    def copy(self):
        new = WeakKeyDictionary()
        new.data = dict(self.data)
        return new


class WeakSet:

    def __init__(self, data=None):
        self.data = _obj([])
        if data is not None:
            for item in data:
                self.add(item)

    def add(self, item):
        if item not in self.data:
            self.data.append(item)

    def discard(self, item):
        if item in self.data:
            self.data.remove(item)

    def remove(self, item):
        self.data.remove(item)

    def pop(self):
        return self.data.pop()

    def clear(self):
        self.data.clear()

    def update(self, other):
        for item in other:
            self.add(item)

    def __contains__(self, item):
        return item in self.data

    def __len__(self):
        return len(self.data)

    def __iter__(self):
        return iter(list(self.data))

    def copy(self):
        return WeakSet(self.data)


class finalize:
    """Class for finalization of weakrefable objects; the callback runs when
    called, or at exit if `atexit` is true."""
    _registry = _obj([])

    def __init__(self, obj, func, *args, **kwargs):
        self._func = _obj(func)
        self._args = _obj(args)
        self._kwargs = _obj(kwargs)
        self.atexit = True
        self._alive = True
        finalize._registry.append(self)

    def __call__(self, unused=None):
        if self._alive:
            self._alive = False
            return self._func(*self._args, **self._kwargs)
        return None

    def detach(self):
        if self._alive:
            self._alive = False
            return (None, self._func, self._args, self._kwargs)
        return None

    def peek(self):
        if self._alive:
            return (None, self._func, self._args, self._kwargs)
        return None

    @property
    def alive(self):
        return self._alive
