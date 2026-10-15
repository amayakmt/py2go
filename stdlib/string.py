"""string: constants and a few helpers."""

whitespace = " \t\n\r\x0b\x0c"
ascii_lowercase = "abcdefghijklmnopqrstuvwxyz"
ascii_uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
ascii_letters = ascii_lowercase + ascii_uppercase
digits = "0123456789"
hexdigits = digits + "abcdef" + "ABCDEF"
octdigits = "01234567"
punctuation = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
printable = digits + ascii_letters + punctuation + whitespace


def capwords(s, sep=None):
    return (sep or " ").join(x.capitalize() for x in s.split(sep))


import re as _re

_PATTERN = _re.compile(r"\$(?:(?P<escaped>\$)|(?P<named>[_a-zA-Z][_a-zA-Z0-9]*)|\{(?P<braced>[_a-zA-Z][_a-zA-Z0-9]*)\}|(?P<invalid>))")


class Template:
    delimiter = "$"
    idpattern = r"(?a:[_a-z][_a-z0-9]*)"
    braceidpattern = None
    flags = _re.IGNORECASE

    def __init__(self, template):
        self.template = template

    def _invalid(self, mo):
        i = mo.start("invalid")
        lines = self.template[:i].splitlines(keepends=True)
        if not lines:
            colno = 1
            lineno = 1
        else:
            colno = i - len("".join(lines[:-1]))
            lineno = len(lines)
        raise ValueError("Invalid placeholder in string: line %d, col %d" % (lineno, colno))

    def substitute(self, mapping={}, /, **kws):
        if kws:
            m = dict(mapping)
            m.update(kws)
            mapping = m

        def convert(mo):
            named = mo.group("named")
            if named is None:
                named = mo.group("braced")
            if named is not None:
                return str(mapping[named])
            if mo.group("escaped") is not None:
                return "$"
            if mo.group("invalid") is not None:
                self._invalid(mo)
            raise ValueError("Unrecognized named group in pattern")

        return _PATTERN.sub(convert, self.template)

    def safe_substitute(self, mapping={}, /, **kws):
        if kws:
            m = dict(mapping)
            m.update(kws)
            mapping = m

        def convert(mo):
            named = mo.group("named")
            if named is None:
                named = mo.group("braced")
            if named is not None:
                try:
                    return str(mapping[named])
                except KeyError:
                    return mo.group()
            if mo.group("escaped") is not None:
                return "$"
            if mo.group("invalid") is not None:
                return mo.group()
            raise ValueError("Unrecognized named group in pattern")

        return _PATTERN.sub(convert, self.template)

    def is_valid(self):
        for mo in _PATTERN.finditer(self.template):
            if mo.group("invalid") is not None:
                return False
            if mo.group("named") is None and mo.group("braced") is None and mo.group("escaped") is None:
                return False
        return True

    def get_identifiers(self):
        ids = []
        for mo in _PATTERN.finditer(self.template):
            named = mo.group("named") or mo.group("braced")
            if named is not None and named not in ids:
                ids.append(named)
            elif named is None and mo.group("escaped") is None and mo.group("invalid") is None:
                raise ValueError("Unrecognized named group in pattern")
        return ids


class Formatter:
    def format(self, format_string, /, *args, **kwargs):
        return format_string.format(*args, **kwargs)

    def vformat(self, format_string, args, kwargs):
        return format_string.format(*args, **kwargs)
