"""os.path for POSIX paths (str paths; bytes paths are not supported)."""

sep = "/"
curdir = "."
pardir = ".."
extsep = "."
pathsep = ":"
altsep = None
defpath = "/bin:/usr/bin"
devnull = "/dev/null"

supports_unicode_filenames = False


def _fspath(path):
    if isinstance(path, (str, bytes)):
        return path
    if hasattr(path, "__fspath__"):
        r = path.__fspath__()
        if isinstance(r, (str, bytes)):
            return r
        raise TypeError("expected " + type(path).__name__ + ".__fspath__() to return str or bytes, not " + type(r).__name__)
    raise TypeError("expected str, bytes or os.PathLike object, not " + type(path).__name__)


def _str(path):
    p = _fspath(path)
    if isinstance(p, bytes):
        raise TypeError("bytes paths are not supported")
    return p


def normcase(s):
    return _fspath(s)


def isabs(s):
    return _str(s).startswith("/")


def join(a, *p):
    path = _str(a)
    try:
        for b in p:
            b = _str(b)
            if b.startswith("/"):
                path = b
            elif not path or path.endswith("/"):
                path += b
            else:
                path += "/" + b
    except TypeError:
        for b in (a,) + p:
            if not isinstance(b, str):
                raise TypeError("join() argument must be str, bytes, or os.PathLike object, not '" + type(b).__name__ + "'") from None
        raise
    return path


def split(p):
    p = _str(p)
    i = p.rfind("/") + 1
    head, tail = p[:i], p[i:]
    if head and head != "/" * len(head):
        head = head.rstrip("/")
    return (head, tail)


def basename(p):
    p = _str(p)
    return p[p.rfind("/") + 1:]


def dirname(p):
    p = _str(p)
    i = p.rfind("/") + 1
    head = p[:i]
    if head and head != "/" * len(head):
        head = head.rstrip("/")
    return head


def splitext(p):
    p = _str(p)
    sep_index = p.rfind("/")
    dot_index = p.rfind(".")
    if dot_index > sep_index:
        filename_index = sep_index + 1
        while filename_index < dot_index:
            if p[filename_index] != ".":
                return (p[:dot_index], p[dot_index:])
            filename_index += 1
    return (p, p[:0])


def splitdrive(p):
    p = _str(p)
    return (p[:0], p)


def splitroot(p):
    p = _str(p)
    if not p.startswith("/"):
        return ("", "", p)
    elif p.startswith("//") and not p.startswith("///"):
        return ("", "//", p[2:])
    else:
        return ("", "/", p[1:])


def normpath(path):
    path = _str(path)
    if not path:
        return "."
    initial_slashes = path.startswith("/")
    if initial_slashes and path.startswith("//") and not path.startswith("///"):
        initial_slashes = 2
    comps = path.split("/")
    new_comps = []
    for comp in comps:
        if comp == "" or comp == ".":
            continue
        if comp != ".." or (not initial_slashes and not new_comps) or (new_comps and new_comps[-1] == ".."):
            new_comps.append(comp)
        elif new_comps:
            new_comps.pop()
    path = "/".join(new_comps)
    if initial_slashes:
        path = "/" * int(initial_slashes) + path
    return path or "."


def abspath(path):
    path = _str(path)
    if not path.startswith("/"):
        path = join(_native("posix.getcwd"), path)
    return normpath(path)


def realpath(filename, *, strict=False):
    filename = _str(filename)
    rest = filename.split("/")[::-1]
    part_count = len(rest)
    path = "/" if filename.startswith("/") else _native("posix.getcwd")
    seen = {}
    while part_count:
        name = rest.pop()
        if name is None:
            seen[rest.pop()] = path
            continue
        part_count -= 1
        if not name or name == ".":
            continue
        if name == "..":
            path = path[:path.rindex("/")] or "/"
            continue
        if path == "/":
            newpath = path + name
        else:
            newpath = path + "/" + name
        try:
            st = _native("posix.stat", newpath, False)
            if (st[0] & 0o170000) != 0o120000:
                if strict and part_count and (st[0] & 0o170000) != 0o040000:
                    raise OSError(20, "Not a directory", newpath)
                path = newpath
                continue
            if newpath in seen:
                path = seen[newpath]
                if path is not None:
                    continue
                if strict:
                    _native("posix.stat", newpath, True)
                path = newpath
                continue
            target = _native("posix.readlink", newpath)
        except OSError:
            if strict:
                raise
        else:
            if target.startswith("/"):
                path = "/"
            seen[newpath] = None
            rest.append(newpath)
            rest.append(None)
            target_parts = target.split("/")[::-1]
            rest.extend(target_parts)
            part_count += len(target_parts)
            continue
        if path == "/":
            path = path + name
        else:
            path = path + "/" + name
    return path


def relpath(path, start=None):
    path = _str(path)
    if not path:
        raise ValueError("no path specified")
    if start is None:
        start = "."
    else:
        start = _str(start)
    start_tail = abspath(start).lstrip("/")
    path_tail = abspath(path).lstrip("/")
    start_list = start_tail.split("/") if start_tail else []
    path_list = path_tail.split("/") if path_tail else []
    i = len(commonprefix([start_list, path_list]))
    rel_list = [".."] * (len(start_list) - i) + path_list[i:]
    if not rel_list:
        return "."
    return "/".join(rel_list)


def commonprefix(m):
    if not m:
        return ""
    if not isinstance(m[0], (list, tuple)):
        m = tuple(map(_fspath, m))
    s1 = min(m)
    s2 = max(m)
    for i in range(len(s1)):
        if s1[i] != s2[i]:
            return s1[:i]
    return s1


def commonpath(paths):
    paths = tuple(map(_str, paths))
    if not paths:
        raise ValueError("commonpath() arg is an empty sequence")
    split_paths = [p.split("/") for p in paths]
    absolute = {p.startswith("/") for p in paths}
    if len(absolute) != 1:
        raise ValueError("Can't mix absolute and relative paths")
    isabs_ = paths[0].startswith("/")
    split_paths = [[c for c in s if c and c != "."] for s in split_paths]
    s1 = min(split_paths)
    s2 = max(split_paths)
    common = s1
    for i, c in enumerate(s1):
        if c != s2[i]:
            common = s1[:i]
            break
    prefix = "/" if isabs_ else ""
    return prefix + "/".join(common)


def _stat(path, follow):
    return _native("posix.stat", _str(path), follow)


def exists(path):
    try:
        _stat(path, True)
    except (OSError, ValueError):
        return False
    return True


def lexists(path):
    try:
        _stat(path, False)
    except (OSError, ValueError):
        return False
    return True


def isfile(path):
    try:
        st = _stat(path, True)
    except (OSError, ValueError):
        return False
    return (st[0] & 0o170000) == 0o100000


def isdir(s):
    try:
        st = _stat(s, True)
    except (OSError, ValueError):
        return False
    return (st[0] & 0o170000) == 0o040000


def islink(path):
    try:
        st = _stat(path, False)
    except (OSError, ValueError):
        return False
    return (st[0] & 0o170000) == 0o120000


def isjunction(path):
    return False


def isdevdrive(path):
    return False


def ismount(path):
    try:
        s1 = _stat(path, False)
    except (OSError, ValueError):
        return False
    if (s1[0] & 0o170000) == 0o120000:
        return False
    parent = join(_str(path), "..")
    try:
        s2 = _stat(parent, False)
    except OSError:
        parent = realpath(parent)
        try:
            s2 = _stat(parent, False)
        except OSError:
            return False
    return s1[2] != s2[2] or s1[1] == s2[1]


def getsize(filename):
    return _stat(filename, True)[6]


def getatime(filename):
    return _stat(filename, True)[10]


def getmtime(filename):
    return _stat(filename, True)[11]


def getctime(filename):
    return _stat(filename, True)[12]


def samefile(f1, f2):
    s1 = _stat(f1, True)
    s2 = _stat(f2, True)
    return s1[1] == s2[1] and s1[2] == s2[2]


def sameopenfile(fp1, fp2):
    s1 = _native("posix.fstat", fp1)
    s2 = _native("posix.fstat", fp2)
    return s1[1] == s2[1] and s1[2] == s2[2]


def samestat(s1, s2):
    return s1.st_ino == s2.st_ino and s1.st_dev == s2.st_dev


def expanduser(path):
    path = _str(path)
    if not path.startswith("~"):
        return path
    i = path.find("/", 1)
    if i < 0:
        i = len(path)
    if i == 1:
        home = _native("os.getenv", "HOME", None)
        if home is None:
            home = _native("posix.home", None)
            if home is None:
                return path
        userhome = home
    else:
        userhome = _native("posix.home", path[1:i])
        if userhome is None:
            return path
    userhome = userhome.rstrip("/")
    return (userhome + path[i:]) or "/"


def expandvars(path):
    path = _str(path)
    if "$" not in path:
        return path
    import re
    def repl(m):
        name = m[1]
        if name.startswith("{"):
            if not name.endswith("}"):
                return m[0]
            name = name[1:-1]
        value = _native("os.getenv", name, None)
        if value is None:
            return m[0]
        return value
    return re.sub(r"\$(\w+|\{[^}]*\}?)", repl, path, flags=re.ASCII)
