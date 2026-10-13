"""csv: reading and writing delimited text."""

import re
from io import StringIO

from builtins import csv_pm_Error as Error

QUOTE_MINIMAL = 0
QUOTE_ALL = 1
QUOTE_NONNUMERIC = 2
QUOTE_NONE = 3
QUOTE_STRINGS = 4
QUOTE_NOTNULL = 5

__version__ = "1.0"


def _obj(x):
    return x


class _State:
    def __init__(self):
        self.limit = _obj(131072)
        self.dialects = {}


_st = _State()


def field_size_limit(*args):
    old = _st.limit
    if len(args) > 1:
        raise TypeError("field_size_limit expected at most 1 argument, got " + str(len(args)))
    if len(args) == 1:
        if not isinstance(args[0], int):
            raise TypeError("limit must be an integer")
        _st.limit = args[0]
    return old


class Dialect:
    _name = ""
    _valid = False
    delimiter = ""
    quotechar = ""
    escapechar = ""
    doublequote = False
    skipinitialspace = False
    lineterminator = ""
    quoting = 0
    strict = False

    def __init__(self):
        if self.__class__ != Dialect:
            self._valid = True
        self._validate()

    def _validate(self):
        if self._valid:
            _build(self, {})


class excel(Dialect):
    delimiter = ","
    quotechar = '"'
    doublequote = True
    skipinitialspace = False
    lineterminator = "\r\n"
    quoting = QUOTE_MINIMAL


class excel_tab(excel):
    delimiter = "\t"


class unix_dialect(Dialect):
    delimiter = ","
    quotechar = '"'
    doublequote = True
    skipinitialspace = False
    lineterminator = "\n"
    quoting = QUOTE_ALL


class _Params:
    def __init__(self, delimiter, quotechar, escapechar, doublequote, skipinitialspace, lineterminator, quoting, strict):
        self.delimiter = delimiter
        self.quotechar = quotechar
        self.escapechar = escapechar
        self.doublequote = doublequote
        self.skipinitialspace = skipinitialspace
        self.lineterminator = lineterminator
        self.quoting = quoting
        self.strict = strict


def _char(name, value, default, allow_none):
    if value is None:
        return None
    if not isinstance(value, str):
        raise TypeError('"' + name + '" must be string or None, not ' + type(value).__name__)
    if len(value) != 1:
        if len(value) == 0 and allow_none:
            return None
        raise TypeError('"' + name + '" must be a 1-character string')
    return value


def _build(dialect, kwargs):
    base = None
    if isinstance(dialect, str):
        if dialect not in _st.dialects:
            raise Error("unknown dialect")
        base = _st.dialects[dialect]
    elif dialect is not None:
        base = dialect
    def pick(name, default):
        if name in kwargs:
            return kwargs[name]
        if base is not None and hasattr(base, name):
            return getattr(base, name)
        return default
    delimiter = pick("delimiter", ",")
    quotechar = pick("quotechar", '"')
    escapechar = pick("escapechar", None)
    doublequote = pick("doublequote", True)
    skipinitialspace = pick("skipinitialspace", False)
    lineterminator = pick("lineterminator", "\r\n")
    quoting = pick("quoting", QUOTE_MINIMAL)
    strict = pick("strict", False)
    for key in kwargs:
        if key not in ("delimiter", "quotechar", "escapechar", "doublequote", "skipinitialspace", "lineterminator", "quoting", "strict"):
            raise TypeError("this function got an unexpected keyword argument '" + key + "'")
    if delimiter is None or not isinstance(delimiter, str):
        raise TypeError('"delimiter" must be string, not ' + type(delimiter).__name__)
    if len(delimiter) != 1:
        raise TypeError('"delimiter" must be a 1-character string')
    quotechar = _char("quotechar", quotechar, '"', True)
    escapechar = _char("escapechar", escapechar, None, True)
    if not isinstance(lineterminator, str):
        raise TypeError('"lineterminator" must be a string')
    if not isinstance(quoting, int):
        raise TypeError('"quoting" must be an integer')
    if quoting < 0 or quoting > 5:
        raise TypeError('bad "quoting" value')
    if quotechar is None and quoting != QUOTE_NONE and "quotechar" in kwargs:
        raise TypeError("quotechar must be set if quoting enabled")
    if quotechar is None and quoting != QUOTE_NONE:
        raise TypeError("quotechar must be set if quoting enabled")
    return _Params(delimiter, quotechar, escapechar, bool(doublequote), bool(skipinitialspace), lineterminator, quoting, bool(strict))


def register_dialect(name, dialect=None, **kwargs):
    if not isinstance(name, str):
        raise TypeError("dialect name must be a string")
    d = _build(dialect, kwargs)
    _st.dialects[name] = d


def unregister_dialect(name):
    if name not in _st.dialects:
        raise Error("unknown dialect")
    del _st.dialects[name]


def get_dialect(name):
    if name not in _st.dialects:
        raise Error("unknown dialect")
    return _st.dialects[name]


def list_dialects():
    return list(_st.dialects.keys())


_START_RECORD = 0
_START_FIELD = 1
_ESCAPED_CHAR = 2
_IN_FIELD = 3
_IN_QUOTED_FIELD = 4
_ESCAPE_IN_QUOTED_FIELD = 5
_QUOTE_IN_QUOTED_FIELD = 6
_EAT_CRNL = 7
_AFTER_ESCAPED_CRNL = 8


class Reader:
    def __init__(self, iterable, dialect="excel", **kwargs):
        self._it = iter(iterable)
        self.dialect = _build(dialect, kwargs)
        self.line_num = 0
        self._fields = []
        self._field = []
        self._state = _START_RECORD
        self._numeric = False
        self._quoted = False

    def __iter__(self):
        return self

    def _save_field(self):
        text = "".join(self._field)
        self._field = []
        d = self.dialect
        if self._numeric:
            self._numeric = False
            self._fields.append(float(text))
        elif d.quoting == QUOTE_NOTNULL and text == "" and not self._quoted:
            self._fields.append(None)
        elif d.quoting == QUOTE_STRINGS and not self._quoted:
            if text == "":
                self._fields.append(None)
            else:
                self._fields.append(float(text))
        else:
            self._fields.append(text)
        self._quoted = False

    def _add(self, c):
        if len(self._field) >= _st.limit:
            raise Error("field larger than field limit (" + str(_st.limit) + ")")
        self._field.append(c)

    def _process(self, c):
        d = self.dialect
        s = self._state
        if s == _START_RECORD:
            if c is None:
                return
            elif c == "\n" or c == "\r":
                self._state = _EAT_CRNL
                return
            s = _START_FIELD
            self._state = s
        if s == _START_FIELD:
            if c == "\n" or c == "\r" or c is None:
                self._save_field()
                self._state = _START_RECORD if c is None else _EAT_CRNL
            elif c == d.quotechar and d.quoting != QUOTE_NONE:
                self._quoted = True
                self._state = _IN_QUOTED_FIELD
            elif c == d.escapechar:
                self._state = _ESCAPED_CHAR
            elif c == " " and d.skipinitialspace:
                pass
            elif c == d.delimiter:
                self._save_field()
            else:
                if d.quoting == QUOTE_NONNUMERIC:
                    self._numeric = True
                self._add(c)
                self._state = _IN_FIELD
        elif s == _ESCAPED_CHAR:
            if c == "\n" or c == "\r":
                self._add(c)
                self._state = _AFTER_ESCAPED_CRNL
                return
            if c is None:
                c = "\n"
            self._add(c)
            self._state = _IN_FIELD
        elif s == _AFTER_ESCAPED_CRNL:
            if c is None:
                return
            self._state = _IN_FIELD
            self._process(c)
        elif s == _IN_FIELD:
            if c == "\n" or c == "\r" or c is None:
                self._save_field()
                self._state = _START_RECORD if c is None else _EAT_CRNL
            elif c == d.escapechar:
                self._state = _ESCAPED_CHAR
            elif c == d.delimiter:
                self._save_field()
                self._state = _START_FIELD
            else:
                self._add(c)
        elif s == _IN_QUOTED_FIELD:
            if c is None:
                pass
            elif c == d.escapechar:
                self._state = _ESCAPE_IN_QUOTED_FIELD
            elif c == d.quotechar and d.quoting != QUOTE_NONE:
                if d.doublequote:
                    self._state = _QUOTE_IN_QUOTED_FIELD
                else:
                    self._state = _IN_FIELD
            else:
                self._add(c)
        elif s == _ESCAPE_IN_QUOTED_FIELD:
            if c is None:
                c = "\n"
            self._add(c)
            self._state = _IN_QUOTED_FIELD
        elif s == _QUOTE_IN_QUOTED_FIELD:
            if d.quoting != QUOTE_NONE and c == d.quotechar:
                self._add(c)
                self._state = _IN_QUOTED_FIELD
            elif c == d.delimiter:
                self._save_field()
                self._state = _START_FIELD
            elif c == "\n" or c == "\r" or c is None:
                self._save_field()
                self._state = _START_RECORD if c is None else _EAT_CRNL
            elif not d.strict:
                self._add(c)
                self._state = _IN_FIELD
            else:
                raise Error("'" + d.delimiter + "' expected after '" + d.quotechar + "'")
        elif s == _EAT_CRNL:
            if c == "\n" or c == "\r":
                pass
            elif c is None:
                self._state = _START_RECORD
            else:
                raise Error("new-line character seen in unquoted field - do you need to open the file with newline=''?")

    def __next__(self):
        self._fields = []
        self._field = []
        self._state = _START_RECORD
        self._numeric = False
        self._quoted = False
        while True:
            try:
                line = next(self._it)
            except StopIteration:
                if len(self._field) != 0 or self._state == _IN_QUOTED_FIELD:
                    if self.dialect.strict:
                        raise Error("unexpected end of data")
                    self._save_field()
                    break
                raise
            if not isinstance(line, str):
                raise Error("iterator should return strings, not " + type(line).__name__ + " (the file should be opened in text mode)")
            self.line_num += 1
            for c in line:
                self._process(c)
            self._process(None)
            if self._state == _START_RECORD:
                break
        fields = self._fields
        self._fields = []
        return fields


class Writer:
    def __init__(self, fileobj, dialect="excel", **kwargs):
        if not hasattr(fileobj, "write"):
            raise TypeError("argument 1 must have a \"write\" method")
        self._write = fileobj.write
        self.dialect = _build(dialect, kwargs)

    def _is_number(self, v):
        return isinstance(v, (int, float, complex)) and not isinstance(v, bool) or isinstance(v, bool)

    def writerow(self, row):
        try:
            it = iter(row)
        except TypeError:
            raise Error("iterable expected, not " + type(row).__name__) from None
        d = self.dialect
        parts = []
        n = 0
        for field in it:
            n += 1
            quoted = False
            if d.quoting == QUOTE_NONNUMERIC:
                quoted = not self._is_number(field)
            elif d.quoting == QUOTE_ALL:
                quoted = True
            elif d.quoting == QUOTE_STRINGS:
                quoted = isinstance(field, str)
            elif d.quoting == QUOTE_NOTNULL:
                quoted = field is not None
            if field is None:
                text = ""
            elif isinstance(field, str):
                text = field
            elif isinstance(field, float):
                text = repr(field)
            else:
                text = str(field)
            out = []
            for c in text:
                want_escape = False
                if c == d.delimiter or c == d.escapechar or c == d.quotechar or c in d.lineterminator:
                    if d.quoting == QUOTE_NONE:
                        want_escape = True
                    else:
                        if c == d.quotechar:
                            if d.doublequote:
                                out.append(d.quotechar)
                            else:
                                want_escape = True
                        elif c == d.escapechar:
                            want_escape = True
                        if not want_escape:
                            quoted = True
                    if want_escape:
                        if d.escapechar is None:
                            raise Error("need to escape, but no escapechar set")
                        out.append(d.escapechar)
                out.append(c)
            body = "".join(out)
            if quoted:
                body = d.quotechar + body + d.quotechar
            parts.append(body)
        if n == 1 and parts[0] == "":
            if d.quoting == QUOTE_NONE:
                raise Error("single empty field record must be quoted")
            parts[0] = d.quotechar + d.quotechar
        line = d.delimiter.join(parts) + d.lineterminator
        return self._write(line)

    def writerows(self, rows):
        for row in rows:
            self.writerow(row)


def reader(csvfile, dialect="excel", **fmtparams):
    return Reader(csvfile, dialect, **fmtparams)


def writer(csvfile, dialect="excel", **fmtparams):
    return Writer(csvfile, dialect, **fmtparams)


register_dialect("excel", excel)
register_dialect("excel-tab", excel_tab)
register_dialect("unix", unix_dialect)


class DictReader:
    def __init__(self, f, fieldnames=None, restkey=None, restval=None, dialect="excel", *args, **kwds):
        self._fieldnames = fieldnames
        self.restkey = restkey
        self.restval = restval
        self.reader = reader(f, dialect, *args, **kwds)
        self.dialect = dialect
        self.line_num = 0

    def __iter__(self):
        return self

    @property
    def fieldnames(self):
        if self._fieldnames is None:
            try:
                self._fieldnames = next(self.reader)
            except StopIteration:
                pass
        self.line_num = self.reader.line_num
        return self._fieldnames

    @fieldnames.setter
    def fieldnames(self, value):
        self._fieldnames = value

    def __next__(self):
        if self.line_num == 0:
            self.fieldnames
        row = next(self.reader)
        self.line_num = self.reader.line_num
        while row == []:
            row = next(self.reader)
        names = self.fieldnames
        d = dict(zip(names, row))
        lf = len(names)
        lr = len(row)
        if lf < lr:
            d[self.restkey] = row[lf:]
        elif lf > lr:
            for key in names[lr:]:
                d[key] = self.restval
        return d


class DictWriter:
    def __init__(self, f, fieldnames, restval="", extrasaction="raise", dialect="excel", *args, **kwds):
        if extrasaction.lower() not in ("raise", "ignore"):
            raise ValueError("extrasaction (" + extrasaction + ") must be 'raise' or 'ignore'")
        self.fieldnames = fieldnames
        self.restval = restval
        self.extrasaction = extrasaction
        self.writer = writer(f, dialect, *args, **kwds)

    def writeheader(self):
        header = dict(zip(self.fieldnames, self.fieldnames))
        return self.writerow(header)

    def _dict_to_list(self, rowdict):
        if self.extrasaction == "raise":
            wrong_fields = [k for k in rowdict if k not in self.fieldnames]
            if wrong_fields:
                raise ValueError("dict contains fields not in fieldnames: " + ", ".join([repr(x) for x in wrong_fields]))
        return [rowdict.get(key, self.restval) for key in self.fieldnames]

    def writerow(self, rowdict):
        return self.writer.writerow(self._dict_to_list(rowdict))

    def writerows(self, rowdicts):
        return self.writer.writerows(map(self._dict_to_list, rowdicts))
