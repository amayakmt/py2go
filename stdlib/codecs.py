"""codecs: access to the codec registry and base classes for codecs."""

import sys

BOM_UTF8 = b"\xef\xbb\xbf"
BOM_LE = b"\xff\xfe"
BOM_BE = b"\xfe\xff"
BOM_UTF16_LE = b"\xff\xfe"
BOM_UTF16_BE = b"\xfe\xff"
BOM_UTF16 = b"\xff\xfe"
BOM_UTF32_LE = b"\xff\xfe\x00\x00"
BOM_UTF32_BE = b"\x00\x00\xfe\xff"
BOM_UTF32 = b"\xff\xfe\x00\x00"
BOM = b"\xff\xfe"

_NAMES = {
    "utf8": "utf-8", "utf_8": "utf-8", "u8": "utf-8", "utf": "utf-8", "cp65001": "utf-8",
    "utf8sig": "utf-8-sig", "utf_8_sig": "utf-8-sig",
    "ascii": "ascii", "us_ascii": "ascii", "646": "ascii", "ansi_x3.4_1968": "ascii",
    "latin1": "iso8859-1", "latin_1": "iso8859-1", "iso8859_1": "iso8859-1", "iso_8859_1": "iso8859-1", "l1": "iso8859-1", "latin": "iso8859-1", "8859": "iso8859-1", "cp819": "iso8859-1", "iso88591": "iso8859-1",
    "utf16": "utf-16", "utf_16": "utf-16", "u16": "utf-16", "utf16le": "utf-16-le", "utf_16_le": "utf-16-le", "utf16be": "utf-16-be", "utf_16_be": "utf-16-be",
    "utf32": "utf-32", "utf_32": "utf-32", "u32": "utf-32", "utf32le": "utf-32-le", "utf_32_le": "utf-32-le", "utf32be": "utf-32-be", "utf_32_be": "utf-32-be",
    "unicode_escape": "unicode-escape", "unicodeescape": "unicode-escape", "raw_unicode_escape": "raw-unicode-escape",
    "hex": "hex", "hex_codec": "hex", "rot13": "rot-13", "rot_13": "rot-13",
}

_registry = []
_cache = {}
_handlers = {}


class CodecInfo:
    def __init__(self, encode, decode, streamreader=None, streamwriter=None, incrementalencoder=None, incrementaldecoder=None, name=None, *, _is_text_encoding=None):
        self.name = name
        self.encode = encode
        self.decode = decode
        self.incrementalencoder = incrementalencoder
        self.incrementaldecoder = incrementaldecoder
        self.streamreader = streamreader
        self.streamwriter = streamwriter
        if _is_text_encoding is not None:
            self._is_text_encoding = _is_text_encoding

    def __repr__(self):
        return "<codecs.CodecInfo object for encoding " + str(self.name) + " at " + hex(id(self)) + ">"

    def __getitem__(self, i):
        return (self.encode, self.decode, self.streamreader, self.streamwriter)[i]

    def __iter__(self):
        return iter((self.encode, self.decode, self.streamreader, self.streamwriter))

    def __len__(self):
        return 4


def _normalize(encoding):
    out = []
    for ch in encoding.lower():
        if ch == " " or ch == "-":
            out.append("_")
        else:
            out.append(ch)
    return "".join(out)


def _canonical(encoding):
    key = _normalize(encoding)
    if key in _NAMES:
        return _NAMES[key]
    return None


def _builtin_encode(name):
    def encode(obj, errors="strict"):
        if name == "hex":
            return (obj.hex().encode("ascii"), len(obj))
        if name == "rot-13":
            return (_rot13(obj), len(obj))
        return (obj.encode(name, errors), len(obj))
    return encode


def _builtin_decode(name):
    def decode(obj, errors="strict"):
        if name == "hex":
            return (bytes.fromhex(bytes(obj).decode("ascii")), len(obj))
        if name == "rot-13":
            return (_rot13(obj), len(obj))
        return (bytes(obj).decode(name, errors), len(obj))
    return decode


def _rot13(s):
    out = []
    for ch in s:
        o = ord(ch)
        if 97 <= o <= 122:
            out.append(chr((o - 97 + 13) % 26 + 97))
        elif 65 <= o <= 90:
            out.append(chr((o - 65 + 13) % 26 + 65))
        else:
            out.append(ch)
    return "".join(out)


def register(search_function):
    if not callable(search_function):
        raise TypeError("argument must be callable")
    _registry.append(search_function)


def unregister(search_function):
    if search_function in _registry:
        _registry.remove(search_function)
    _cache.clear()


def lookup(encoding):
    key = _normalize(encoding)
    if key in _cache:
        return _cache[key]
    for fn in _registry:
        info = fn(key)
        if info is not None:
            if not isinstance(info, CodecInfo):
                raise TypeError("codec search functions must return 4-tuples")
            _cache[key] = info
            return info
    canon = _canonical(encoding)
    if canon is None:
        raise LookupError("unknown encoding: " + encoding)
    dec = IncrementalDecoder
    if canon == "utf-8":
        dec = _UTF8IncrementalDecoder
    info = CodecInfo(_builtin_encode(canon), _builtin_decode(canon), StreamReader, StreamWriter, IncrementalEncoder, dec, canon)
    _cache[key] = info
    return info


def encode(obj, encoding="utf-8", errors="strict"):
    info = lookup(encoding)
    r = info.encode(obj, errors)
    return r[0]


def decode(obj, encoding="utf-8", errors="strict"):
    info = lookup(encoding)
    r = info.decode(obj, errors)
    return r[0]


def getencoder(encoding):
    return lookup(encoding).encode


def getdecoder(encoding):
    return lookup(encoding).decode


def getincrementalencoder(encoding):
    return lookup(encoding).incrementalencoder


def getincrementaldecoder(encoding):
    return lookup(encoding).incrementaldecoder


def getreader(encoding):
    return lookup(encoding).streamreader


def getwriter(encoding):
    return lookup(encoding).streamwriter


def _strict_errors(exc):
    raise exc


def register_error(name, handler):
    if not callable(handler):
        raise TypeError("handler must be callable")
    _handlers[name] = handler


def lookup_error(name):
    if name in _handlers:
        return _handlers[name]
    if name in ("strict", "ignore", "replace", "xmlcharrefreplace", "backslashreplace", "namereplace", "surrogateescape", "surrogatepass"):
        return _strict_errors
    raise LookupError("unknown error handler name '" + name + "'")


def strict_errors(exc):
    raise exc


def _utf8_tail(data):
    n = len(data)
    i = n - 1
    back = 0
    while i >= 0 and back < 4 and (data[i] & 0xC0) == 0x80:
        i -= 1
        back += 1
    if i < 0:
        return 0
    lead = data[i]
    if lead < 0xC2:
        return 0
    need = 2 if lead < 0xE0 else 3 if lead < 0xF0 else 4 if lead < 0xF5 else 0
    have = n - i
    if need and have < need:
        return have
    return 0


def utf_8_encode(s, errors="strict"):
    return (s.encode("utf-8", errors), len(s))


def utf_8_decode(data, errors="strict", final=False):
    b = bytes(data)
    tail = 0
    if not final:
        tail = _utf8_tail(b)
    if tail:
        b = b[:len(b) - tail]
    return (b.decode("utf-8", errors), len(b))


def latin_1_encode(s, errors="strict"):
    return (s.encode("latin-1", errors), len(s))


def latin_1_decode(data, errors="strict"):
    return (bytes(data).decode("latin-1", errors), len(data))


def ascii_encode(s, errors="strict"):
    return (s.encode("ascii", errors), len(s))


def ascii_decode(data, errors="strict"):
    return (bytes(data).decode("ascii", errors), len(data))


def utf_16_encode(s, errors="strict"):
    return (s.encode("utf-16", errors), len(s))


def utf_16_decode(data, errors="strict", final=False):
    return (bytes(data).decode("utf-16", errors), len(data))


def utf_16_le_encode(s, errors="strict"):
    return (s.encode("utf-16-le", errors), len(s))


def utf_16_le_decode(data, errors="strict", final=False):
    return (bytes(data).decode("utf-16-le", errors), len(data))


def utf_16_be_encode(s, errors="strict"):
    return (s.encode("utf-16-be", errors), len(s))


def utf_16_be_decode(data, errors="strict", final=False):
    return (bytes(data).decode("utf-16-be", errors), len(data))


def utf_32_encode(s, errors="strict"):
    return (s.encode("utf-32", errors), len(s))


def utf_32_decode(data, errors="strict", final=False):
    return (bytes(data).decode("utf-32", errors), len(data))


def unicode_escape_encode(s, errors="strict"):
    return (s.encode("unicode-escape", errors), len(s))


def unicode_escape_decode(data, errors="strict", final=True):
    return (bytes(data).decode("unicode-escape", errors), len(data))


def escape_encode(data, errors=None):
    out = bytearray()
    for b in data:
        if b == 92:
            out.extend(b"\\\\")
        elif b == 39:
            out.extend(b"\\'")
        elif b == 9:
            out.extend(b"\\t")
        elif b == 10:
            out.extend(b"\\n")
        elif b == 13:
            out.extend(b"\\r")
        elif b < 32 or b >= 127:
            out.extend(("\\x%02x" % b).encode("ascii"))
        else:
            out.append(b)
    return (bytes(out), len(data))


class Codec:
    def encode(self, input, errors="strict"):
        raise NotImplementedError

    def decode(self, input, errors="strict"):
        raise NotImplementedError


class IncrementalEncoder:
    def __init__(self, errors="strict"):
        self.errors = errors
        self.buffer = ""
        self._name = "utf-8"

    def encode(self, input, final=False):
        raise NotImplementedError

    def reset(self):
        pass

    def getstate(self):
        return 0

    def setstate(self, state):
        pass


class IncrementalDecoder:
    def __init__(self, errors="strict"):
        self.errors = errors
        self.buffer = b""

    def decode(self, input, final=False):
        raise NotImplementedError

    def reset(self):
        self.buffer = b""

    def getstate(self):
        return (self.buffer, 0)

    def setstate(self, state):
        self.buffer = state[0]


class BufferedIncrementalEncoder(IncrementalEncoder):
    def __init__(self, errors="strict"):
        IncrementalEncoder.__init__(self, errors)
        self.buffer = ""

    def _buffer_encode(self, input, errors, final):
        return (b"", 0)

    def encode(self, input, final=False):
        data = self.buffer + input
        result, consumed = self._buffer_encode(data, self.errors, final)
        self.buffer = data[consumed:]
        return result

    def reset(self):
        self.buffer = ""


class BufferedIncrementalDecoder(IncrementalDecoder):
    def __init__(self, errors="strict"):
        IncrementalDecoder.__init__(self, errors)
        self.buffer = b""

    def _buffer_decode(self, input, errors, final):
        return ("", 0)

    def decode(self, input, final=False):
        data = self.buffer + bytes(input)
        result, consumed = self._buffer_decode(data, self.errors, final)
        self.buffer = data[consumed:]
        return result


class _UTF8IncrementalDecoder(BufferedIncrementalDecoder):
    def _buffer_decode(self, input, errors, final):
        return utf_8_decode(input, errors, final)


class StreamWriter(Codec):
    def __init__(self, stream, errors="strict"):
        self.stream = stream
        self.errors = errors
        self.encoding = "utf-8"

    def write(self, object):
        self.stream.write(object.encode(self.encoding, self.errors))

    def writelines(self, list):
        self.write("".join(list))

    def reset(self):
        pass

    def __enter__(self):
        return self

    def __exit__(self, type, value, tb):
        self.stream.close()


class StreamReader(Codec):
    def __init__(self, stream, errors="strict"):
        self.stream = stream
        self.errors = errors
        self.encoding = "utf-8"

    def read(self, size=-1, chars=-1, firstline=False):
        data = self.stream.read() if size < 0 else self.stream.read(size)
        return data.decode(self.encoding, self.errors)

    def readline(self, size=None, keepends=True):
        line = self.stream.readline()
        text = line.decode(self.encoding, self.errors)
        if not keepends:
            text = text.rstrip("\r\n")
        return text

    def readlines(self, sizehint=None, keepends=True):
        return self.read().splitlines(keepends)

    def reset(self):
        pass

    def __iter__(self):
        return iter(self.readlines())

    def __enter__(self):
        return self

    def __exit__(self, type, value, tb):
        self.stream.close()


def iterencode(iterator, encoding, errors="strict", **kwargs):
    for input in iterator:
        output = encode(input, encoding, errors)
        if output:
            yield output


def iterdecode(iterator, encoding, errors="strict", **kwargs):
    decoder = getincrementaldecoder(encoding)(errors)
    for input in iterator:
        output = decoder.decode(input)
        if output:
            yield output
    output = decoder.decode(b"", True)
    if output:
        yield output


def open(filename, mode="r", encoding=None, errors="strict", buffering=-1):
    if encoding is not None and "b" not in mode:
        mode = mode + "b"
    f = builtins_open(filename, mode)
    if encoding is None:
        return f
    info = lookup(encoding)
    return StreamReaderWriter(f, info.streamreader, info.streamwriter, errors, info.name)


def builtins_open(filename, mode):
    from builtins import open as _open
    return _open(filename, mode)


class StreamReaderWriter:
    def __init__(self, stream, Reader, Writer, errors="strict", encoding="utf-8"):
        self.stream = stream
        self.errors = errors
        self.encoding = encoding

    def read(self, size=-1):
        data = self.stream.read() if size < 0 else self.stream.read(size)
        return data.decode(self.encoding, self.errors)

    def readline(self, size=None):
        return self.stream.readline().decode(self.encoding, self.errors)

    def readlines(self, sizehint=None):
        return self.read().splitlines(True)

    def write(self, data):
        return self.stream.write(data.encode(self.encoding, self.errors))

    def writelines(self, list):
        self.write("".join(list))

    def __iter__(self):
        return iter(self.readlines())

    def close(self):
        self.stream.close()

    def __enter__(self):
        return self

    def __exit__(self, type, value, tb):
        self.stream.close()
