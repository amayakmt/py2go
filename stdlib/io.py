"""io: in-memory text and binary streams."""

SEEK_SET = 0
SEEK_CUR = 1
SEEK_END = 2
DEFAULT_BUFFER_SIZE = 8192


class UnsupportedOperation(OSError, ValueError):
    pass


class StringIO:
    def __init__(self, initial_value="", newline="\n"):
        self._newline = newline
        self._buf = ""
        self._pos = 0
        self._closed = False
        if initial_value:
            self.write(initial_value)
            self._pos = 0

    def _check(self):
        if self._closed:
            raise ValueError("I/O operation on closed file")

    @property
    def closed(self):
        return self._closed

    def close(self):
        self._closed = True

    def getvalue(self):
        self._check()
        return self._buf

    def write(self, s):
        self._check()
        if not isinstance(s, str):
            raise TypeError("string argument expected, got '" + type(s).__name__ + "'")
        if self._newline is None:
            s = s.replace("\r\n", "\n").replace("\r", "\n")
        elif self._newline == "\r" or self._newline == "\r\n":
            s = s.replace("\n", self._newline)
        if self._pos > len(self._buf):
            self._buf += "\0" * (self._pos - len(self._buf))
        self._buf = self._buf[:self._pos] + s + self._buf[self._pos + len(s):]
        self._pos += len(s)
        return len(s)

    def writelines(self, lines):
        for line in lines:
            self.write(line)

    def read(self, size=-1):
        self._check()
        if size is None or size < 0:
            data = self._buf[self._pos:]
        else:
            data = self._buf[self._pos:self._pos + size]
        self._pos += len(data)
        return data

    def readline(self, size=-1):
        self._check()
        if self._pos >= len(self._buf):
            return ""
        nl = self._newline
        if nl is None or nl == "":
            end = len(self._buf)
            i = self._pos
            n = len(self._buf)
            while i < n:
                ch = self._buf[i]
                if ch == "\n":
                    end = i + 1
                    break
                if ch == "\r":
                    end = i + 1
                    if i + 1 < n and self._buf[i + 1] == "\n":
                        end = i + 2
                    break
                i += 1
        else:
            if nl == "":
                nl = "\n"
            end = self._buf.find(nl, self._pos)
            end = len(self._buf) if end < 0 else end + len(nl)
        if size is not None and size >= 0 and end - self._pos > size:
            end = self._pos + size
        line = self._buf[self._pos:end]
        self._pos = end
        return line

    def readlines(self, hint=-1):
        lines = []
        while True:
            line = self.readline()
            if not line:
                return lines
            lines.append(line)

    def __iter__(self):
        return iter(self.readlines())

    def seek(self, pos, whence=0):
        self._check()
        if whence == 0:
            if pos < 0:
                raise ValueError("Negative seek position " + str(pos))
            self._pos = pos
        elif whence == 1:
            self._pos += pos
        else:
            self._pos = len(self._buf) + pos
        return self._pos

    def tell(self):
        self._check()
        return self._pos

    def truncate(self, size=None):
        self._check()
        if size is None:
            size = self._pos
        self._buf = self._buf[:size]
        return size

    def readable(self):
        return True

    def writable(self):
        return True

    def seekable(self):
        return True

    def flush(self):
        self._check()

    def isatty(self):
        return False

    def __enter__(self):
        self._check()
        return self

    def __exit__(self, exc_type, exc, tb):
        self.close()
        return False


class BytesIO:
    def __init__(self, initial_bytes=b""):
        self._buf = bytes(initial_bytes)
        self._pos = 0
        self._closed = False

    def _check(self):
        if self._closed:
            raise ValueError("I/O operation on closed file.")

    @property
    def closed(self):
        return self._closed

    def close(self):
        self._closed = True

    def getvalue(self):
        self._check()
        return self._buf

    def getbuffer(self):
        return bytearray(self._buf)

    def write(self, b):
        self._check()
        if isinstance(b, str):
            raise TypeError("a bytes-like object is required, not 'str'")
        data = bytes(b)
        if self._pos > len(self._buf):
            self._buf += b"\0" * (self._pos - len(self._buf))
        self._buf = self._buf[:self._pos] + data + self._buf[self._pos + len(data):]
        self._pos += len(data)
        return len(data)

    def writelines(self, lines):
        for line in lines:
            self.write(line)

    def read(self, size=-1):
        self._check()
        if size is None or size < 0:
            data = self._buf[self._pos:]
        else:
            data = self._buf[self._pos:self._pos + size]
        self._pos += len(data)
        return data

    def read1(self, size=-1):
        return self.read(size)

    def readline(self, size=-1):
        self._check()
        if self._pos >= len(self._buf):
            return b""
        end = self._buf.find(b"\n", self._pos)
        end = len(self._buf) if end < 0 else end + 1
        if size is not None and size >= 0 and end - self._pos > size:
            end = self._pos + size
        line = self._buf[self._pos:end]
        self._pos = end
        return line

    def readlines(self, hint=-1):
        lines = []
        while True:
            line = self.readline()
            if not line:
                return lines
            lines.append(line)

    def __iter__(self):
        return iter(self.readlines())

    def seek(self, pos, whence=0):
        self._check()
        if whence == 0:
            self._pos = pos
        elif whence == 1:
            self._pos += pos
        else:
            self._pos = len(self._buf) + pos
        return self._pos

    def tell(self):
        self._check()
        return self._pos

    def truncate(self, size=None):
        self._check()
        if size is None:
            size = self._pos
        self._buf = self._buf[:size]
        return size

    def readable(self):
        return True

    def writable(self):
        return True

    def seekable(self):
        return True

    def flush(self):
        self._check()

    def isatty(self):
        return False

    def __enter__(self):
        self._check()
        return self

    def __exit__(self, exc_type, exc, tb):
        self.close()
        return False
