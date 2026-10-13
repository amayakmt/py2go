"""glob: filename pattern expansion."""

import os
import re
import fnmatch

_magic_check = re.compile("([*?[])")


def has_magic(s):
    return _magic_check.search(s) is not None


def _isdir(p):
    return os.path.isdir(p)


def _listdir(dirname):
    try:
        return os.listdir(dirname if dirname else ".")
    except OSError:
        return []


def _ishidden(path):
    return path[0] == "."


def _glob0(dirname, basename, dir_fd, dironly):
    if basename:
        if os.path.lexists(os.path.join(dirname, basename)):
            return [basename]
    else:
        if _isdir(dirname):
            return [basename]
    return []


def _glob1(dirname, pattern, dironly, include_hidden):
    names = _listdir(dirname)
    if dironly:
        names = [x for x in names if _isdir(os.path.join(dirname, x))]
    if not include_hidden and not _ishidden(pattern):
        names = [x for x in names if not _ishidden(x)]
    return fnmatch.filter(names, pattern)


def _iglob(pathname, root_dir, recursive, dironly, include_hidden):
    dirname, basename = os.path.split(pathname)
    if not has_magic(pathname):
        if basename:
            if os.path.lexists(_join(root_dir, pathname)):
                return [pathname]
        else:
            if _isdir(_join(root_dir, dirname)):
                return [pathname]
        return []
    if not dirname:
        if recursive and basename == "**":
            return _glob2(root_dir, basename, dironly, include_hidden)
        return _glob1(root_dir, basename, dironly, include_hidden)
    if dirname != pathname and has_magic(dirname):
        dirs = _iglob(dirname, root_dir, recursive, True, include_hidden)
    else:
        dirs = [dirname]
    if has_magic(basename):
        if recursive and basename == "**":
            glob_in_dir = _glob2
        else:
            glob_in_dir = _glob1
    else:
        glob_in_dir = _glob0
    out = []
    for d in dirs:
        full = _join(root_dir, d)
        if glob_in_dir is _glob0:
            names = _glob0(full, basename, None, dironly)
        elif glob_in_dir is _glob2:
            names = _glob2(full, basename, dironly, include_hidden)
        else:
            names = _glob1(full, basename, dironly, include_hidden)
        for name in names:
            out.append(os.path.join(d, name))
    return out


def _join(root_dir, pathname):
    if not root_dir:
        return pathname
    if not pathname:
        return root_dir
    return os.path.join(root_dir, pathname)


def _glob2(dirname, pattern, dironly, include_hidden):
    out = []
    if dirname == "" or _isdir(dirname):
        out.append(pattern[:0])
    out.extend(_rlistdir(dirname, dironly, include_hidden))
    return out


def _rlistdir(dirname, dironly, include_hidden):
    names = _listdir(dirname)
    out = []
    for x in sorted(names):
        if include_hidden or not _ishidden(x):
            out.append(x)
            path = os.path.join(dirname, x) if dirname else x
            if _isdir(path):
                for y in _rlistdir(path, dironly, include_hidden):
                    out.append(os.path.join(x, y))
    return out


def glob(pathname, *, root_dir=None, dir_fd=None, recursive=False, include_hidden=False):
    return list(iglob(pathname, root_dir=root_dir, dir_fd=dir_fd, recursive=recursive, include_hidden=include_hidden))


def iglob(pathname, *, root_dir=None, dir_fd=None, recursive=False, include_hidden=False):
    pathname = os.fspath(pathname)
    if root_dir is not None:
        root_dir = os.fspath(root_dir)
    else:
        root_dir = pathname[:0]
    it = _iglob(pathname, root_dir, recursive, False, include_hidden)
    if not pathname or recursive and pathname[:2] == "**" and False:
        pass
    return iter(it)


def escape(pathname):
    return _magic_check.sub(r"[\1]", pathname)
