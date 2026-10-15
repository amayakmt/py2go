"""Common pathname manipulations, WindowsNT/95 version.

Instead of importing this module directly, import os and refer to this
module as os.path. A reduced version of CPython's: it works on str and on
bytes (decoded with the file system encoding and encoded back), and keeps
no Windows API: the file tests are those of the system it runs on.
"""
import os
import re
import stat

curdir = '.'
pardir = '..'
extsep = '.'
sep = '\\'
pathsep = ';'
altsep = '/'
defpath = '.;C:\\bin'
devnull = 'nul'
supports_unicode_filenames = True
ALLOW_MISSING = object()

__all__ = ['normcase', 'isabs', 'join', 'splitdrive', 'splitroot', 'split', 'splitext', 'basename', 'dirname', 'commonprefix', 'getsize', 'getmtime', 'getatime', 'getctime', 'islink', 'exists', 'lexists', 'isdir', 'isfile', 'ismount', 'isreserved', 'expanduser', 'expandvars', 'normpath', 'abspath', 'curdir', 'pardir', 'sep', 'pathsep', 'defpath', 'altsep', 'extsep', 'devnull', 'realpath', 'supports_unicode_filenames', 'relpath', 'samefile', 'sameopenfile', 'samestat', 'commonpath', 'isjunction', 'isdevdrive', 'ALLOW_MISSING']


def _obj(x: object) -> object:
    return x


def _dec(p):
    """The str of a path-like, and whether it was bytes."""
    p = os.fspath(p)
    if isinstance(p, bytes):
        return os.fsdecode(p), True
    return p, False


def _enc(s, was_bytes):
    if was_bytes:
        return os.fsencode(s)
    return s


def _check(path):
    p = os.fspath(path)
    if not isinstance(p, (str, bytes)):
        raise TypeError('expected str, bytes or os.PathLike object, not ' + type(p).__name__)
    return p


def normcase(s):
    """Normalize case of pathname.

    Makes all characters lowercase and all slashes into backslashes."""
    s, b = _dec(s)
    return _enc(s.replace('/', '\\').lower(), b)


def isabs(s):
    """Test whether a path is absolute"""
    s, b = _dec(s)
    s = s[:3].replace('/', '\\')
    return s.startswith(':\\', 1) or s.startswith('\\\\')


def _splitroot(p):
    normp = p.replace('/', '\\')
    if normp[:1] == '\\':
        if normp[1:2] == '\\':
            start = 8 if normp[:8].upper() == '\\\\?\\UNC\\' else 2
            index = normp.find('\\', start)
            if index == -1:
                return (p, '', '')
            index2 = normp.find('\\', index + 1)
            if index2 == -1:
                return (p, '', '')
            return (p[:index2], p[index2:index2 + 1], p[index2 + 1:])
        return ('', p[:1], p[1:])
    elif normp[1:2] == ':':
        if normp[2:3] == '\\':
            return (p[:2], p[2:3], p[3:])
        return (p[:2], '', p[2:])
    return ('', '', p)


def splitroot(p):
    """Split a pathname into drive, root and tail.

    The tail contains anything after the root."""
    s, b = _dec(p)
    d, r, t = _splitroot(s)
    return (_enc(d, b), _enc(r, b), _enc(t, b))


def join(path, *paths):
    p, b = _dec(_check(path))
    result_drive, result_root, result_path = _splitroot(p)
    for q in paths:
        qs, qb = _dec(_check(q))
        if qb != b:
            raise TypeError("Can't mix strings and bytes in path components")
        p_drive, p_root, p_path = _splitroot(qs)
        if p_root:
            if p_drive or not result_drive:
                result_drive = p_drive
            result_root = p_root
            result_path = p_path
            continue
        elif p_drive and p_drive != result_drive:
            if p_drive.lower() != result_drive.lower():
                result_drive = p_drive
                result_root = p_root
                result_path = p_path
                continue
            result_drive = p_drive
        if result_path and result_path[-1] not in '\\/':
            result_path = result_path + '\\'
        result_path = result_path + p_path
    if result_path and (not result_root) and result_drive and (result_drive[-1] not in ':\\/'):
        return _enc(result_drive + '\\' + result_path, b)
    return _enc(result_drive + result_root + result_path, b)


def splitdrive(p):
    """Split a pathname into drive/UNC sharepoint and relative path specifiers.
    Returns a 2-tuple (drive_or_unc, path); either part may be empty."""
    s, b = _dec(p)
    drive, root, tail = _splitroot(s)
    return (_enc(drive, b), _enc(root + tail, b))


def split(p):
    """Split a pathname.

    Return tuple (head, tail) where tail is everything after the final
    slash.  Either part may be empty."""
    s, b = _dec(p)
    d, r, rest = _splitroot(s)
    i = len(rest)
    while i and rest[i - 1] not in '\\/':
        i -= 1
    head, tail = (rest[:i], rest[i:])
    return (_enc(d + r + head.rstrip('\\/'), b), _enc(tail, b))


def splitext(p):
    s, b = _dec(p)
    sepIndex = max(s.rfind('\\'), s.rfind('/'))
    dotIndex = s.rfind('.')
    if dotIndex > sepIndex:
        filenameIndex = sepIndex + 1
        while filenameIndex < dotIndex:
            if s[filenameIndex] != '.':
                return (_enc(s[:dotIndex], b), _enc(s[dotIndex:], b))
            filenameIndex += 1
    return (_enc(s, b), _enc('', b))


def basename(p):
    """Returns the final component of a pathname"""
    return split(p)[1]


def dirname(p):
    """Returns the directory component of a pathname"""
    return split(p)[0]


def normpath(path):
    """Normalize path, eliminating double slashes, etc."""
    p, b = _dec(path)
    p = p.replace('/', '\\')
    drive, root, rest = _splitroot(p)
    prefix = drive + root
    comps = rest.split('\\')
    i = 0
    while i < len(comps):
        if not comps[i] or comps[i] == '.':
            del comps[i]
        elif comps[i] == '..':
            if i > 0 and comps[i - 1] != '..':
                del comps[i - 1:i + 1]
                i -= 1
            elif i == 0 and root:
                del comps[i]
            else:
                i += 1
        else:
            i += 1
    if not prefix and (not comps):
        comps.append('.')
    return _enc(prefix + '\\'.join(comps), b)


def abspath(path):
    """Return the absolute version of a path."""
    p, b = _dec(path)
    if not isabs(p):
        p = join(os.getcwd(), p)
    return _enc(normpath(p), b)


def realpath(path, *, strict=False):
    return abspath(path)


def commonprefix(m):
    "Given a list of pathnames, returns the longest common leading component"
    if not m:
        return ''
    if not isinstance(m[0], (list, tuple)):
        m = tuple(map(os.fspath, m))
    s1 = min(m)
    s2 = max(m)
    for i, c in enumerate(s1):
        if c != s2[i]:
            return s1[:i]
    return s1


def relpath(path, start=None):
    """Return a relative version of a path"""
    p, b = _dec(path)
    if not p:
        raise ValueError('no path specified')
    if start is None:
        st = '.'
    else:
        st, sb = _dec(start)
    start_abs = abspath(st)
    path_abs = abspath(p)
    start_drive, _r1, start_rest = _splitroot(start_abs)
    path_drive, _r2, path_rest = _splitroot(path_abs)
    if start_drive.replace('/', '\\').lower() != path_drive.replace('/', '\\').lower():
        raise ValueError('path is on mount %r, start on mount %r' % (path_drive, start_drive))
    start_list = start_rest.split('\\') if start_rest else []
    path_list = path_rest.split('\\') if path_rest else []
    i = 0
    for e1, e2 in zip(start_list, path_list):
        if e1.lower() != e2.lower():
            break
        i += 1
    rel_list = ['..'] * (len(start_list) - i) + path_list[i:]
    if not rel_list:
        return _enc('.', b)
    return _enc('\\'.join(rel_list), b)


def commonpath(paths):
    """Given an iterable of path names, returns the longest common sub-path."""
    plist = [os.fspath(p) for p in paths]
    if not plist:
        raise ValueError('commonpath() arg is an empty iterable')
    b = isinstance(plist[0], bytes)
    strs = [os.fsdecode(p) if isinstance(p, bytes) else p for p in plist]
    drivesplits = [_splitroot(p.replace('/', '\\').lower()) for p in strs]
    split_paths = [t[2].split('\\') for t in drivesplits]
    if len(set([t[0] for t in drivesplits])) != 1:
        raise ValueError("Paths don't have the same drive")
    drive, root, path = _splitroot(strs[0].replace('/', '\\'))
    if len(set([t[1] for t in drivesplits])) != 1:
        if drive:
            raise ValueError("Can't mix absolute and relative paths")
        raise ValueError("Can't mix rooted and not-rooted paths")
    common = path.split('\\')
    common = [c for c in common if c and c != '.']
    split_paths = [[c for c in s if c and c != '.'] for s in split_paths]
    s1 = min(split_paths)
    s2 = max(split_paths)
    for i, c in enumerate(s1):
        if c != s2[i]:
            common = common[:i]
            break
    else:
        common = common[:len(s1)]
    return _enc(drive + root + '\\'.join(common), b)


_reserved_chars = frozenset({chr(i) for i in range(32)} | {'"', '*', ':', '<', '>', '?', '|', '/', '\\'})
_reserved_names = frozenset({'CON', 'PRN', 'AUX', 'NUL', 'CONIN$', 'CONOUT$'} | {'COM%s' % c for c in '123456789\u00b9\u00b2\u00b3'} | {'LPT%s' % c for c in '123456789\u00b9\u00b2\u00b3'})


def isreserved(path):
    """Return true if the pathname is reserved by the system."""
    s, b = _dec(path)
    rest = _splitroot(s)[2].replace('/', '\\')
    for name in reversed(rest.split('\\')):
        if _isreservedname(name):
            return True
    return False


def _isreservedname(name):
    """Return true if the filename is reserved by the system."""
    if name[-1:] in ('.', ' '):
        return name not in ('.', '..')
    for c in name:
        if c in _reserved_chars:
            return True
    return name.partition('.')[0].rstrip(' ').upper() in _reserved_names


def expanduser(path):
    """Expand ~ and ~user constructs.

    If user or home directory is unknown, do nothing."""
    p, b = _dec(path)
    if not p.startswith('~'):
        return path
    i, n = (1, len(p))
    while i < n and p[i] not in '\\/':
        i += 1
    if 'USERPROFILE' in os.environ:
        userhome = os.environ['USERPROFILE']
    elif 'HOMEPATH' not in os.environ:
        return path
    else:
        drive = os.environ.get('HOMEDRIVE', '')
        userhome = join(drive, os.environ['HOMEPATH'])
    if i != 1:
        target_user = p[1:i]
        current_user = os.environ.get('USERNAME')
        if target_user != current_user:
            if current_user != basename(userhome):
                return path
            userhome = join(dirname(userhome), target_user)
    return _enc(userhome + p[i:], b)


_varpattern = "'[^']*'?|%(%|[^%]*%?)|\\$(\\$|[-\\w]+|\\{[^}]*\\}?)"


def _expand_repl(m):
    lastindex = m.lastindex
    if lastindex is None:
        return m[0]
    name = m[lastindex]
    if lastindex == 1:
        if name == '%':
            return name
        if not name.endswith('%'):
            return m[0]
        name = name[:-1]
    else:
        if name == '$':
            return name
        if name.startswith('{'):
            if not name.endswith('}'):
                return m[0]
            name = name[1:-1]
    if name in os.environ:
        return os.environ[name]
    return m[0]


def expandvars(path):
    """Expand shell variables of the forms $var, ${var} and %var%.

    Unknown variables are left unchanged."""
    p, b = _dec(path)
    if '$' not in p and '%' not in p:
        return path
    return _enc(re.sub(_varpattern, _expand_repl, p, flags=re.ASCII), b)


def exists(path):
    try:
        os.stat(path)
    except (OSError, ValueError):
        return False
    return True


def lexists(path):
    try:
        os.lstat(path)
    except (OSError, ValueError):
        return False
    return True


def isdir(s):
    try:
        st = os.stat(s)
    except (OSError, ValueError):
        return False
    return stat.S_ISDIR(st.st_mode)


def isfile(path):
    try:
        st = os.stat(path)
    except (OSError, ValueError):
        return False
    return stat.S_ISREG(st.st_mode)


def islink(path):
    try:
        st = os.lstat(path)
    except (OSError, ValueError, AttributeError):
        return False
    return stat.S_ISLNK(st.st_mode)


def isjunction(path):
    return False


def isdevdrive(path):
    return False


def getsize(filename):
    return os.stat(filename).st_size


def getmtime(filename):
    return os.stat(filename).st_mtime


def getatime(filename):
    return os.stat(filename).st_atime


def getctime(filename):
    return os.stat(filename).st_ctime


def samestat(s1, s2):
    return s1.st_ino == s2.st_ino and s1.st_dev == s2.st_dev


def samefile(f1, f2):
    s1 = os.stat(f1)
    s2 = os.stat(f2)
    return samestat(s1, s2)


def sameopenfile(fp1, fp2):
    s1 = os.fstat(fp1)
    s2 = os.fstat(fp2)
    return samestat(s1, s2)


def ismount(path):
    """Test whether a path is a mount point (a drive root or the root of a share)"""
    p, b = _dec(path)
    p = abspath(p)
    drive, root, rest = _splitroot(p)
    if drive and drive[0] in '\\/':
        return not rest
    if root and (not rest):
        return True
    return False
