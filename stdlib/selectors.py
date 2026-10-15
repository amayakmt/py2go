"""selectors: high-level I/O multiplexing (built on select)."""
import select
import sys

__all__ = ["EVENT_READ", "EVENT_WRITE", "SelectorKey", "BaseSelector", "SelectSelector", "PollSelector",
           "EpollSelector", "DefaultSelector"]

def _obj(x: object) -> object:
    return x


EVENT_READ = 1
EVENT_WRITE = 2


def _fileobj_to_fd(fileobj):
    if isinstance(fileobj, int):
        fd = fileobj
    else:
        try:
            fd = int(fileobj.fileno())
        except (AttributeError, TypeError, ValueError):
            raise ValueError("Invalid file object: " + repr(fileobj)) from None
    if fd < 0:
        raise ValueError("Invalid file descriptor: {}".format(fd))
    return fd


class SelectorKey:
    def __init__(self, fileobj, fd, events, data):
        self.fileobj = fileobj
        self.fd = fd
        self.events = events
        self.data = data

    def _t(self):
        return (self.fileobj, self.fd, self.events, self.data)

    def __getitem__(self, i):
        return self._t()[i]

    def __iter__(self):
        return iter(self._t())

    def __len__(self):
        return 4

    def __eq__(self, other):
        return self._t() == tuple(other)

    def __hash__(self):
        return hash(self._t())

    def __repr__(self):
        return "SelectorKey(fileobj=%r, fd=%r, events=%r, data=%r)" % self._t()


class _SelectorMapping:
    def __init__(self, selector):
        self._selector = selector

    def __len__(self):
        return len(self._selector._fd_to_key)

    def __getitem__(self, fileobj):
        try:
            fd = self._selector._fileobj_lookup(fileobj)
            return self._selector._fd_to_key[fd]
        except KeyError:
            raise KeyError("{!r} is not registered".format(fileobj)) from None

    def get(self, fileobj, default=None):
        try:
            return self[fileobj]
        except KeyError:
            return default

    def __iter__(self):
        return iter(self._selector._fd_to_key)

    def keys(self):
        return list(self._selector._fd_to_key)

    def values(self):
        return list(self._selector._fd_to_key.values())

    def items(self):
        return list(self._selector._fd_to_key.items())

    def __contains__(self, fileobj):
        return self.get(fileobj) is not None


class BaseSelector:
    def register(self, fileobj, events, data=None):
        raise NotImplementedError

    def unregister(self, fileobj):
        raise NotImplementedError

    def modify(self, fileobj, events, data=None):
        self.unregister(fileobj)
        return self.register(fileobj, events, data)

    def select(self, timeout=None):
        raise NotImplementedError

    def close(self):
        pass

    def get_key(self, fileobj):
        mapping = self.get_map()
        if mapping is None:
            raise RuntimeError("Selector is closed")
        try:
            return mapping[fileobj]
        except KeyError:
            raise KeyError("{!r} is not registered".format(fileobj)) from None

    def get_map(self):
        raise NotImplementedError

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        self.close()


class _BaseSelectorImpl(BaseSelector):
    def __init__(self):
        self._fd_to_key = {}
        self._map = _obj(_SelectorMapping(self))

    def _fileobj_lookup(self, fileobj):
        try:
            return _fileobj_to_fd(fileobj)
        except ValueError:
            for key in self._fd_to_key.values():
                if key.fileobj is fileobj:
                    return key.fd
            raise

    def register(self, fileobj, events, data=None):
        if (not events) or (events & ~(EVENT_READ | EVENT_WRITE)):
            raise ValueError("Invalid events: {!r}".format(events))
        key = SelectorKey(fileobj, self._fileobj_lookup(fileobj), events, data)
        if key.fd in self._fd_to_key:
            raise KeyError("{!r} (FD {}) is already registered".format(fileobj, key.fd))
        self._fd_to_key[key.fd] = key
        return key

    def unregister(self, fileobj):
        try:
            key = self._fd_to_key.pop(self._fileobj_lookup(fileobj))
        except KeyError:
            raise KeyError("{!r} is not registered".format(fileobj)) from None
        return key

    def modify(self, fileobj, events, data=None):
        try:
            key = self._fd_to_key[self._fileobj_lookup(fileobj)]
        except KeyError:
            raise KeyError("{!r} is not registered".format(fileobj)) from None
        if events != key.events:
            self.unregister(fileobj)
            key = self.register(fileobj, events, data)
        elif data != key.data:
            key = SelectorKey(key.fileobj, key.fd, key.events, data)
            self._fd_to_key[key.fd] = key
        return key

    def close(self):
        self._fd_to_key.clear()
        self._map = _obj(None)

    def get_map(self):
        return self._map

    def _key_from_fd(self, fd):
        try:
            return self._fd_to_key[fd]
        except KeyError:
            return None


class SelectSelector(_BaseSelectorImpl):
    def __init__(self):
        _BaseSelectorImpl.__init__(self)
        self._readers = set()
        self._writers = set()

    def register(self, fileobj, events, data=None):
        key = _BaseSelectorImpl.register(self, fileobj, events, data)
        if events & EVENT_READ:
            self._readers.add(key.fd)
        if events & EVENT_WRITE:
            self._writers.add(key.fd)
        return key

    def unregister(self, fileobj):
        key = _BaseSelectorImpl.unregister(self, fileobj)
        self._readers.discard(key.fd)
        self._writers.discard(key.fd)
        return key

    def select(self, timeout=None):
        timeout = None if timeout is None else max(timeout, 0)
        ready = []
        r, w, _ = select.select(list(self._readers), list(self._writers), [], timeout)
        r = set(r)
        w = set(w)
        for fd in r | w:
            events = 0
            if fd in r:
                events |= EVENT_READ
            if fd in w:
                events |= EVENT_WRITE
            key = self._key_from_fd(fd)
            if key:
                ready.append((key, events & key.events))
        return ready


class PollSelector(SelectSelector):
    pass


class EpollSelector(SelectSelector):
    pass


class DefaultSelector(SelectSelector):
    pass
