"""gzip: functions that read and write gzipped files."""

import struct
import time
import zlib
import io
import os

from builtins import open as _builtin_open
from builtins import gzip_pm_BadGzipFile as BadGzipFile

FTEXT, FHCRC, FEXTRA, FNAME, FCOMMENT = 1, 2, 4, 8, 16
READ = "rb"
WRITE = "wb"
_COMPRESS_LEVEL_FAST = 1
_COMPRESS_LEVEL_TRADEOFF = 6
_COMPRESS_LEVEL_BEST = 9


def _header(mtime, compresslevel, name):
    flags = 0
    fname = b""
    if name:
        fname = os.path.basename(name)
        if fname.endswith(".gz"):
            fname = fname[:-3]
        fname = fname.encode("latin-1", "replace") + b"\x00" if False else (fname if isinstance(fname, bytes) else fname.encode("latin-1", "replace")) + b"\x00"
        flags = FNAME
    xfl = 0
    if compresslevel == _COMPRESS_LEVEL_BEST:
        xfl = 2
    elif compresslevel == _COMPRESS_LEVEL_FAST:
        xfl = 4
    return b"\x1f\x8b\x08" + bytes([flags]) + struct.pack("<L", int(mtime) & 0xffffffff) + bytes([xfl]) + b"\xff" + fname


def _member(data, compresslevel, mtime, name):
    body = zlib.compress(data, compresslevel, -15)
    return _header(mtime, compresslevel, name) + body + struct.pack("<LL", zlib.crc32(data) & 0xffffffff, len(data) & 0xffffffff)


def _read_members(raw, short=False):
    out = []
    pos = 0
    n = len(raw)
    first = True
    mtime = 0
    while pos < n:
        if raw[pos:pos + 2] != b"\x1f\x8b":
            if first:
                raise BadGzipFile("Not a gzipped file (" + repr(raw[pos:pos + 2]) + ")")
            break
        if raw[pos + 2] != 8:
            raise BadGzipFile("Unknown compression method")
        flag = raw[pos + 3]
        if first:
            mtime = struct.unpack("<L", raw[pos + 4:pos + 8])[0]
        p = pos + 10
        if flag & FEXTRA:
            xlen = struct.unpack("<H", raw[p:p + 2])[0]
            p += 2 + xlen
        if flag & FNAME:
            while raw[p] != 0:
                p += 1
            p += 1
        if flag & FCOMMENT:
            while raw[p] != 0:
                p += 1
            p += 1
        if flag & FHCRC:
            p += 2
        d = zlib.decompressobj(-15)
        try:
            chunk = d.decompress(raw[p:])
        except zlib.error as e:
            raise BadGzipFile(str(e)) from None
        if not d.eof:
            raise EOFError("Compressed file ended before the end-of-stream marker was reached")
        rest = d.unused_data
        if len(rest) < 8:
            raise EOFError("Compressed file ended before the end-of-stream marker was reached")
        crc, isize = struct.unpack("<LL", rest[:8])
        if crc != zlib.crc32(chunk) & 0xffffffff:
            if short:
                raise BadGzipFile("CRC check failed")
            raise BadGzipFile("CRC check failed " + hex(zlib.crc32(chunk) & 0xffffffff) + " != " + hex(crc))
        if isize != len(chunk) & 0xffffffff:
            raise BadGzipFile("Incorrect length of data produced")
        out.append(chunk)
        pos = n - (len(rest) - 8)
        first = False
    return b"".join(out), mtime


class GzipFile:
    def __init__(self, filename=None, mode=None, compresslevel=9, fileobj=None, mtime=None):
        if mode and ("t" in mode or "U" in mode):
            raise ValueError("Invalid mode: " + repr(mode))
        if mode and "b" not in mode:
            mode += "b"
        if fileobj is None:
            fileobj = _builtin_open(filename, mode or "rb")
            self._own = True
        else:
            self._own = False
        if filename is None:
            filename = getattr(fileobj, "name", "")
            if not isinstance(filename, str):
                filename = ""
        if mode is None:
            mode = getattr(fileobj, "mode", "rb")
        self.mode = READ if mode.startswith("r") else WRITE
        self.name = filename if isinstance(filename, str) else str(filename)
        self.fileobj = fileobj
        self.compresslevel = compresslevel
        self._mtime = mtime
        self._closed = False
        self._buf = b""
        self._pos = 0
        self._loaded = False
        self.mtime = None
        if self.mode == WRITE:
            self._pending = bytearray()
            self.mtime = mtime

    @property
    def closed(self):
        return self._closed

    def _load(self):
        if not self._loaded:
            raw = self.fileobj.read()
            data, mtime = _read_members(raw)
            self._buf = data
            self.mtime = mtime
            self._loaded = True

    def _check_read(self):
        if self._closed:
            raise ValueError("I/O operation on closed file.")
        if self.mode != READ:
            raise io.UnsupportedOperation("read")
        self._load()

    def read(self, size=-1):
        self._check_read()
        if size is None or size < 0:
            data = self._buf[self._pos:]
        else:
            data = self._buf[self._pos:self._pos + size]
        self._pos += len(data)
        return data

    def read1(self, size=-1):
        return self.read(size)

    def peek(self, n):
        self._check_read()
        return self._buf[self._pos:self._pos + max(n, 1)]

    def readline(self, size=-1):
        self._check_read()
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

    def readable(self):
        return self.mode == READ

    def writable(self):
        return self.mode == WRITE

    def seekable(self):
        return True

    def tell(self):
        if self.mode == READ:
            return self._pos
        return len(self._pending)

    def seek(self, offset, whence=0):
        if self.mode == WRITE:
            raise OSError("Seek from end not supported" if whence == 2 else "Negative seek in write mode")
        self._load()
        if whence == 1:
            offset += self._pos
        elif whence == 2:
            offset += len(self._buf)
        self._pos = max(offset, 0)
        return self._pos

    def write(self, data):
        if self._closed:
            raise ValueError("write() on closed GzipFile object")
        if self.mode != WRITE:
            raise OSError(9, "write() on read-only GzipFile object")
        if isinstance(data, str):
            raise TypeError("a bytes-like object is required, not 'str'")
        self._pending.extend(data)
        return len(data)

    def flush(self, zlib_mode=2):
        if self.mode == WRITE and not self._closed:
            pass

    def close(self):
        if self._closed:
            return
        self._closed = True
        if self.mode == WRITE:
            mtime = self._mtime
            if mtime is None:
                mtime = time.time()
            self.fileobj.write(_member(bytes(self._pending), self.compresslevel, mtime, self.name))
        if self._own:
            self.fileobj.close()

    def __enter__(self):
        if self._closed:
            raise ValueError("I/O operation on closed file.")
        return self

    def __exit__(self, t, v, tb):
        self.close()

    def fileno(self):
        return self.fileobj.fileno()


class _GzipText:
    def __init__(self, gz, encoding, errors, newline):
        self._gz = gz
        self._encoding = encoding or "utf-8"
        self._errors = errors or "strict"
        self.mode = "rt" if gz.mode == READ else "wt"
        self.name = gz.name
        self._lines = None
        self._text = None
        self._pos = 0

    def _load(self):
        if self._text is None:
            self._text = self._gz.read().decode(self._encoding, self._errors)

    def read(self, size=-1):
        self._load()
        if size is None or size < 0:
            data = self._text[self._pos:]
        else:
            data = self._text[self._pos:self._pos + size]
        self._pos += len(data)
        return data

    def readline(self, size=-1):
        self._load()
        end = self._text.find("\n", self._pos)
        end = len(self._text) if end < 0 else end + 1
        line = self._text[self._pos:end]
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

    def write(self, s):
        self._gz.write(s.encode(self._encoding, self._errors))
        return len(s)

    def writelines(self, lines):
        for line in lines:
            self.write(line)

    def flush(self):
        pass

    def close(self):
        self._gz.close()

    @property
    def closed(self):
        return self._gz.closed

    def __enter__(self):
        return self

    def __exit__(self, t, v, tb):
        self.close()


def open(filename, mode="rb", compresslevel=9, encoding=None, errors=None, newline=None):
    if "t" in mode:
        if "b" in mode:
            raise ValueError("Invalid mode: " + repr(mode))
    else:
        if encoding is not None:
            raise ValueError("Argument 'encoding' not supported in binary mode")
        if errors is not None:
            raise ValueError("Argument 'errors' not supported in binary mode")
        if newline is not None:
            raise ValueError("Argument 'newline' not supported in binary mode")
    gz_mode = mode.replace("t", "")
    if isinstance(filename, (str, bytes)):
        gz = GzipFile(filename, gz_mode, compresslevel)
    elif hasattr(filename, "read") or hasattr(filename, "write"):
        gz = GzipFile(None, gz_mode, compresslevel, filename)
    else:
        raise TypeError("filename must be a str or bytes object, or a file")
    if "t" in mode:
        return _GzipText(gz, encoding, errors, newline)
    return gz


def compress(data, compresslevel=9, *, mtime=None):
    if mtime is None:
        mtime = time.time()
    return _member(bytes(data), compresslevel, mtime, "")


def decompress(data):
    out, mtime = _read_members(bytes(data), True)
    return out
