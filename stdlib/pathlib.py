"""pathlib: object-oriented file system paths (POSIX flavour)."""
import os
import fnmatch

__all__ = ["UnsupportedOperation", "PurePath", "PurePosixPath", "Path", "PosixPath"]

from builtins import io_pm_UnsupportedOperation as UnsupportedOperation

_IGNORED_ERRNOS = (2, 20, 9, 40)


def _ignore_error(exception):
    return getattr(exception, "errno", None) in _IGNORED_ERRNOS


class _PathParents:
    def __init__(self, path):
        self._path = path
        self._tail = path._tail
        self._drv = path.drive
        self._root = path.root

    def __len__(self):
        return len(self._tail)

    def __getitem__(self, idx):
        if isinstance(idx, slice):
            return tuple(self[i] for i in range(len(self))[idx])
        if idx >= len(self) or idx < -len(self):
            raise IndexError(idx)
        if idx < 0:
            idx += len(self)
        return self._path._from_parsed_parts(self._drv, self._root, self._tail[:-idx - 1])

    def __iter__(self):
        for i in range(len(self)):
            yield self[i]

    def __repr__(self):
        return "<{}.parents>".format(type(self._path).__name__)


class PurePath:
    def __init__(self, *args):
        paths = []
        for arg in args:
            if isinstance(arg, PurePath):
                paths.extend(arg._raw_paths)
            else:
                try:
                    path = os.fspath(arg)
                except TypeError:
                    path = arg
                if not isinstance(path, str):
                    raise TypeError("argument should be a str or an os.PathLike "
                                    "object where __fspath__ returns a str, "
                                    "not " + repr(type(path).__name__))
                paths.append(path)
        self._raw_paths = paths
        self._drv = ""
        self._root = ""
        self._tail_cached = []
        self._parsed = False
        self._str = ""

    def with_segments(self, *pathsegments):
        return type(self)(*pathsegments)

    def _load_parts(self):
        paths = self._raw_paths
        if not paths:
            path = ""
        elif len(paths) == 1:
            path = paths[0]
        else:
            path = os.path.join(*paths)
        drv, root, tail = self._parse(path)
        self._drv = drv
        self._root = root
        self._tail_cached = tail
        self._parsed = True

    def _parse(self, path):
        if path.startswith("/"):
            if path.startswith("//") and not path.startswith("///"):
                root = "//"
            else:
                root = "/"
            rel = path.lstrip("/") if root == "/" else path[2:]
        else:
            root = ""
            rel = path
        tail = [x for x in rel.split("/") if x and x != "."]
        return "", root, tail

    def _from_parsed_parts(self, drv, root, tail):
        path_str = self._format_parsed_parts(drv, root, tail)
        path = self.with_segments(path_str)
        path._str = path_str or "."
        path._drv = drv
        path._root = root
        path._tail_cached = tail
        path._parsed = True
        return path

    def _format_parsed_parts(self, drv, root, tail):
        if drv or root:
            return drv + root + "/".join(tail)
        elif tail and ":" in tail[0]:
            return "./" + "/".join(tail)
        else:
            return "/".join(tail)

    @property
    def drive(self):
        if not self._parsed:
            self._load_parts()
        return self._drv

    @property
    def root(self):
        if not self._parsed:
            self._load_parts()
        return self._root

    @property
    def _tail(self):
        if not self._parsed:
            self._load_parts()
        return self._tail_cached

    @property
    def anchor(self):
        return self.drive + self.root

    @property
    def parts(self):
        if self.drive or self.root:
            return tuple([self.drive + self.root] + self._tail)
        return tuple(self._tail)

    def __str__(self):
        if self._str == "":
            self._str = self._format_parsed_parts(self.drive, self.root, self._tail) or "."
        return self._str

    def __fspath__(self):
        return str(self)

    def as_posix(self):
        return str(self)

    def __bytes__(self):
        return os.fsencode(self)

    def __repr__(self):
        return "{}({!r})".format(_class_label(self), self.as_posix())

    def as_uri(self):
        if not self.is_absolute():
            raise ValueError("relative path can't be expressed as a file URI")
        return "file://" + _quote(os.fsencode(str(self)))

    def joinpath(self, *pathsegments):
        return self.with_segments(self, *pathsegments)

    def __truediv__(self, key):
        if not isinstance(key, (str, PurePath)) and not hasattr(key, "__fspath__"):
            return NotImplemented
        return self.with_segments(self, key)

    def __rtruediv__(self, key):
        if not isinstance(key, (str, PurePath)) and not hasattr(key, "__fspath__"):
            return NotImplemented
        return self.with_segments(key, self)

    def _key(self):
        return str(self).split("/")

    def __eq__(self, other):
        if not isinstance(other, PurePath):
            return NotImplemented
        return str(self) == str(other)

    def __hash__(self):
        return hash(str(self))

    def __lt__(self, other):
        if not isinstance(other, PurePath):
            return NotImplemented
        return self._key() < other._key()

    def __le__(self, other):
        if not isinstance(other, PurePath):
            return NotImplemented
        return self._key() <= other._key()

    def __gt__(self, other):
        if not isinstance(other, PurePath):
            return NotImplemented
        return self._key() > other._key()

    def __ge__(self, other):
        if not isinstance(other, PurePath):
            return NotImplemented
        return self._key() >= other._key()

    @property
    def parent(self):
        drv = self.drive
        root = self.root
        tail = self._tail
        if not tail:
            return self
        return self._from_parsed_parts(drv, root, tail[:-1])

    @property
    def parents(self):
        return _PathParents(self)

    @property
    def name(self):
        tail = self._tail
        if not tail:
            return ""
        return tail[-1]

    @property
    def suffix(self):
        name = self.name
        i = name.rfind(".")
        if 0 < i < len(name) - 1:
            return name[i:]
        else:
            return ""

    @property
    def suffixes(self):
        name = self.name
        if name.endswith("."):
            return []
        name = name.lstrip(".")
        return ["." + suffix for suffix in name.split(".")[1:]]

    @property
    def stem(self):
        name = self.name
        i = name.rfind(".")
        if 0 < i < len(name) - 1:
            return name[:i]
        else:
            return name

    def with_name(self, name):
        if not name or "/" in name or name == ".":
            raise ValueError("Invalid name {!r}".format(name))
        tail = list(self._tail)
        if not tail:
            raise ValueError("{!r} has an empty name".format(self))
        tail[-1] = name
        return self._from_parsed_parts(self.drive, self.root, tail)

    def with_stem(self, stem):
        return self.with_name(stem + self.suffix)

    def with_suffix(self, suffix):
        stem = self.stem
        if not stem:
            raise ValueError("{!r} has an empty name".format(self))
        elif suffix and not suffix.startswith(".") or suffix == ".":
            raise ValueError("Invalid suffix {!r}".format(suffix))
        else:
            return self.with_name(stem + suffix)

    def relative_to(self, other, /, *_deprecated, walk_up=False):
        if _deprecated:
            other = self.with_segments(other, *_deprecated)
        elif not isinstance(other, PurePath):
            other = self.with_segments(other)
        step = 0
        found = None
        candidates = [other] + list(other.parents)
        for path in candidates:
            if path == self or path in self.parents:
                found = path
                break
            elif not walk_up:
                raise ValueError("{!r} is not in the subpath of {!r}".format(str(self), str(other)))
            elif path.name == "..":
                raise ValueError("'..' segment in {!r} cannot be walked".format(str(other)))
            step += 1
        if found is None:
            raise ValueError("{!r} and {!r} have different anchors".format(str(self), str(other)))
        parts = [".."] * step + self._tail[len(found._tail):]
        return self._from_parsed_parts("", "", parts)

    def is_relative_to(self, other, /, *_deprecated):
        if _deprecated:
            other = self.with_segments(other, *_deprecated)
        elif not isinstance(other, PurePath):
            other = self.with_segments(other)
        return other == self or other in self.parents

    def is_absolute(self):
        for path in self._raw_paths:
            if path.startswith("/"):
                return True
        return False

    def is_reserved(self):
        return False

    def match(self, path_pattern, *, case_sensitive=None):
        if not isinstance(path_pattern, PurePath):
            path_pattern = self.with_segments(path_pattern)
        pat_parts = list(path_pattern.parts)
        if not pat_parts:
            raise ValueError("empty pattern")
        parts = list(self.parts)
        if path_pattern.anchor:
            if len(parts) != len(pat_parts):
                return False
        elif len(pat_parts) > len(parts):
            return False
        for i in range(1, len(pat_parts) + 1):
            if not _match_part(parts[-i], pat_parts[-i], case_sensitive):
                return False
        return True

    def full_match(self, pattern, *, case_sensitive=None):
        if not isinstance(pattern, PurePath):
            pattern = self.with_segments(pattern)
        return _full_match(list(self.parts), list(pattern.parts), case_sensitive)


def _class_label(p):
    name = type(p).__name__
    if name == "Path":
        return "PosixPath"
    if name == "PurePath":
        return "PurePosixPath"
    return name


def _quote(data):
    out = []
    safe = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_.-~/"
    for b in data:
        c = chr(b)
        if c in safe:
            out.append(c)
        else:
            out.append("%%%02X" % b)
    return "".join(out)


def _match_part(name, pat, case_sensitive):
    if case_sensitive is False:
        return fnmatch.fnmatchcase(name.lower(), pat.lower())
    return fnmatch.fnmatchcase(name, pat)


def _full_match(parts, pats, case_sensitive):
    if not pats:
        return not parts
    if pats[0] == "**":
        for i in range(len(parts) + 1):
            if _full_match(parts[i:], pats[1:], case_sensitive):
                return True
        return False
    if not parts:
        return False
    if not _match_part(parts[0], pats[0], case_sensitive):
        return False
    return _full_match(parts[1:], pats[1:], case_sensitive)


class PurePosixPath(PurePath):
    pass


def _has_magic(s):
    return "*" in s or "?" in s or "[" in s


class Path(PurePath):
    def __init__(self, *args, **kwargs):
        PurePath.__init__(self, *args)

    def stat(self, *, follow_symlinks=True):
        return os.stat(self, follow_symlinks=follow_symlinks)

    def lstat(self):
        return os.lstat(self)

    def exists(self, *, follow_symlinks=True):
        if follow_symlinks:
            return os.path.exists(self)
        return os.path.lexists(self)

    def is_dir(self, *, follow_symlinks=True):
        if follow_symlinks:
            return os.path.isdir(self)
        try:
            return _s_isdir(self.stat(follow_symlinks=follow_symlinks).st_mode)
        except (OSError, ValueError):
            return False

    def is_file(self, *, follow_symlinks=True):
        if follow_symlinks:
            return os.path.isfile(self)
        try:
            return _s_isreg(self.stat(follow_symlinks=follow_symlinks).st_mode)
        except (OSError, ValueError):
            return False

    def is_mount(self):
        return os.path.ismount(self)

    def is_symlink(self):
        return os.path.islink(self)

    def is_junction(self):
        return False

    def is_block_device(self):
        try:
            return _s_isfmt(self.stat().st_mode, 0o060000)
        except (OSError, ValueError):
            return False

    def is_char_device(self):
        try:
            return _s_isfmt(self.stat().st_mode, 0o020000)
        except (OSError, ValueError):
            return False

    def is_fifo(self):
        try:
            return _s_isfmt(self.stat().st_mode, 0o010000)
        except (OSError, ValueError):
            return False

    def is_socket(self):
        try:
            return _s_isfmt(self.stat().st_mode, 0o140000)
        except (OSError, ValueError):
            return False

    def samefile(self, other_path):
        st = self.stat()
        try:
            other_st = other_path.stat()
        except AttributeError:
            other_st = self.with_segments(other_path).stat()
        return st.st_ino == other_st.st_ino and st.st_dev == other_st.st_dev

    def open(self, mode="r", buffering=-1, encoding=None, errors=None, newline=None):
        if "b" not in mode and encoding is None:
            encoding = "utf-8"
        return open(self, mode, buffering, encoding, errors, newline)

    def read_bytes(self):
        with self.open(mode="rb", buffering=0) as f:
            return f.read()

    def read_text(self, encoding=None, errors=None, newline=None):
        with self.open(mode="r", encoding=encoding, errors=errors, newline=newline) as f:
            return f.read()

    def write_bytes(self, data):
        view = bytes(data)
        with self.open(mode="wb") as f:
            return f.write(view)

    def write_text(self, data, encoding=None, errors=None, newline=None):
        if not isinstance(data, str):
            raise TypeError("data must be str, not %s" % type(data).__name__)
        with self.open(mode="w", encoding=encoding, errors=errors, newline=newline) as f:
            return f.write(data)

    def iterdir(self):
        root_dir = str(self)
        with os.scandir(root_dir) as scandir_it:
            names = [entry.name for entry in scandir_it]
        return iter([self._make_child_relpath(n) for n in names])

    def _make_child_relpath(self, name):
        path_str = str(self)
        tail = self._tail
        if tail:
            path_str = path_str + "/" + name
        elif path_str != ".":
            path_str = path_str + name
        else:
            path_str = name
        path = self.with_segments(path_str)
        path._str = path_str
        path._drv = self.drive
        path._root = self.root
        path._tail_cached = tail + [name]
        path._parsed = True
        return path

    def glob(self, pattern, *, case_sensitive=None, recurse_symlinks=False):
        if not isinstance(pattern, PurePath):
            pattern = self.with_segments(pattern)
        if pattern.anchor:
            raise NotImplementedError("Non-relative patterns are unsupported")
        parts = list(pattern._tail)
        if not parts:
            raise ValueError("Unacceptable pattern: {!r}".format(pattern))
        raw = pattern._raw_paths[-1] if pattern._raw_paths else ""
        dirs_only = raw.endswith("/")
        results = []
        seen = set()
        self._glob(self, parts, case_sensitive, recurse_symlinks, dirs_only, results, seen)
        return iter(results)

    def _glob(self, base, parts, case_sensitive, recurse_symlinks, dirs_only, results, seen):
        if not parts:
            key = str(base)
            if key not in seen:
                seen.add(key)
                if not dirs_only or base.is_dir():
                    results.append(base)
            return
        part = parts[0]
        rest = parts[1:]
        if part == "**":
            self._glob(base, rest, case_sensitive, recurse_symlinks, dirs_only, results, seen)
            try:
                with os.scandir(str(base)) as it:
                    entries = list(it)
            except OSError:
                return
            for entry in entries:
                try:
                    is_dir = entry.is_dir(follow_symlinks=recurse_symlinks)
                except OSError:
                    is_dir = False
                child = base._make_child_relpath(entry.name)
                if not rest and not is_dir:
                    self._glob(child, rest, case_sensitive, recurse_symlinks, dirs_only, results, seen)
                if is_dir:
                    self._glob(child, parts, case_sensitive, recurse_symlinks, dirs_only, results, seen)
            return
        if not _has_magic(part):
            child = base._make_child_relpath(part)
            if rest:
                if child.is_dir():
                    self._glob(child, rest, case_sensitive, recurse_symlinks, dirs_only, results, seen)
            else:
                if child.exists(follow_symlinks=False) if False else os.path.lexists(str(child)):
                    self._glob(child, rest, case_sensitive, recurse_symlinks, dirs_only, results, seen)
            return
        try:
            with os.scandir(str(base)) as it:
                entries = list(it)
        except OSError:
            return
        for entry in entries:
            if not _match_part(entry.name, part, case_sensitive):
                continue
            child = base._make_child_relpath(entry.name)
            if rest:
                try:
                    ok = entry.is_dir()
                except OSError:
                    ok = False
                if ok:
                    self._glob(child, rest, case_sensitive, recurse_symlinks, dirs_only, results, seen)
            else:
                self._glob(child, rest, case_sensitive, recurse_symlinks, dirs_only, results, seen)

    def rglob(self, pattern, *, case_sensitive=None, recurse_symlinks=False):
        if not isinstance(pattern, PurePath):
            pattern = self.with_segments(pattern)
        return self.glob("**/" + str(pattern) if str(pattern) != "." else "**", case_sensitive=case_sensitive,
                         recurse_symlinks=recurse_symlinks)

    def walk(self, top_down=True, on_error=None, follow_symlinks=False):
        for path_str, dirnames, filenames in os.walk(str(self), top_down, on_error, follow_symlinks):
            yield (self.with_segments(path_str), dirnames, filenames)

    @classmethod
    def cwd(cls):
        return cls(os.getcwd())

    @classmethod
    def home(cls):
        return cls("~").expanduser()

    def absolute(self):
        if self.is_absolute():
            return self
        cwd = os.getcwd()
        if not self.root and not self._tail:
            return self.with_segments(cwd)
        return self.with_segments(cwd, self)

    def resolve(self, strict=False):
        return self.with_segments(os.path.realpath(self, strict=strict))

    def owner(self, *, follow_symlinks=True):
        import pwd
        return pwd.getpwuid(self.stat(follow_symlinks=follow_symlinks).st_uid).pw_name

    def group(self, *, follow_symlinks=True):
        import grp
        return grp.getgrgid(self.stat(follow_symlinks=follow_symlinks).st_gid).gr_name

    def readlink(self):
        return self.with_segments(os.readlink(self))

    def touch(self, mode=438, exist_ok=True):
        if exist_ok:
            try:
                os.utime(self, None)
            except OSError:
                pass
            else:
                return
        flags = os.O_CREAT | os.O_WRONLY
        if not exist_ok:
            flags |= os.O_EXCL
        fd = os.open(self, flags, mode)
        os.close(fd)

    def mkdir(self, mode=511, parents=False, exist_ok=False):
        try:
            os.mkdir(self, mode)
        except FileNotFoundError:
            if not parents or self.parent == self:
                raise
            self.parent.mkdir(parents=True, exist_ok=True)
            self.mkdir(mode, parents=False, exist_ok=exist_ok)
        except OSError:
            if not exist_ok or not self.is_dir():
                raise

    def chmod(self, mode, *, follow_symlinks=True):
        os.chmod(self, mode, follow_symlinks=follow_symlinks)

    def lchmod(self, mode):
        self.chmod(mode, follow_symlinks=False)

    def unlink(self, missing_ok=False):
        try:
            os.unlink(self)
        except FileNotFoundError:
            if not missing_ok:
                raise

    def rmdir(self):
        os.rmdir(self)

    def rename(self, target):
        os.rename(self, target)
        return self.with_segments(target)

    def replace(self, target):
        os.replace(self, target)
        return self.with_segments(target)

    def symlink_to(self, target, target_is_directory=False):
        os.symlink(target, self, target_is_directory)

    def hardlink_to(self, target):
        os.link(target, self)

    def expanduser(self):
        if not (self.drive or self.root) and self._tail and self._tail[0][:1] == "~":
            homedir = os.path.expanduser(self._tail[0])
            if homedir[:1] == "~":
                raise RuntimeError("Could not determine home directory.")
            drv, root, tail = self._parse(homedir)
            return self._from_parsed_parts(drv, root, tail + self._tail[1:])
        return self

    @classmethod
    def from_uri(cls, uri):
        if not uri.startswith("file:"):
            raise ValueError("URI does not start with 'file:': {!r}".format(uri))
        path = uri[5:]
        if path[:3] == "///":
            path = path[2:]
        elif path[:12] == "//localhost/":
            path = path[11:]
        if path[:3] == "///" or (path[:1] == "/" and path[1:2] != "/"):
            path = cls(os.fsdecode(_unquote(path)))
            if path.is_absolute():
                return path
        raise ValueError("URI is not absolute: {!r}".format(uri))


def _s_isfmt(mode, fmt):
    return (mode & 0o170000) == fmt


def _s_isdir(mode):
    return (mode & 0o170000) == 0o040000


def _s_isreg(mode):
    return (mode & 0o170000) == 0o100000


class PosixPath(Path, PurePosixPath):
    pass


def _unquote(s):
    out = bytearray()
    raw = s.encode("utf-8")
    i = 0
    while i < len(raw):
        c = raw[i]
        if c == 37 and i + 2 < len(raw) + 0 and i + 2 <= len(raw) - 1 + 0:
            try:
                out.append(int(raw[i + 1:i + 3].decode("ascii"), 16))
                i += 3
                continue
            except ValueError:
                pass
        out.append(c)
        i += 1
    return bytes(out)
