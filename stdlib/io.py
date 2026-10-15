"""io: in-memory text and binary streams."""

SEEK_SET = 0
SEEK_CUR = 1
SEEK_END = 2
DEFAULT_BUFFER_SIZE = 8192


from builtins import io_pm_UnsupportedOperation as UnsupportedOperation
from builtins import open as _builtin_open


def open(file, mode="r", buffering=-1, encoding=None, errors=None, newline=None, closefd=True, opener=None):
    return _builtin_open(file, mode, buffering, encoding, errors, newline, closefd)


def open_code(path):
    return _builtin_open(path, "rb")


def text_encoding(encoding, stacklevel=2):
    if encoding is None:
        return "locale"
    return encoding


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


class IOBase:
    def __init__(self):
        self._iobase_closed = False

    def _checkClosed(self):
        if self._iobase_closed:
            raise ValueError("I/O operation on closed file.")

    @property
    def closed(self):
        return self._iobase_closed

    def close(self):
        if not self._iobase_closed:
            try:
                self.flush()
            finally:
                self._iobase_closed = True

    def flush(self):
        self._checkClosed()

    def seekable(self):
        return False

    def readable(self):
        return False

    def writable(self):
        return False

    def isatty(self):
        self._checkClosed()
        return False

    def fileno(self):
        raise UnsupportedOperation("fileno")

    def seek(self, pos, whence=0):
        raise UnsupportedOperation("seek")

    def tell(self):
        return self.seek(0, 1)

    def truncate(self, pos=None):
        raise UnsupportedOperation("truncate")

    def __enter__(self):
        self._checkClosed()
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        self.close()

    def __iter__(self):
        self._checkClosed()
        return self

    def __next__(self):
        line = self.readline()
        if not line:
            raise StopIteration
        return line

    def readline(self, size=-1):
        if size is None:
            size = -1
        res = bytearray()
        while size < 0 or len(res) < size:
            b = self.read(1)
            if not b:
                break
            res += b
            if res.endswith(b"\n"):
                break
        return bytes(res)

    def readlines(self, hint=None):
        if hint is None or hint <= 0:
            return list(self)
        n = 0
        lines = []
        for line in self:
            lines.append(line)
            n += len(line)
            if n >= hint:
                break
        return lines

    def writelines(self, lines):
        self._checkClosed()
        for line in lines:
            self.write(line)


class RawIOBase(IOBase):
    def read(self, size=-1):
        if size is None:
            size = -1
        if size < 0:
            return self.readall()
        b = bytearray(size)
        n = self.readinto(b)
        if n is None:
            return None
        return bytes(b[:n])

    def readall(self):
        res = bytearray()
        while True:
            data = self.read(8192)
            if not data:
                break
            res += data
        if res:
            return bytes(res)
        return data

    def readinto(self, b):
        raise UnsupportedOperation("readinto")

    def write(self, b):
        raise UnsupportedOperation("write")


class BufferedIOBase(IOBase):
    def read(self, size=-1):
        raise UnsupportedOperation("read")

    def read1(self, size=-1):
        raise UnsupportedOperation("read1")

    def readinto(self, b):
        data = self.read(len(b))
        n = len(data)
        b[:n] = data
        return n

    def readinto1(self, b):
        data = self.read1(len(b))
        n = len(data)
        b[:n] = data
        return n

    def write(self, b):
        raise UnsupportedOperation("write")

    def detach(self):
        raise UnsupportedOperation("detach")


class TextIOBase(IOBase):
    @property
    def encoding(self) -> object:
        return None

    @property
    def errors(self) -> object:
        return None

    @property
    def newlines(self) -> object:
        return None

    def read(self, size=-1):
        raise UnsupportedOperation("read")

    def write(self, s):
        raise UnsupportedOperation("write")

    def readline(self, size=-1):
        raise UnsupportedOperation("readline")

    def detach(self):
        raise UnsupportedOperation("detach")


class BufferedReader(BufferedIOBase):
    def __init__(self, raw, buffer_size=DEFAULT_BUFFER_SIZE):
        IOBase.__init__(self)
        self.raw = raw
        self._rbuf = b""

    def readable(self):
        return True

    def read(self, size=-1):
        if size is None or size < 0:
            data = self._rbuf + self.raw.read()
            self._rbuf = b""
            return data
        while len(self._rbuf) < size:
            chunk = self.raw.read(size - len(self._rbuf))
            if not chunk:
                break
            self._rbuf += chunk
        data = self._rbuf[:size]
        self._rbuf = self._rbuf[size:]
        return data

    def read1(self, size=-1):
        if not self._rbuf:
            self._rbuf = self.raw.read(8192 if size is None or size < 0 else size) or b""
        if size is None or size < 0:
            size = len(self._rbuf)
        data = self._rbuf[:size]
        self._rbuf = self._rbuf[size:]
        return data

    def peek(self, size=0):
        if not self._rbuf:
            self._rbuf = self.raw.read(8192) or b""
        return self._rbuf

    def readline(self, size=-1):
        res = b""
        while size is None or size < 0 or len(res) < size:
            i = self._rbuf.find(b"\n")
            if i >= 0:
                res += self._rbuf[:i + 1]
                self._rbuf = self._rbuf[i + 1:]
                break
            res += self._rbuf
            self._rbuf = self.raw.read(8192) or b""
            if not self._rbuf:
                break
        return res

    def close(self):
        if not self.raw.closed:
            self.raw.close()
        self._iobase_closed = True

    @property
    def closed(self):
        return self.raw.closed

    def seekable(self):
        return self.raw.seekable()

    def seek(self, pos, whence=0):
        self._rbuf = b""
        return self.raw.seek(pos, whence)

    def tell(self):
        return self.raw.tell() - len(self._rbuf)

    def fileno(self):
        return self.raw.fileno()

    def detach(self):
        raw = self.raw
        return raw


class BufferedWriter(BufferedIOBase):
    def __init__(self, raw, buffer_size=DEFAULT_BUFFER_SIZE):
        IOBase.__init__(self)
        self.raw = raw
        self._wbuf = []

    def writable(self):
        return True

    def write(self, b):
        self._wbuf.append(bytes(b))
        return len(b)

    def flush(self):
        if self._wbuf:
            data = b"".join(self._wbuf)
            self._wbuf = []
            self.raw.write(data)
        if hasattr(self.raw, "flush"):
            self.raw.flush()

    def close(self):
        if not self.raw.closed:
            try:
                self.flush()
            finally:
                self.raw.close()
        self._iobase_closed = True

    @property
    def closed(self):
        return self.raw.closed

    def seekable(self):
        return self.raw.seekable()

    def seek(self, pos, whence=0):
        self.flush()
        return self.raw.seek(pos, whence)

    def tell(self):
        self.flush()
        return self.raw.tell()

    def fileno(self):
        return self.raw.fileno()


class BufferedRandom(BufferedReader):
    def __init__(self, raw, buffer_size=DEFAULT_BUFFER_SIZE):
        BufferedReader.__init__(self, raw, buffer_size)

    def writable(self):
        return True

    def write(self, b):
        if self._rbuf:
            self.raw.seek(-len(self._rbuf), 1)
            self._rbuf = b""
        return self.raw.write(b)

    def flush(self):
        if hasattr(self.raw, "flush"):
            self.raw.flush()


class BufferedRWPair(BufferedIOBase):
    def __init__(self, reader, writer, buffer_size=DEFAULT_BUFFER_SIZE):
        IOBase.__init__(self)
        self.reader = BufferedReader(reader)
        self.writer = BufferedWriter(writer)

    def read(self, size=-1):
        return self.reader.read(size)

    def write(self, b):
        return self.writer.write(b)

    def flush(self):
        self.writer.flush()

    def readable(self):
        return True

    def writable(self):
        return True

    def close(self):
        try:
            self.writer.close()
        finally:
            self.reader.close()


class TextIOWrapper(TextIOBase):
    def __init__(self, buffer, encoding=None, errors=None, newline=None, line_buffering=False, write_through=False):
        IOBase.__init__(self)
        if newline is not None and newline not in ("", "\n", "\r", "\r\n"):
            raise ValueError("illegal newline value: %s" % newline)
        self.buffer = buffer
        self._encoding = "utf-8" if encoding is None else encoding
        self._errors = "strict" if errors is None else errors
        self._newline = newline
        self.line_buffering = line_buffering
        self.write_through = write_through
        self._text = ""
        self._pos = 0
        self._loaded = False
        self._seen = 0
        self._detached = False

    @property
    def encoding(self) -> object:
        return self._encoding

    @property
    def errors(self) -> object:
        return self._errors

    @property
    def newlines(self) -> object:
        if self._newline is not None and self._newline != "":
            return None
        if self._seen == 0:
            return None
        names = []
        if self._seen & 2:
            names.append("\r")
        if self._seen & 1:
            names.append("\n")
        if self._seen & 4:
            names.append("\r\n")
        if len(names) == 1:
            return names[0]
        return tuple(names)

    @property
    def name(self):
        return self.buffer.name

    @property
    def closed(self):
        return self.buffer.closed

    def readable(self):
        return self.buffer.readable()

    def writable(self):
        return self.buffer.writable()

    def seekable(self):
        return self.buffer.seekable()

    def fileno(self):
        return self.buffer.fileno()

    def isatty(self):
        return self.buffer.isatty()

    def _load(self):
        if self._loaded:
            return
        self._loaded = True
        data = self.buffer.read()
        if data is None:
            data = b""
        text = bytes(data).decode(self._encoding, self._errors)
        if self._newline is None:
            if "\r\n" in text:
                self._seen |= 4
            text = text.replace("\r\n", "\n")
            if "\r" in text:
                self._seen |= 2
            if "\n" in text:
                self._seen |= 1
            text = text.replace("\r", "\n")
        elif self._newline == "":
            self._seen |= 1 if "\n" in text else 0
        self._text = text
        self._pos = 0

    def read(self, size=-1):
        if self.closed:
            raise ValueError("I/O operation on closed file.")
        self._load()
        if size is None or size < 0:
            r = self._text[self._pos:]
            self._pos = len(self._text)
        else:
            r = self._text[self._pos:self._pos + size]
            self._pos += len(r)
        return r

    def readline(self, size=-1):
        if self.closed:
            raise ValueError("I/O operation on closed file.")
        self._load()
        text = self._text
        start = self._pos
        if self._newline is None or self._newline == "\n":
            term = "\n"
            i = text.find("\n", start)
            end = len(text) if i < 0 else i + 1
        elif self._newline == "\r":
            i = text.find("\r", start)
            end = len(text) if i < 0 else i + 1
        elif self._newline == "\r\n":
            i = text.find("\r\n", start)
            end = len(text) if i < 0 else i + 2
        else:
            i = start
            end = len(text)
            while i < len(text):
                if text[i] == "\n":
                    end = i + 1
                    break
                if text[i] == "\r":
                    end = i + 2 if text[i + 1:i + 2] == "\n" else i + 1
                    break
                i += 1
        if size is not None and size >= 0 and end - start > size:
            end = start + size
        self._pos = end
        return text[start:end]

    def __next__(self):
        line = self.readline()
        if not line:
            raise StopIteration
        return line

    def write(self, s):
        if self.closed:
            raise ValueError("I/O operation on closed file.")
        if not isinstance(s, str):
            raise TypeError("write() argument must be str, not " + type(s).__name__)
        text = s
        if self._newline == "\r" or self._newline == "\r\n":
            text = text.replace("\n", self._newline)
        self.buffer.write(text.encode(self._encoding, self._errors))
        if self.write_through or (self.line_buffering and ("\n" in s or "\r" in s)):
            self.flush()
        return len(s)

    def flush(self):
        if self.closed:
            raise ValueError("I/O operation on closed file.")
        self.buffer.flush()

    def close(self):
        if not self.buffer.closed:
            try:
                self.flush()
            finally:
                self.buffer.close()

    def seek(self, cookie, whence=0):
        if whence == 0:
            self._loaded = False
            self._text = ""
            self._pos = 0
            self.buffer.seek(cookie)
            return cookie
        if whence == 1 and cookie == 0:
            return self.tell()
        if whence == 2 and cookie == 0:
            self._loaded = True
            self._text = ""
            self._pos = 0
            return self.buffer.seek(0, 2)
        raise UnsupportedOperation("can't do nonzero cur-relative seeks")

    def tell(self):
        if self._loaded:
            return len(self._text[:self._pos].encode(self._encoding, self._errors))
        return self.buffer.tell()

    def detach(self):
        buffer = self.buffer
        self._detached = True
        return buffer

    def truncate(self, pos=None):
        self.flush()
        return self.buffer.truncate(pos)


def FileIO(name, mode="r", closefd=True, opener=None):
    m = mode if "b" in mode else mode + "b"
    return _builtin_open(name, m, 0, None, None, None, closefd)

