"""Cache lines from Python source files (reads the file; no tokenize)."""
import os

__all__ = ['getline', 'clearcache', 'checkcache', 'lazycache', 'getlines']

cache = {}


def _read(filename):
    try:
        with open(filename, 'r', encoding='utf-8', errors='replace') as fp:
            lines = fp.readlines()
    except OSError:
        return []
    if lines and not lines[-1].endswith('\n'):
        lines[-1] += '\n'
    return lines


def getlines(filename, module_globals=None):
    if filename in cache:
        return cache[filename]
    lines = _read(filename)
    if lines:
        cache[filename] = lines
    return lines


def getline(filename, lineno, module_globals=None):
    lines = getlines(filename, module_globals)
    if 1 <= lineno <= len(lines):
        return lines[lineno - 1]
    return ''


def clearcache():
    cache.clear()


def checkcache(filename=None):
    if filename is None:
        cache.clear()
    elif filename in cache:
        del cache[filename]


def lazycache(filename, module_globals):
    return False
