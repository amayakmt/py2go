"""json.encoder"""

INFINITY = float("inf")


def encode_basestring_ascii(s):
    return _native("json.escape", s, True)


def encode_basestring(s):
    return _native("json.escape", s, False)


class JSONEncoder:
    item_separator = ", "
    key_separator = ": "

    def __init__(self, *, skipkeys=False, ensure_ascii=True, check_circular=True, allow_nan=True, sort_keys=False, indent=None, separators=None, default=None):
        self.skipkeys = skipkeys
        self.ensure_ascii = ensure_ascii
        self.check_circular = check_circular
        self.allow_nan = allow_nan
        self.sort_keys = sort_keys
        self.indent = indent
        if separators is not None:
            self.item_separator, self.key_separator = separators
        elif indent is not None:
            self.item_separator = ","
        self._default = default

    def default(self, o):
        raise TypeError("Object of type " + type(o).__name__ + " is not JSON serializable")

    def encode(self, o):
        if isinstance(o, str):
            return _native("json.escape", o, self.ensure_ascii)
        return "".join(self.iterencode(o))

    def iterencode(self, o, _one_shot=False):
        out = []
        self._encode(o, out, 0, [])
        return out

    def _float(self, o):
        if o != o:
            text = "NaN"
        elif o == INFINITY:
            text = "Infinity"
        elif o == -INFINITY:
            text = "-Infinity"
        else:
            return repr(o)
        if not self.allow_nan:
            raise ValueError("Out of range float values are not JSON compliant: " + repr(o))
        return text

    def _key(self, k):
        if isinstance(k, str):
            return k
        if k is True:
            return "true"
        if k is False:
            return "false"
        if k is None:
            return "null"
        if isinstance(k, float):
            return self._float(k)
        if isinstance(k, int):
            return repr(k)
        return None

    def _newline(self, level):
        if self.indent is None:
            return ""
        if isinstance(self.indent, str):
            return "\n" + self.indent * level
        return "\n" + " " * (self.indent * level)

    def _encode(self, o, out, level, markers):
        if isinstance(o, str):
            out.append(_native("json.escape", o, self.ensure_ascii))
        elif o is None:
            out.append("null")
        elif o is True:
            out.append("true")
        elif o is False:
            out.append("false")
        elif isinstance(o, int):
            out.append(repr(o))
        elif isinstance(o, float):
            out.append(self._float(o))
        elif isinstance(o, (list, tuple)):
            self._list(o, out, level, markers)
        elif isinstance(o, dict):
            self._dict(o, out, level, markers)
        else:
            if self.check_circular:
                mark = id(o)
                if mark in markers:
                    raise ValueError("Circular reference detected")
                markers.append(mark)
            if self._default is not None:
                replaced = self._default(o)
            else:
                replaced = self.default(o)
            self._encode(replaced, out, level, markers)
            if self.check_circular:
                markers.pop()

    def _list(self, o, out, level, markers):
        if not o:
            out.append("[]")
            return
        if self.check_circular:
            mark = id(o)
            if mark in markers:
                raise ValueError("Circular reference detected")
            markers.append(mark)
        out.append("[")
        first = True
        for item in o:
            if not first:
                out.append(self.item_separator)
            first = False
            if self.indent is not None:
                out.append(self._newline(level + 1))
            self._encode(item, out, level + 1, markers)
        if self.indent is not None:
            out.append(self._newline(level))
        out.append("]")
        if self.check_circular:
            markers.pop()

    def _dict(self, o, out, level, markers):
        if not o:
            out.append("{}")
            return
        if self.check_circular:
            mark = id(o)
            if mark in markers:
                raise ValueError("Circular reference detected")
            markers.append(mark)
        out.append("{")
        items = list(o.items())
        if self.sort_keys:
            items = sorted(items, key=lambda kv: kv[0])
        first = True
        for k, v in items:
            key = self._key(k)
            if key is None:
                if self.skipkeys:
                    continue
                raise TypeError("keys must be str, int, float, bool or None, not " + type(k).__name__)
            if not first:
                out.append(self.item_separator)
            first = False
            if self.indent is not None:
                out.append(self._newline(level + 1))
            out.append(_native("json.escape", key, self.ensure_ascii))
            out.append(self.key_separator)
            self._encode(v, out, level + 1, markers)
        if self.indent is not None:
            out.append(self._newline(level))
        out.append("}")
        if self.check_circular:
            markers.pop()
