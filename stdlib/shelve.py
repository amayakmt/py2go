"""shelve: manage shelves of pickled objects (dbm.dumb based)."""
import io
import pickle
import dbm

__all__ = ["Shelf", "BsdDbShelf", "DbfilenameShelf", "open"]


class Shelf:
    def __init__(self, dict, protocol=None, writeback=False, keyencoding="utf-8"):
        self.dict = dict
        if protocol is None:
            protocol = 5
        self._protocol = protocol
        self.writeback = writeback
        self.cache = {}
        self.keyencoding = keyencoding

    def keys(self):
        return [k.decode(self.keyencoding) for k in self.dict.keys()]

    def __len__(self):
        return len(self.dict)

    def __contains__(self, key):
        return key.encode(self.keyencoding) in self.dict

    def get(self, key, default=None):
        if key.encode(self.keyencoding) in self.dict:
            return self[key]
        return default

    def __getitem__(self, key):
        try:
            value = self.cache[key]
        except KeyError:
            f = io.BytesIO(self.dict[key.encode(self.keyencoding)])
            value = pickle.load(f)
            if self.writeback:
                self.cache[key] = value
        return value

    def __setitem__(self, key, value):
        if self.writeback:
            self.cache[key] = value
        f = io.BytesIO()
        p = pickle.Pickler(f, self._protocol)
        p.dump(value)
        self.dict[key.encode(self.keyencoding)] = f.getvalue()

    def __delitem__(self, key):
        del self.dict[key.encode(self.keyencoding)]
        try:
            del self.cache[key]
        except KeyError:
            pass

    def __iter__(self):
        for k in self.dict.keys():
            yield k.decode(self.keyencoding)

    def items(self):
        return [(k, self[k]) for k in self]

    def values(self):
        return [self[k] for k in self]

    def setdefault(self, key, default=None):
        if key not in self:
            self[key] = default
        return self[key]

    def pop(self, key, *args):
        if key in self:
            v = self[key]
            del self[key]
            return v
        if args:
            return args[0]
        raise KeyError(key)

    def update(self, other=(), **kw):
        if hasattr(other, "keys"):
            for k in other.keys():
                self[k] = other[k]
        else:
            for k, v in other:
                self[k] = v
        for k in kw:
            self[k] = kw[k]

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        self.close()

    def close(self):
        if self.dict is None:
            return
        try:
            self.sync()
            try:
                self.dict.close()
            except AttributeError:
                pass
        finally:
            self.dict = _ClosedDict()

    def sync(self):
        if self.writeback and self.cache:
            self.writeback = False
            for key, entry in self.cache.items():
                self[key] = entry
            self.writeback = True
            self.cache = {}
        if hasattr(self.dict, "sync"):
            self.dict.sync()


class _ClosedDict:
    def keys(self) -> object:
        raise ValueError("invalid operation on closed shelf")

    def __getitem__(self, key) -> object:
        raise ValueError("invalid operation on closed shelf")

    def __setitem__(self, key, value):
        raise ValueError("invalid operation on closed shelf")

    def __contains__(self, key) -> bool:
        raise ValueError("invalid operation on closed shelf")

    def __len__(self) -> int:
        raise ValueError("invalid operation on closed shelf")


class BsdDbShelf(Shelf):
    pass


class DbfilenameShelf(Shelf):
    def __init__(self, filename, flag="c", protocol=None, writeback=False):
        Shelf.__init__(self, dbm.open(filename, flag), protocol, writeback)


def open(filename, flag="c", protocol=None, writeback=False):
    return DbfilenameShelf(filename, flag, protocol, writeback)
