"""array: efficient arrays of numeric values."""

import struct as _struct

typecodes = "bBuwhHiIlLqQfd"

_SIZES = {"b": 1, "B": 1, "u": 4, "w": 4, "h": 2, "H": 2, "i": 4, "I": 4, "l": 8, "L": 8, "q": 8, "Q": 8, "f": 4, "d": 8}
_RANGES = {
    "b": (-128, 127), "B": (0, 255), "h": (-32768, 32767), "H": (0, 65535),
    "i": (-2147483648, 2147483647), "I": (0, 4294967295),
    "l": (-9223372036854775808, 9223372036854775807), "L": (0, 18446744073709551615),
    "q": (-9223372036854775808, 9223372036854775807), "Q": (0, 18446744073709551615),
}
_LOW = {"b": "signed char is less than minimum", "B": "unsigned byte integer is less than minimum", "h": "signed short integer is less than minimum", "H": "unsigned short is less than minimum", "i": "signed integer is less than minimum"}
_HIGH = {"b": "signed char is greater than maximum", "B": "unsigned byte integer is greater than maximum", "h": "signed short integer is greater than maximum", "H": "unsigned short is greater than maximum", "i": "signed integer is greater than maximum", "I": "unsigned int is greater than maximum"}


def _index(v):
    if isinstance(v, int):
        return int(v)
    if isinstance(v, float):
        raise TypeError("'float' object cannot be interpreted as an integer")
    if hasattr(v, "__index__"):
        return v.__index__()
    raise TypeError("'" + type(v).__name__ + "' object cannot be interpreted as an integer")


def _check(tc, v):
    if tc in "bBhHiIlLqQ":
        n = _index(v)
        lo, hi = _RANGES[tc]
        if n < lo:
            if tc == "I" or tc == "L":
                raise OverflowError("can't convert negative value to unsigned int")
            if tc == "Q":
                raise OverflowError("can't convert negative int to unsigned")
            if tc == "q":
                raise OverflowError("int too big to convert")
            if tc == "l":
                raise OverflowError("Python int too large to convert to C long")
            raise OverflowError(_LOW[tc])
        if n > hi:
            if tc == "l":
                raise OverflowError("Python int too large to convert to C long")
            if tc == "L":
                raise OverflowError("Python int too large to convert to C unsigned long" if False else "int too big to convert")
            if tc == "q" or tc == "Q":
                raise OverflowError("int too big to convert")
            raise OverflowError(_HIGH[tc])
        return n
    if tc == "f":
        if isinstance(v, (int, float)):
            return _struct.unpack("f", _struct.pack("f", float(v)))[0]
        raise TypeError("must be real number, not " + type(v).__name__)
    if tc == "d":
        if isinstance(v, (int, float)):
            return float(v)
        raise TypeError("must be real number, not " + type(v).__name__)
    if not isinstance(v, str) or len(v) != 1:
        raise TypeError("array item must be unicode character")
    return v


class array:
    def __init__(self, typecode, initializer=None):
        if not isinstance(typecode, str) or len(typecode) != 1:
            raise TypeError("array() argument 1 must be a unicode character, not " + type(typecode).__name__)
        if typecode not in typecodes:
            raise ValueError("bad typecode (must be b, B, u, w, h, H, i, I, l, L, q, Q, f or d)")
        self.typecode = typecode
        self.itemsize = _SIZES[typecode]
        self._items = []
        if initializer is None:
            return
        if isinstance(initializer, (bytes, bytearray)):
            self.frombytes(bytes(initializer))
        elif isinstance(initializer, str):
            if typecode != "u" and typecode != "w":
                raise TypeError("cannot use a str to initialize an array with typecode '" + typecode + "'")
            for ch in initializer:
                self._items.append(ch)
        elif isinstance(initializer, array):
            if (typecode == "u" or typecode == "w") != (initializer.typecode == "u" or initializer.typecode == "w"):
                raise TypeError("cannot use a unicode array to initialize an array with typecode '" + typecode + "'" if initializer.typecode in "uw" else "cannot use a non-unicode array to initialize a unicode array")
            for v in initializer._items:
                self._items.append(_check(typecode, v))
        else:
            for v in initializer:
                self._items.append(_check(typecode, v))

    def __len__(self):
        return len(self._items)

    def __iter__(self):
        return iter(list(self._items))

    def __contains__(self, v):
        return v in self._items

    def __repr__(self):
        if not self._items:
            return "array('" + self.typecode + "')"
        if self.typecode == "u" or self.typecode == "w":
            return "array('" + self.typecode + "', " + repr("".join(self._items)) + ")"
        return "array('" + self.typecode + "', " + repr(self._items) + ")"

    def _new(self, items):
        a = array(self.typecode)
        a._items = items
        return a

    def __getitem__(self, i):
        if isinstance(i, slice):
            return self._new(self._items[i])
        n = _index(i)
        if n < 0:
            n += len(self._items)
        if n < 0 or n >= len(self._items):
            raise IndexError("array index out of range")
        return self._items[n]

    def __setitem__(self, i, v):
        if isinstance(i, slice):
            if not isinstance(v, array):
                raise TypeError("can only assign array (not \"" + type(v).__name__ + "\") to array slice")
            if v.typecode != self.typecode:
                raise TypeError("bad argument type for built-in operation")
            self._items[i] = list(v._items)
            return
        n = _index(i)
        if n < 0:
            n += len(self._items)
        if n < 0 or n >= len(self._items):
            raise IndexError("array assignment index out of range")
        self._items[n] = _check(self.typecode, v)

    def __delitem__(self, i):
        if isinstance(i, slice):
            del self._items[i]
            return
        n = _index(i)
        if n < 0:
            n += len(self._items)
        if n < 0 or n >= len(self._items):
            raise IndexError("array assignment index out of range")
        del self._items[n]

    def __add__(self, other):
        if not isinstance(other, array):
            raise TypeError("can only append array (not \"" + type(other).__name__ + "\") to array")
        if other.typecode != self.typecode:
            raise TypeError("bad argument type for built-in operation")
        return self._new(self._items + other._items)

    def __iadd__(self, other):
        if not isinstance(other, array):
            raise TypeError("can only extend array with array (not \"" + type(other).__name__ + "\")")
        if other.typecode != self.typecode:
            raise TypeError("bad argument type for built-in operation")
        self._items.extend(other._items)
        return self

    def __mul__(self, n):
        if not isinstance(n, int):
            raise TypeError("can't multiply sequence by non-int of type '" + type(n).__name__ + "'")
        return self._new(self._items * n)

    def __rmul__(self, n):
        return self * n

    def __imul__(self, n):
        self._items = self._items * n
        return self

    def __eq__(self, other):
        if not isinstance(other, array):
            return NotImplemented
        return self._items == other._items

    def __ne__(self, other):
        if not isinstance(other, array):
            return NotImplemented
        return self._items != other._items

    def __lt__(self, other):
        if not isinstance(other, array):
            return NotImplemented
        return self._items < other._items

    def __le__(self, other):
        if not isinstance(other, array):
            return NotImplemented
        return self._items <= other._items

    def __gt__(self, other):
        if not isinstance(other, array):
            return NotImplemented
        return self._items > other._items

    def __ge__(self, other):
        if not isinstance(other, array):
            return NotImplemented
        return self._items >= other._items

    def append(self, v):
        self._items.append(_check(self.typecode, v))

    def extend(self, iterable):
        if isinstance(iterable, array):
            if iterable.typecode != self.typecode:
                raise TypeError("can only extend with array of same kind")
            self._items.extend(iterable._items)
            return
        for v in iterable:
            self._items.append(_check(self.typecode, v))

    def insert(self, i, v):
        n = _index(i)
        self._items.insert(n, _check(self.typecode, v))

    def pop(self, i=-1):
        n = _index(i)
        if n < 0:
            n += len(self._items)
        if n < 0 or n >= len(self._items):
            raise IndexError("pop index out of range")
        return self._items.pop(n)

    def remove(self, v):
        try:
            self._items.remove(v)
        except ValueError:
            raise ValueError("array.remove(x): x not in array") from None

    def index(self, v, start=0, stop=9223372036854775807):
        n = len(self._items)
        lo = start
        if lo < 0:
            lo = max(lo + n, 0)
        hi = stop
        if hi < 0:
            hi += n
        for i in range(lo, min(hi, n)):
            if self._items[i] == v:
                return i
        raise ValueError("array.index(x): x not in array")

    def count(self, v):
        return self._items.count(v)

    def reverse(self):
        self._items.reverse()

    def tolist(self):
        return list(self._items)

    def buffer_info(self):
        return (id(self), len(self._items))

    def tobytes(self):
        tc = self.typecode
        if tc == "u" or tc == "w":
            return "".join(self._items).encode("utf-32-le")
        return _struct.pack("<" + str(len(self._items)) + tc if tc not in "lL" else "<" + str(len(self._items)) + ("q" if tc == "l" else "Q"), *self._items)

    def frombytes(self, b):
        if not isinstance(b, (bytes, bytearray)):
            raise TypeError("a bytes-like object is required, not '" + type(b).__name__ + "'")
        size = self.itemsize
        if len(b) % size != 0:
            raise ValueError("bytes length not a multiple of item size")
        n = len(b) // size
        tc = self.typecode
        if tc == "u" or tc == "w":
            for ch in bytes(b).decode("utf-32-le"):
                self._items.append(ch)
            return
        code = tc
        if tc == "l":
            code = "q"
        elif tc == "L":
            code = "Q"
        for v in _struct.unpack("<" + str(n) + code, bytes(b)):
            self._items.append(v)

    def tounicode(self):
        if self.typecode != "u" and self.typecode != "w":
            raise ValueError("tounicode() may only be called on unicode type arrays ('u' or 'w')")
        return "".join(self._items)

    def fromunicode(self, s):
        if self.typecode != "u" and self.typecode != "w":
            raise ValueError("fromunicode() may only be called on unicode type arrays ('u' or 'w')")
        if not isinstance(s, str):
            raise TypeError("fromunicode() argument must be str, not " + type(s).__name__)
        for ch in s:
            self._items.append(ch)

    def fromlist(self, lst):
        if not isinstance(lst, list):
            raise TypeError("arg must be list")
        checked = [_check(self.typecode, v) for v in lst]
        self._items.extend(checked)

    def tofile(self, f):
        f.write(self.tobytes())

    def fromfile(self, f, n):
        data = f.read(n * self.itemsize)
        if len(data) != n * self.itemsize:
            self.frombytes(data[:len(data) // self.itemsize * self.itemsize])
            raise EOFError("read() didn't return enough bytes")
        self.frombytes(data)

    def byteswap(self):
        if self.typecode == "u" or self.typecode == "w":
            return
        size = self.itemsize
        raw = self.tobytes()
        out = bytearray()
        for i in range(0, len(raw), size):
            out.extend(raw[i:i + size][::-1])
        self._items = []
        self.frombytes(bytes(out))


ArrayType = array
