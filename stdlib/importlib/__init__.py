"""importlib: imports by name when the program runs.

A py2go program is linked when it is compiled, so import_module finds the
modules the program links: the ones it imports anywhere, and the ones named
by a literal string in a call of import_module, which the linker links too.
Any other name raises ModuleNotFoundError, as for a module that is not
installed."""

import sys


def _resolve_name(name, package, level):
    """Resolve a relative module name to an absolute one."""
    bits = package.rsplit('.', level - 1)
    if len(bits) < level:
        raise ImportError('attempted relative import beyond top-level package')
    base = bits[0]
    return f'{base}.{name}' if name else base


def import_module(name, package=None):
    """Import a module.

    The 'package' argument is required when performing a relative import. It
    specifies the package to use as the anchor point from which to resolve the
    relative import to an absolute import.
    """
    level = 0
    if name.startswith('.'):
        if not package:
            raise TypeError("the 'package' argument is required to perform a "
                            f"relative import for {name!r}")
        for character in name:
            if character != '.':
                break
            level += 1
        name = _resolve_name(name[level:], package, level)
    return _native("module.import", name)


def reload(module):
    """Reload the module and return it. A linked module is not read again:
    the module is returned as it is."""
    name = getattr(module, "__name__", None)
    if name is None or not _native("module.linked", name):
        raise TypeError("reload() argument must be a module")
    return module


def invalidate_caches():
    """Call the invalidate_caches() method on all meta path finders."""
    pass
