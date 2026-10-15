"""Policy objects for the email package: compat32 and a reduced EmailPolicy (headers stay str)."""
import re
from email._policybase import Policy, Compat32, compat32
from email.utils import _has_surrogates

__all__ = ["Compat32", "compat32", "Policy", "EmailPolicy", "default", "strict", "SMTP", "HTTP"]

linesep_splitter = re.compile(r"\n|\r\n?")


class EmailPolicy(Compat32):
    def _default_mangle(self):
        return False

    def _init_extra(self):
        self.utf8 = False
        self.refold_source = "long"
        self.header_factory = None
        self.content_manager = None

    def _all_names(self):
        return self._names + ("utf8", "refold_source", "header_factory", "content_manager")

    def header_max_count(self, name):
        return None

    def header_source_parse(self, sourcelines):
        name, value = sourcelines[0].split(":", 1)
        value = value.lstrip(" \t") + "".join(sourcelines[1:])
        return (name, value.rstrip("\r\n"))

    def header_store_parse(self, name, value):
        if isinstance(value, str) and len(value.splitlines()) > 1:
            raise ValueError("Header values may not contain linefeed or carriage return characters")
        return (name, value)

    def header_fetch_parse(self, name, value):
        if isinstance(value, str):
            value = value.replace("\r\n", "").replace("\n", "") if False else value
        return value

    def fold(self, name, value):
        return self._efold(name, value, True)

    def fold_binary(self, name, value):
        folded = self._efold(name, value, self.cte_type == "7bit")
        charset = "utf8" if self.utf8 else "ascii"
        return folded.encode(charset, "surrogateescape")

    def _efold(self, name, value, refold_binary):
        lines = value.splitlines() if isinstance(value, str) else [str(value)]
        maxlen = self.max_line_length if self.max_line_length else 998
        if len(name) + 2 + len(lines[0]) <= maxlen and len(lines) == 1 and not (not self.utf8 and not value.isascii()):
            return name + ": " + lines[0] + self.linesep
        return Compat32._fold(self, name, value, True)


def _make_content_manager():
    from email.contentmanager import raw_data_manager
    return raw_data_manager


default = EmailPolicy()
default.content_manager = _make_content_manager()
strict = default.clone(raise_on_defect=True)
SMTP = default.clone(linesep="\r\n")
HTTP = default.clone(linesep="\r\n", max_line_length=None)
SMTPUTF8 = SMTP.clone(utf8=True)
