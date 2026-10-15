"""traceback: extract and format tracebacks (a reduced version: the runtime keeps no frames)."""
import sys

__all__ = ["format_exc", "print_exc", "print_exception", "format_exception", "format_exception_only",
           "print_tb", "format_tb", "extract_tb", "print_stack", "format_stack", "extract_stack", "format_list",
           "print_last", "clear_frames", "walk_tb", "walk_stack"]


def format_exception_only(exc, value=None, *, show_group=False):
    if value is None and not _native("pickle.is_class", exc):
        value = exc
        exc = type(value)
    if value is None:
        return ["NoneType: None\n"]
    name = type(value).__name__
    text = str(value)
    if text:
        return [name + ": " + text + "\n"]
    return [name + "\n"]


def format_exception(exc, value=None, tb=None, limit=None, chain=True):
    if value is None and not _native("pickle.is_class", exc):
        value = exc
    if value is None:
        return ["NoneType: None\n"]
    lines = []
    if chain:
        cause = getattr(value, "__cause__", None)
        context = getattr(value, "__context__", None)
        if cause is not None:
            lines.extend(format_exception(cause, cause, None, limit, chain))
            lines.append("\nThe above exception was the direct cause of the following exception:\n\n")
        elif context is not None and not getattr(value, "__suppress_context__", False):
            lines.extend(format_exception(context, context, None, limit, chain))
            lines.append("\nDuring handling of the above exception, another exception occurred:\n\n")
    lines.append("Traceback (most recent call last):\n")
    lines.extend(format_exception_only(type(value), value))
    return lines


def print_exception(exc, value=None, tb=None, limit=None, file=None, chain=True):
    if file is None:
        file = sys.stderr
    for line in format_exception(exc, value, tb, limit, chain):
        file.write(line)


def format_exc(limit=None, chain=True):
    info = sys.exc_info()
    if info[1] is None:
        return "NoneType: None\n"
    return "".join(format_exception(info[1], info[1], None, limit, chain))


def print_exc(limit=None, file=None, chain=True):
    if file is None:
        file = sys.stderr
    file.write(format_exc(limit, chain))


def print_last(limit=None, file=None, chain=True):
    raise ValueError("no last exception")


def extract_tb(tb, limit=None):
    return []


def format_tb(tb, limit=None):
    return []


def print_tb(tb, limit=None, file=None):
    pass


def extract_stack(f=None, limit=None):
    return []


def format_stack(f=None, limit=None):
    return []


def print_stack(f=None, limit=None, file=None):
    pass


def format_list(extracted_list):
    return []


def clear_frames(tb):
    pass


def walk_tb(tb):
    return iter(())


def walk_stack(f):
    return iter(())
