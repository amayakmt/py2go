"""dbm.dumb: a dumb and slow but simple dbm clone (the .dat/.dir/.bak format of CPython's)."""
import os
import io

__all__ = ["error", "open"]

_BLOCKSIZE = 512

error = OSError


def _obj(x):
    return x


class _Database:
    def __init__(self, filebasename, mode, flag="c"):
        self._mode = mode
        self._readonly = flag == "r"
        self._dirfile = filebasename + ".dir"
        self._datfile = filebasename + ".dat"
        self._bakfile = filebasename + ".bak"
        self._index = _obj({})
        self._create(flag)
        self._update(flag)

    def _create(self, flag):
        if flag == "n":
            for filename in (self._datfile, self._bakfile, self._dirfile):
                try:
                    os.remove(filename)
                except OSError:
                    pass
        try:
            f = io.open(self._datfile, "r", encoding="Latin-1")
        except OSError:
            if flag not in ("c", "n"):
                raise
            with io.open(self._datfile, "w", encoding="Latin-1") as f:
                self._chmod(self._datfile)
        else:
            f.close()

    def _chmod(self, file):
        os.chmod(file, self._mode)

    def _update(self, flag):
        self._modified = False
        self._index = _obj({})
        try:
            f = io.open(self._dirfile, "r", encoding="Latin-1")
        except OSError:
            if flag not in ("c", "n"):
                raise
            self._modified = True
        else:
            with f:
                for line in f:
                    line = line.rstrip()
                    key, pos_and_siz_pair = _parse_line(line)
                    key = key.encode("Latin-1")
                    self._index[key] = pos_and_siz_pair

    def _commit(self):
        if self._index is None or not self._modified:
            return
        try:
            os.unlink(self._bakfile)
        except OSError:
            pass
        try:
            os.rename(self._dirfile, self._bakfile)
        except OSError:
            pass
        with io.open(self._dirfile, "w", encoding="Latin-1") as f:
            self._chmod(self._dirfile)
            for key, pos_and_siz_pair in self._index.items():
                entry = "%r, %r\n" % (key.decode("Latin-1"), pos_and_siz_pair)
                f.write(entry)

    def sync(self):
        self._commit()

    def _verify_open(self):
        if self._index is None:
            raise error("DBM object has already been closed")

    def __getitem__(self, key):
        orig = key
        if isinstance(key, str):
            key = key.encode("utf-8")
        self._verify_open()
        if key not in self._index:
            raise KeyError(orig)
        pos, siz = self._index[key]
        with io.open(self._datfile, "rb") as f:
            f.seek(pos)
            dat = f.read(siz)
        return dat

    def _addval(self, val):
        with io.open(self._datfile, "rb+") as f:
            f.seek(0, 2)
            pos = int(f.tell())
            npos = ((pos + _BLOCKSIZE - 1) // _BLOCKSIZE) * _BLOCKSIZE
            f.write(b"\0" * (npos - pos))
            pos = npos
            f.write(val)
        return (pos, len(val))

    def _setval(self, pos, val):
        with io.open(self._datfile, "rb+") as f:
            f.seek(pos)
            f.write(val)
        return (pos, len(val))

    def _addkey(self, key, pos_and_siz_pair):
        self._index[key] = pos_and_siz_pair
        with io.open(self._dirfile, "a", encoding="Latin-1") as f:
            self._chmod(self._dirfile)
            f.write("%r, %r\n" % (key.decode("Latin-1"), pos_and_siz_pair))

    def __setitem__(self, key, val):
        if self._readonly:
            raise error("The database is opened for reading only")
        if isinstance(key, str):
            key = key.encode("utf-8")
        elif not isinstance(key, (bytes, bytearray)):
            raise TypeError("keys must be bytes or strings")
        if isinstance(val, str):
            val = val.encode("utf-8")
        elif not isinstance(val, (bytes, bytearray)):
            raise TypeError("values must be bytes or strings")
        self._verify_open()
        if key not in self._index:
            self._addkey(key, self._addval(val))
        else:
            pos, siz = self._index[key]
            oldblocks = (siz + _BLOCKSIZE - 1) // _BLOCKSIZE
            newblocks = (len(val) + _BLOCKSIZE - 1) // _BLOCKSIZE
            if newblocks <= oldblocks:
                self._index[key] = self._setval(pos, val)
            else:
                self._index[key] = self._addval(val)
            self._modified = True

    def __delitem__(self, key):
        if self._readonly:
            raise error("The database is opened for reading only")
        if isinstance(key, str):
            key = key.encode("utf-8")
        self._verify_open()
        del self._index[key]
        self._modified = True
        self._commit()

    def keys(self):
        try:
            return list(self._index)
        except TypeError:
            raise error("DBM object has already been closed") from None

    def items(self):
        self._verify_open()
        return [(key, self[key]) for key in self._index.keys()]

    def __contains__(self, key):
        if isinstance(key, str):
            key = key.encode("utf-8")
        try:
            return key in self._index
        except TypeError:
            if self._index is None:
                raise error("DBM object has already been closed") from None
            else:
                raise

    def get(self, key, default=None):
        if key in self:
            return self[key]
        return default

    def setdefault(self, key, default=b""):
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

    def iterkeys(self):
        try:
            return iter(self._index)
        except TypeError:
            raise error("DBM object has already been closed") from None

    def __len__(self):
        try:
            return len(self._index)
        except TypeError:
            raise error("DBM object has already been closed") from None

    def close(self):
        try:
            self._commit()
        finally:
            self._index = _obj(None)
            self._datfile = self._dirfile = self._bakfile = ""

    def __enter__(self):
        return self

    def __exit__(self, *args):
        self.close()


def _parse_line(line):
    # "'key', (pos, size)"
    q = line[0]
    i = 1
    key = ""
    while i < len(line):
        c = line[i]
        if c == "\\" and i + 1 < len(line):
            n = line[i + 1]
            if n == "x":
                key += chr(int(line[i + 2:i + 4], 16))
                i += 4
                continue
            m = {"n": "\n", "t": "\t", "r": "\r", "\\": "\\", "'": "'", '"': '"'}
            key += m.get(n, n)
            i += 2
            continue
        if c == q:
            i += 1
            break
        key += c
        i += 1
    rest = line[i:].strip()
    rest = rest.lstrip(",").strip().strip("()")
    pos, siz = rest.split(",")
    return key, (int(pos), int(siz))


def open(file, flag="c", mode=0o666):
    if flag not in ("r", "w", "c", "n"):
        raise ValueError("Flag must be one of 'r', 'w', 'c', or 'n'")
    return _Database(os.fspath(file), mode, flag)
