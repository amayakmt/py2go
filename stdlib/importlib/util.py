"""importlib.util: helpers for importers (find_spec and resolve_name; the
modules of a py2go program are linked in, so there is nothing to load)."""

from importlib import _resolve_name
from importlib.machinery import ModuleSpec


def resolve_name(name, package):
    """Resolve a relative module name to an absolute one."""
    if not name.startswith('.'):
        return name
    elif not package:
        raise ImportError(f'no package specified for {repr(name)} '
                          '(required for relative module names)')
    level = 0
    for character in name:
        if character != '.':
            break
        level += 1
    return _resolve_name(name[level:], package, level)


def find_spec(name, package=None):
    """Return the spec of the module, or None when the program does not link
    it."""
    fullname = resolve_name(name, package) if name.startswith('.') else name
    if not _native("module.linked", fullname):
        return None
    return ModuleSpec(fullname, None, origin="py2go")


def module_from_spec(spec):
    raise ImportError("py2go links its modules when the program is compiled; "
                      "a module cannot be made from a spec")


def spec_from_file_location(name, location=None, *, loader=None, submodule_search_locations=None):
    raise ImportError("py2go links its modules when the program is compiled; "
                      "a module cannot be loaded from a file")
