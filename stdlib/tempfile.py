"""tempfile: temporary files and directories."""
import os
import errno
import io
import shutil
import warnings
from random import Random as _Random

__all__ = [
    "NamedTemporaryFile", "TemporaryFile", "SpooledTemporaryFile",
    "TemporaryDirectory", "mkstemp", "mkdtemp", "mktemp", "TMP_MAX",
    "gettempprefix", "tempdir", "gettempdir", "gettempprefixb", "gettempdirb",
]

TMP_MAX = 10000


def _obj(x: object) -> object:
    return x

template = "tmp"

_text_openflags = os.O_RDWR | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW
_bin_openflags = _text_openflags

tempdir = None


class _RandomNameSequence:
    characters = "abcdefghijklmnopqrstuvwxyz0123456789_"

    def __init__(self):
        self._rng = _Random()

    def __iter__(self):
        return self

    def __next__(self):
        return "".join(self._rng.choices(self.characters, k=8))


def _candidate_tempdir_list():
    dirlist = []
    for envname in ("TMPDIR", "TEMP", "TMP"):
        dirname = os.getenv(envname)
        if dirname:
            dirlist.append(dirname)
    dirlist.extend(["/tmp", "/var/tmp", "/usr/tmp"])
    try:
        dirlist.append(os.getcwd())
    except (AttributeError, OSError):
        dirlist.append(os.curdir)
    return dirlist


def _get_default_tempdir():
    namer = _RandomNameSequence()
    dirlist = _candidate_tempdir_list()
    for dir in dirlist:
        if dir != os.curdir:
            dir = os.path.abspath(dir)
        for seq in range(100):
            name = next(namer)
            filename = os.path.join(dir, name)
            try:
                fd = os.open(filename, _bin_openflags, 0o600)
                try:
                    try:
                        os.write(fd, b"blat")
                    finally:
                        os.close(fd)
                finally:
                    os.unlink(filename)
                return dir
            except FileExistsError:
                pass
            except PermissionError:
                if os.name == "nt" and os.path.isdir(dir) and os.access(dir, os.W_OK):
                    continue
                break
            except OSError:
                break
    raise FileNotFoundError(errno.ENOENT, "No usable temporary directory found in %s" % dirlist)


_name_sequence = None


def _get_candidate_names():
    global _name_sequence
    if _name_sequence is None:
        _name_sequence = _RandomNameSequence()
    return _name_sequence


def _mkstemp_inner(dir, pre, suf, flags, output_type):
    dir = os.path.abspath(dir)
    names = _get_candidate_names()
    for seq in range(TMP_MAX):
        name = next(names)
        file = os.path.join(dir, pre + name + suf)
        try:
            fd = os.open(file, flags, 0o600)
        except FileExistsError:
            continue
        except PermissionError:
            if os.name == "nt" and os.path.isdir(dir) and os.access(dir, os.W_OK):
                continue
            else:
                raise
        return fd, file
    raise FileExistsError(errno.EEXIST, "No usable temporary file name found")


def _sanitize_params(prefix, suffix, dir):
    if suffix is None:
        suffix = ""
    if prefix is None:
        prefix = template
    if dir is None:
        dir = gettempdir()
    else:
        dir = os.fspath(dir)
    return prefix, suffix, dir, "str"


def gettempprefix():
    return template


def gettempprefixb():
    return template.encode()


def gettempdir():
    global tempdir
    if tempdir is None:
        tempdir = _get_default_tempdir()
    return tempdir


def gettempdirb():
    return gettempdir().encode()


def mkstemp(suffix=None, prefix=None, dir=None, text=False):
    prefix, suffix, dir, output_type = _sanitize_params(prefix, suffix, dir)
    if text:
        flags = _text_openflags
    else:
        flags = _bin_openflags
    return _mkstemp_inner(dir, prefix, suffix, flags, output_type)


def mkdtemp(suffix=None, prefix=None, dir=None):
    prefix, suffix, dir, output_type = _sanitize_params(prefix, suffix, dir)
    names = _get_candidate_names()
    for seq in range(TMP_MAX):
        name = next(names)
        file = os.path.join(dir, prefix + name + suffix)
        try:
            os.mkdir(file, 0o700)
        except FileExistsError:
            continue
        except PermissionError:
            if os.name == "nt" and os.path.isdir(dir) and os.access(dir, os.W_OK):
                continue
            else:
                raise
        return os.path.abspath(file)
    raise FileExistsError(errno.EEXIST, "No usable temporary directory name found")


def mktemp(suffix="", prefix=template, dir=None):
    if dir is None:
        dir = gettempdir()
    names = _get_candidate_names()
    for seq in range(TMP_MAX):
        name = next(names)
        file = os.path.join(dir, prefix + name + suffix)
        if not os.path.lexists(file):
            return file
    raise FileExistsError(errno.EEXIST, "No usable temporary filename found")


class _TemporaryFileCloser:
    def __init__(self, file, name, delete=True, delete_on_close=True):
        self.file = file
        self.name = name
        self.delete = delete
        self.delete_on_close = delete_on_close
        self.cleanup_called = False
        self.close_called = False

    def cleanup(self, unlink=os.unlink):
        if not self.cleanup_called:
            self.cleanup_called = True
            try:
                if not self.close_called:
                    self.close_called = True
                    self.file.close()
            finally:
                if self.delete and not (self.delete_on_close and False):
                    try:
                        unlink(self.name)
                    except FileNotFoundError:
                        pass

    def close(self):
        if not self.close_called:
            self.close_called = True
            try:
                self.file.close()
            finally:
                if self.delete and self.delete_on_close:
                    self.cleanup()


class _TemporaryFileWrapper:
    def __init__(self, file, name, delete=True, delete_on_close=True):
        self.file = file
        self.name = name
        self._closer = _TemporaryFileCloser(file, name, delete, delete_on_close)

    def __getattr__(self, name):
        return getattr(self.file, name)

    def __enter__(self):
        self.file.__enter__()
        return self

    def __exit__(self, exc, value, tb):
        result = self.file.__exit__(exc, value, tb)
        self._closer.cleanup()
        return result

    def close(self):
        self._closer.close()

    def __iter__(self):
        for line in self.file:
            yield line


def NamedTemporaryFile(mode="w+b", buffering=-1, encoding=None, newline=None, suffix=None, prefix=None,
                       dir=None, delete=True, *, errors=None, delete_on_close=True):
    prefix, suffix, dir, output_type = _sanitize_params(prefix, suffix, dir)
    flags = _bin_openflags
    fd, name = _mkstemp_inner(dir, prefix, suffix, flags, output_type)
    try:
        file = open(fd, mode, buffering, encoding, errors, newline)
        file.name = name
        return _TemporaryFileWrapper(file, name, delete, delete_on_close)
    except:
        os.close(fd)
        os.unlink(name)
        raise


def TemporaryFile(mode="w+b", buffering=-1, encoding=None, newline=None, suffix=None, prefix=None,
                  dir=None, *, errors=None):
    prefix, suffix, dir, output_type = _sanitize_params(prefix, suffix, dir)
    flags = _bin_openflags
    fd, name = _mkstemp_inner(dir, prefix, suffix, flags, output_type)
    try:
        os.unlink(name)
        return open(fd, mode, buffering, encoding, errors, newline)
    except:
        os.close(fd)
        raise


class SpooledTemporaryFile:
    _rolled = False

    def __init__(self, max_size=0, mode="w+b", buffering=-1, encoding=None, newline=None,
                 suffix=None, prefix=None, dir=None, *, errors=None):
        if "b" in mode:
            self._file = _obj(io.BytesIO())
        else:
            self._file = _obj(io.StringIO(newline="\n"))
        self._max_size = max_size
        self._rolled = False
        self._TemporaryFileArgs = (mode, buffering, encoding, newline, suffix, prefix, dir, errors)

    def _check(self, file):
        if self._rolled:
            return
        max_size = self._max_size
        if max_size and file.tell() > max_size:
            self.rollover()

    def rollover(self):
        if self._rolled:
            return
        file = self._file
        mode, buffering, encoding, newline, suffix, prefix, dir, errors = self._TemporaryFileArgs
        newfile = TemporaryFile(mode, buffering, encoding, newline, suffix, prefix, dir, errors=errors)
        pos = file.tell()
        if "b" in mode:
            newfile.write(file.getvalue())
        else:
            newfile.write(file.getvalue())
        newfile.seek(pos, 0)
        self._file = newfile
        self._rolled = True

    def __enter__(self):
        if self._file.closed:
            raise ValueError("Cannot enter context with closed file")
        return self

    def __exit__(self, exc, value, tb):
        self._file.close()

    def __iter__(self):
        return self._file.__iter__()

    def __getattr__(self, name):
        return getattr(self._file, name)

    def close(self):
        self._file.close()

    @property
    def closed(self):
        return self._file.closed

    def flush(self):
        self._file.flush()

    def read(self, *args):
        return self._file.read(*args)

    def readline(self, *args):
        return self._file.readline(*args)

    def readlines(self, *args):
        return self._file.readlines(*args)

    def seek(self, *args):
        return self._file.seek(*args)

    def tell(self):
        return self._file.tell()

    def truncate(self, size=None):
        if size is None:
            self._file.truncate()
        else:
            if size > self._max_size:
                self.rollover()
            self._file.truncate(size)

    def write(self, s):
        file = self._file
        rv = file.write(s)
        self._check(file)
        return rv

    def writelines(self, iterable):
        file = self._file
        rv = file.writelines(iterable)
        self._check(file)
        return rv


class TemporaryDirectory:
    def __init__(self, suffix=None, prefix=None, dir=None, ignore_cleanup_errors=False, *, delete=True):
        self.name = mkdtemp(suffix, prefix, dir)
        self._ignore_cleanup_errors = ignore_cleanup_errors
        self._delete = delete
        self._cleaned = False

    def __repr__(self):
        return "<{} {!r}>".format(self.__class__.__name__, self.name)

    def __enter__(self):
        return self.name

    def __exit__(self, exc, value, tb):
        if self._delete:
            self.cleanup()

    def cleanup(self):
        if not self._cleaned:
            self._cleaned = True
            shutil.rmtree(self.name, ignore_errors=self._ignore_cleanup_errors)
