"""Response classes used by urllib (hand-written; no tempfile wrapper)."""
__all__ = ['addbase', 'addclosehook', 'addinfo', 'addinfourl']


def _obj(x: object) -> object:
    return x


class addbase:
    """Minimal file-like wrapper delegating to the underlying object."""

    def __init__(self, fp):
        self.fp = _obj(fp)
        self.name = '<urllib response>'

    def __repr__(self):
        return '<%s at %r whose fp = %r>' % (self.__class__.__name__, id(self), self.fp)

    @property
    def closed(self):
        return self.fp.closed

    def read(self, n=-1):
        return self.fp.read(n)

    def read1(self, n=-1):
        return self.fp.read(n)

    def readline(self, limit=-1):
        return self.fp.readline(limit)

    def readlines(self, hint=-1):
        return self.fp.readlines(hint)

    def readable(self):
        return True

    def fileno(self):
        return self.fp.fileno()

    def __iter__(self):
        return self

    def __next__(self):
        line = self.fp.readline()
        if not line:
            raise StopIteration
        return line

    def close(self):
        self.fp.close()

    def __enter__(self):
        if self.fp.closed:
            raise ValueError('I/O operation on closed file')
        return self

    def __exit__(self, type, value, traceback):
        self.close()


class addclosehook(addbase):
    """Class to add a close hook to an open file."""

    def __init__(self, fp, closehook, *hookargs):
        addbase.__init__(self, fp)
        self.closehook = _obj(closehook)
        self.hookargs = _obj(hookargs)

    def close(self):
        try:
            closehook = self.closehook
            hookargs = self.hookargs
            if closehook:
                self.closehook = None
                self.hookargs = None
                closehook(*hookargs)
        finally:
            addbase.close(self)


class addinfo(addbase):
    """class to add an info() method to an open file."""

    def __init__(self, fp, headers):
        addbase.__init__(self, fp)
        self.headers = _obj(headers)

    def info(self):
        return self.headers


class addinfourl(addinfo):
    """class to add info() and geturl() methods to an open file."""

    def __init__(self, fp, headers, url, code=None):
        addinfo.__init__(self, fp, headers)
        self.url = _obj(url)
        self.code = _obj(code)

    @property
    def status(self):
        return self.code

    def getcode(self):
        return self.code

    def geturl(self):
        return self.url
