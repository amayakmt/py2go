"""warnings: a reduced version. Warnings are printed to stderr once per message."""
import sys

filters = []
_seen = set()
_logs = []
_ignored = ("DeprecationWarning", "PendingDeprecationWarning", "ImportWarning", "ResourceWarning")


class WarningMessage:
    def __init__(self, message, category, filename, lineno, file=None, line=None, source=None):
        self.message = message
        self.category = category
        self.filename = filename
        self.lineno = lineno
        self.file = file
        self.line = line
        self.source = source
        self._category_name = category.__name__ if category else None

    def __str__(self):
        return "{message : %r, category : %r, filename : %r, lineno : %s, line : %r}" % (
            self.message, self._category_name, self.filename, self.lineno, self.line)


def simplefilter(action, category=Warning, lineno=0, append=False):
    filters.insert(0, (action, None, category, None, lineno))


def filterwarnings(action, message="", category=Warning, module="", lineno=0, append=False):
    filters.insert(0, (action, message, category, module, lineno))


def resetwarnings():
    filters.clear()


def _action(text, category):
    for f in filters:
        action = f[0]
        cat = f[2]
        if issubclass(category, cat):
            return action
    if category.__name__ in _ignored:
        return "ignore"
    return "default"


def warn(message, category=None, stacklevel=1, source=None, *, skip_file_prefixes=()):
    if category is None:
        category = UserWarning
    text = str(message)
    action = _action(text, category)
    if action == "ignore":
        return
    if _logs and _logs[-1][0]:
        _logs[-1][1].append(WarningMessage(message, category, "<string>", 1))
        return
    if action == "error":
        raise category(text)
    key = (text, category.__name__)
    if action == "default" and key in _seen:
        return
    _seen.add(key)
    sys.stderr.write("<string>:1: " + category.__name__ + ": " + text + "\n")


def warn_explicit(message, category, filename, lineno, module=None, registry=None, module_globals=None, source=None):
    warn(message, category)


class catch_warnings:
    def __init__(self, *, record=False, module=None, action=None, category=Warning, lineno=0, append=False):
        self._record = record
        self._saved = []

    def __enter__(self):
        self._saved = list(filters)
        log = []
        _logs.append((self._record, log))
        if self._record:
            return log
        return None

    def __exit__(self, exc_type, exc_value, traceback):
        filters[:] = self._saved
        _logs.pop()


def _deprecated(name, message="", *, remove=(3, 14)):
    warn(message % ("{!r}".format(name), "%d.%d" % remove) if "%" in message else message, DeprecationWarning)
