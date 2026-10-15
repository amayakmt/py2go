"""Policy framework for the email package (attributes are set in the constructor)."""
from email import header
from email import charset as _charset
from email.utils import _has_surrogates

__all__ = ["Policy", "Compat32", "compat32"]


class Policy:
    _names = ("raise_on_defect", "linesep", "cte_type", "max_line_length", "mangle_from_", "message_factory",
              "verify_generated_headers")

    def __init__(self, **kw):
        self.raise_on_defect = False
        self.linesep = "\n"
        self.cte_type = "8bit"
        self.max_line_length = 78
        self.mangle_from_ = self._default_mangle()
        self.message_factory = None
        self.verify_generated_headers = True
        self._init_extra()
        for name, value in kw.items():
            if name not in self._all_names():
                raise TypeError("{!r} is an invalid keyword argument for {}".format(name, self.__class__.__name__))
            setattr(self, name, value)

    def _default_mangle(self):
        return False

    def _init_extra(self):
        pass

    def _all_names(self):
        return self._names

    def __repr__(self):
        args = ["{}={!r}".format(name, getattr(self, name)) for name in self._all_names()
                if getattr(self, name) != self._fresh().__dict__.get(name)]
        return "{}({})".format(self.__class__.__name__, ", ".join(args))

    def _fresh(self):
        return self.__class__()

    def clone(self, **kw):
        newpolicy = self.__class__()
        for attr in self._all_names():
            setattr(newpolicy, attr, getattr(self, attr))
        for attr, value in kw.items():
            if attr not in self._all_names():
                raise TypeError("{!r} is an invalid keyword argument for {}".format(attr, self.__class__.__name__))
            setattr(newpolicy, attr, value)
        return newpolicy

    def __add__(self, other):
        base = self._fresh()
        kw = {}
        for attr in other._all_names():
            v = getattr(other, attr)
            if v != getattr(base, attr, None):
                kw[attr] = v
        return self.clone(**kw)

    def handle_defect(self, obj, defect):
        if self.raise_on_defect:
            raise defect
        self.register_defect(obj, defect)

    def register_defect(self, obj, defect):
        obj.defects.append(defect)

    def header_max_count(self, name):
        return None

    def header_source_parse(self, sourcelines):
        raise NotImplementedError

    def header_store_parse(self, name, value):
        raise NotImplementedError

    def header_fetch_parse(self, name, value):
        raise NotImplementedError

    def fold(self, name, value):
        raise NotImplementedError

    def fold_binary(self, name, value):
        raise NotImplementedError


class Compat32(Policy):
    def _default_mangle(self):
        return True

    def _sanitize_header(self, name, value):
        if not isinstance(value, str):
            return value
        if _has_surrogates(value):
            return header.Header(value, charset=_charset.UNKNOWN8BIT, header_name=name)
        else:
            return value

    def header_source_parse(self, sourcelines):
        name, value = sourcelines[0].split(":", 1)
        value = "".join((value, *sourcelines[1:])).lstrip(" \t\r\n")
        return (name, value.rstrip("\r\n"))

    def header_store_parse(self, name, value):
        return (name, value)

    def header_fetch_parse(self, name, value):
        return self._sanitize_header(name, value)

    def fold(self, name, value):
        return self._fold(name, value, sanitize=True)

    def fold_binary(self, name, value):
        folded = self._fold(name, value, sanitize=self.cte_type == "7bit")
        return folded.encode("ascii", "surrogateescape")

    def _fold(self, name, value, sanitize):
        parts = []
        parts.append("%s: " % name)
        if isinstance(value, str):
            if _has_surrogates(value):
                if sanitize:
                    h = header.Header(value, charset=_charset.UNKNOWN8BIT, header_name=name)
                else:
                    parts.append(value)
                    h = None
            else:
                h = header.Header(value, header_name=name)
        else:
            h = value
        if h is not None:
            maxlinelen = 0
            if self.max_line_length is not None:
                maxlinelen = self.max_line_length
            parts.append(h.encode(linesep=self.linesep, maxlinelen=maxlinelen))
        parts.append(self.linesep)
        return "".join(parts)


compat32 = Compat32()
