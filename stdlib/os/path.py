"""os.path for POSIX paths."""

sep = "/"
curdir = "."
pardir = ".."
extsep = "."
pathsep = ":"
altsep = None
defpath = "/bin:/usr/bin"
devnull = "/dev/null"


def isabs(s):
    return s.startswith("/")


def join(a, *p):
    path = a
    for b in p:
        if b.startswith("/"):
            path = b
        elif not path or path.endswith("/"):
            path += b
        else:
            path += "/" + b
    return path


def split(p):
    i = p.rfind("/") + 1
    head, tail = p[:i], p[i:]
    if head and head != "/" * len(head):
        head = head.rstrip("/")
    return (head, tail)


def basename(p):
    return p[p.rfind("/") + 1:]


def dirname(p):
    i = p.rfind("/") + 1
    head = p[:i]
    if head and head != "/" * len(head):
        head = head.rstrip("/")
    return head


def splitext(p):
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
    return (p[:0], p)


def normpath(path):
    if path == "":
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
    if not isabs(path):
        path = join(_native("os.getcwd"), path)
    return normpath(path)


def realpath(filename):
    return _native("os.path.realpath", filename)


def relpath(path, start=None):
    if not path:
        raise ValueError("no path specified")
    if start is None:
        start = "."
    start_list = [x for x in abspath(start).split("/") if x]
    path_list = [x for x in abspath(path).split("/") if x]
    i = 0
    while i < len(start_list) and i < len(path_list) and start_list[i] == path_list[i]:
        i += 1
    rel_list = [".."] * (len(start_list) - i) + path_list[i:]
    if not rel_list:
        return "."
    return join(*rel_list)


def commonprefix(m):
    if not m:
        return ""
    s1 = min(m)
    s2 = max(m)
    for i in range(len(s1)):
        if s1[i] != s2[i]:
            return s1[:i]
    return s1


def commonpath(paths):
    if not paths:
        raise ValueError("commonpath() arg is an empty sequence")
    split_paths = [[c for c in p.split("/") if c and c != "."] for p in paths]
    prefix = "/" if paths[0].startswith("/") else ""
    first = min(split_paths)
    last = max(split_paths)
    common = first
    for i in range(len(first)):
        if first[i] != last[i]:
            common = first[:i]
            break
    return prefix + "/".join(common)


def exists(path):
    return _native("os.path.exists", path)


def lexists(path):
    return _native("os.path.exists", path) or _native("os.path.islink", path)


def isfile(path):
    return _native("os.path.isfile", path)


def isdir(s):
    return _native("os.path.isdir", s)


def islink(path):
    return _native("os.path.islink", path)


def expanduser(path):
    return _native("os.path.expanduser", path)


def getsize(filename):
    return _native("os.stat", filename)[1]


def getmtime(filename):
    return _native("os.stat", filename)[2]
