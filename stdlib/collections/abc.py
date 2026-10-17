"""Abstract Base Classes for containers.

The classes are ABCs: the built-in types are registered with them, and the
one-method ones (Iterable, Sized, Container, Hashable, Callable, ...) also
accept any object that has the method, as __subclasshook__ makes them do."""
from abc import ABC, abstractmethod

__all__ = ['Awaitable', 'Coroutine', 'AsyncIterable', 'AsyncIterator', 'AsyncGenerator', 'Hashable', 'Iterable', 'Iterator', 'Generator', 'Reversible', 'Sized', 'Container', 'Callable', 'Collection', 'Set', 'MutableSet', 'Mapping', 'MutableMapping', 'MappingView', 'KeysView', 'ItemsView', 'ValuesView', 'Sequence', 'MutableSequence', 'ByteString', 'Buffer']


class Hashable(ABC):

    @abstractmethod
    def __hash__(self):
        return 0


class Awaitable(ABC):

    @abstractmethod
    def __await__(self):
        ...


class Coroutine(Awaitable):

    @abstractmethod
    def send(self, value):
        raise StopIteration

    @abstractmethod
    def throw(self, typ, val=None, tb=None):
        raise typ

    def close(self):
        pass


class AsyncIterable(ABC):

    @abstractmethod
    def __aiter__(self):
        return self


class AsyncIterator(AsyncIterable):

    @abstractmethod
    def __anext__(self):
        raise StopAsyncIteration

    def __aiter__(self):
        return self


class AsyncGenerator(AsyncIterator):

    @abstractmethod
    def asend(self, value):
        raise StopAsyncIteration

    @abstractmethod
    def athrow(self, typ, val=None, tb=None):
        raise typ


class Iterable(ABC):

    @abstractmethod
    def __iter__(self):
        return iter([])


class Iterator(Iterable):

    @abstractmethod
    def __next__(self):
        raise StopIteration

    def __iter__(self):
        return self


class Reversible(Iterable):

    @abstractmethod
    def __reversed__(self):
        return iter([])


class Generator(Iterator):

    def __next__(self):
        return self.send(None)

    @abstractmethod
    def send(self, value):
        raise StopIteration

    @abstractmethod
    def throw(self, typ, val=None, tb=None):
        raise typ

    def close(self):
        pass


class Sized(ABC):

    @abstractmethod
    def __len__(self):
        return 0


class Container(ABC):

    @abstractmethod
    def __contains__(self, x):
        return False


class Collection(Sized, Iterable, Container):
    pass


class Buffer(ABC):

    @abstractmethod
    def __buffer__(self, flags):
        raise NotImplementedError


class Callable(ABC):

    @abstractmethod
    def __call__(self, *args, **kwds):
        return False


class Set(Collection):

    def __le__(self, other):
        if len(self) > len(other):
            return False
        for elem in self:
            if elem not in other:
                return False
        return True

    def __lt__(self, other):
        return len(self) < len(other) and self <= other

    def __ge__(self, other):
        if len(self) < len(other):
            return False
        for elem in other:
            if elem not in self:
                return False
        return True

    def __gt__(self, other):
        return len(self) > len(other) and self >= other

    def __eq__(self, other):
        return len(self) == len(other) and self <= other

    def isdisjoint(self, other):
        for value in other:
            if value in self:
                return False
        return True


class MutableSet(Set):

    @abstractmethod
    def add(self, value):
        raise NotImplementedError

    @abstractmethod
    def discard(self, value):
        raise NotImplementedError

    def remove(self, value):
        if value not in self:
            raise KeyError(value)
        self.discard(value)

    def pop(self):
        it = iter(self)
        try:
            value = next(it)
        except StopIteration:
            raise KeyError from None
        self.discard(value)
        return value

    def clear(self):
        try:
            while True:
                self.pop()
        except KeyError:
            pass


class Mapping(Collection):

    @abstractmethod
    def __getitem__(self, key):
        raise KeyError

    def get(self, key, default=None):
        try:
            return self[key]
        except KeyError:
            return default

    def __contains__(self, key):
        try:
            self[key]
        except KeyError:
            return False
        else:
            return True

    def keys(self):
        return list(iter(self))

    def items(self):
        return [(key, self[key]) for key in self]

    def values(self):
        return [self[key] for key in self]

    def __eq__(self, other):
        if not isinstance(other, Mapping):
            return NotImplemented
        return dict(self.items()) == dict(other.items())


class MutableMapping(Mapping):

    @abstractmethod
    def __setitem__(self, key, value):
        raise KeyError

    @abstractmethod
    def __delitem__(self, key):
        raise KeyError

    def pop(self, key, *args):
        try:
            value = self[key]
        except KeyError:
            if args:
                return args[0]
            raise
        else:
            del self[key]
            return value

    def popitem(self):
        try:
            key = next(iter(self))
        except StopIteration:
            raise KeyError from None
        value = self[key]
        del self[key]
        return key, value

    def clear(self):
        try:
            while True:
                self.popitem()
        except KeyError:
            pass

    def update(self, other=(), /, **kwds):
        if hasattr(other, 'keys'):
            for key in other.keys():
                self[key] = other[key]
        else:
            for key, value in other:
                self[key] = value
        for key in kwds:
            self[key] = kwds[key]

    def setdefault(self, key, default=None):
        try:
            return self[key]
        except KeyError:
            self[key] = default
        return default


class MappingView(Sized):

    def __init__(self, mapping):
        self._mapping = mapping

    def __len__(self):
        return len(self._mapping)


class KeysView(MappingView, Set):

    def __contains__(self, key):
        return key in self._mapping

    def __iter__(self):
        return iter(self._mapping)


class ItemsView(MappingView, Set):

    def __contains__(self, item):
        key, value = item
        try:
            v = self._mapping[key]
        except KeyError:
            return False
        else:
            return v is value or v == value

    def __iter__(self):
        return iter([(key, self._mapping[key]) for key in self._mapping])


class ValuesView(MappingView, Collection):

    def __contains__(self, value):
        for key in self._mapping:
            v = self._mapping[key]
            if v is value or v == value:
                return True
        return False

    def __iter__(self):
        return iter([self._mapping[key] for key in self._mapping])


class Sequence(Reversible, Collection):

    @abstractmethod
    def __getitem__(self, index):
        raise IndexError

    def __iter__(self):
        items = []
        i = 0
        try:
            while True:
                items.append(self[i])
                i += 1
        except IndexError:
            pass
        return iter(items)

    def __contains__(self, value):
        for v in self:
            if v is value or v == value:
                return True
        return False

    def __reversed__(self):
        return iter([self[i] for i in reversed(range(len(self)))])

    def index(self, value, start=0, stop=None):
        i = start
        while stop is None or i < stop:
            try:
                v = self[i]
            except IndexError:
                break
            if v is value or v == value:
                return i
            i += 1
        raise ValueError

    def count(self, value):
        return sum(1 for v in self if v is value or v == value)


class ByteString(Sequence):
    pass


class MutableSequence(Sequence):

    @abstractmethod
    def __setitem__(self, index, value):
        raise IndexError

    @abstractmethod
    def __delitem__(self, index):
        raise IndexError

    @abstractmethod
    def insert(self, index, value):
        raise IndexError

    def append(self, value):
        self.insert(len(self), value)

    def clear(self):
        try:
            while True:
                self.pop()
        except IndexError:
            pass

    def reverse(self):
        n = len(self)
        for i in range(n // 2):
            self[i], self[n - i - 1] = self[n - i - 1], self[i]

    def extend(self, values):
        for v in values:
            self.append(v)

    def pop(self, index=-1):
        v = self[index]
        del self[index]
        return v

    def remove(self, value):
        del self[self.index(value)]

    def __iadd__(self, values):
        self.extend(values)
        return self


# The structural checks of __subclasshook__: any object with these methods.
_native("abc.structural", Hashable, ("__hash__",))
_native("abc.structural", Awaitable, ("__await__",))
_native("abc.structural", AsyncIterable, ("__aiter__",))
_native("abc.structural", AsyncIterator, ("__aiter__", "__anext__"))
_native("abc.structural", Iterable, ("__iter__",))
_native("abc.structural", Iterator, ("__iter__", "__next__"))
_native("abc.structural", Reversible, ("__reversed__", "__iter__"))
_native("abc.structural", Generator, ("__iter__", "__next__", "send", "throw", "close"))
_native("abc.structural", Sized, ("__len__",))
_native("abc.structural", Container, ("__contains__",))
_native("abc.structural", Collection, ("__len__", "__iter__", "__contains__"))
_native("abc.structural", Callable, ("__call__",))

for _t in (tuple, str, range, memoryview, bytes):
    Sequence.register(_t)
for _t in (list, bytearray):
    MutableSequence.register(_t)
ByteString.register(bytes)
ByteString.register(bytearray)
Set.register(frozenset)
MutableSet.register(set)
MutableMapping.register(dict)
for _t in (int, float, complex, str, bytes, tuple, frozenset, bool):
    Hashable.register(_t)
