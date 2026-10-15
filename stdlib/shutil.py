"""shutil: high-level file operations (copying, moving, removing trees)."""
import os
import fnmatch
import stat

COPY_BUFSIZE = 65536

__all__ = ["copyfileobj", "copyfile", "copymode", "copystat", "copy", "copy2",
           "copytree", "move", "rmtree", "Error", "SpecialFileError",
           "ExecError", "make_archive", "get_archive_formats",
           "register_archive_format", "unregister_archive_format",
           "get_unpack_formats", "register_unpack_format",
           "unregister_unpack_format", "unpack_archive",
           "ignore_patterns", "chown", "which", "get_terminal_size",
           "SameFileError", "disk_usage"]


class Error(OSError):
    pass


class SameFileError(Error):
    pass


class SpecialFileError(OSError):
    pass


class ExecError(OSError):
    pass


class ReadError(OSError):
    pass


class RegistryError(Exception):
    pass


def copyfileobj(fsrc, fdst, length=0):
    if not length:
        length = COPY_BUFSIZE
    while True:
        buf = fsrc.read(length)
        if not buf:
            break
        fdst.write(buf)


def _samefile(src, dst):
    if isinstance(src, os.DirEntry):
        try:
            return os.path.samestat(src.stat(), os.stat(dst))
        except OSError:
            return False
    try:
        return os.path.samefile(src, dst)
    except OSError:
        return False


def _stat(fn):
    return fn.stat() if isinstance(fn, os.DirEntry) else os.stat(fn)


def _islink(fn):
    return fn.is_symlink() if isinstance(fn, os.DirEntry) else os.path.islink(fn)


def copyfile(src, dst, *, follow_symlinks=True):
    if _samefile(src, dst):
        raise SameFileError("{!r} and {!r} are the same file".format(src, dst))
    for i, fn in enumerate([src, dst]):
        try:
            st = _stat(fn)
        except OSError:
            pass
        else:
            if stat.S_ISFIFO(st.st_mode):
                fn = fn.path if isinstance(fn, os.DirEntry) else fn
                raise SpecialFileError("`%s` is a named pipe" % fn)
    if not follow_symlinks and _islink(src):
        os.symlink(os.readlink(src), dst)
    else:
        with open(src, "rb") as fsrc:
            try:
                with open(dst, "wb") as fdst:
                    copyfileobj(fsrc, fdst)
            except IsADirectoryError as e:
                if not os.path.exists(dst):
                    raise FileNotFoundError("Directory does not exist: {}".format(dst)) from e
                else:
                    raise
    return dst


def copymode(src, dst, *, follow_symlinks=True):
    if not follow_symlinks and _islink(src) and os.path.islink(dst):
        return
    if follow_symlinks or not _islink(src):
        st = _stat(src)
        os.chmod(dst, stat.S_IMODE(st.st_mode))


def copystat(src, dst, *, follow_symlinks=True):
    follow = follow_symlinks or not (_islink(src) and os.path.islink(dst))
    if isinstance(src, os.DirEntry):
        st = src.stat(follow_symlinks=follow)
    else:
        st = os.stat(src, follow_symlinks=follow)
    mode = stat.S_IMODE(st.st_mode)
    if follow:
        os.utime(dst, ns=(st.st_atime_ns, st.st_mtime_ns))
        os.chmod(dst, mode)


def copy(src, dst, *, follow_symlinks=True):
    if os.path.isdir(dst):
        dst = os.path.join(dst, os.path.basename(src))
    copyfile(src, dst, follow_symlinks=follow_symlinks)
    copymode(src, dst, follow_symlinks=follow_symlinks)
    return dst


def copy2(src, dst, *, follow_symlinks=True):
    if os.path.isdir(dst):
        dst = os.path.join(dst, os.path.basename(src))
    copyfile(src, dst, follow_symlinks=follow_symlinks)
    copystat(src, dst, follow_symlinks=follow_symlinks)
    return dst


def ignore_patterns(*patterns):
    def _ignore_patterns(path, names):
        ignored_names = []
        for pattern in patterns:
            ignored_names.extend(fnmatch.filter(names, pattern))
        return set(ignored_names)
    return _ignore_patterns


def _copytree(entries, src, dst, symlinks, ignore, copy_function, ignore_dangling_symlinks, dirs_exist_ok=False):
    if ignore is not None:
        ignored_names = ignore(os.fspath(src), [x.name for x in entries])
    else:
        ignored_names = set()
    os.makedirs(dst, exist_ok=dirs_exist_ok)
    errors = []
    use_srcentry = copy_function is copy2 or copy_function is copy
    for srcentry in entries:
        if srcentry.name in ignored_names:
            continue
        srcname = os.path.join(src, srcentry.name)
        dstname = os.path.join(dst, srcentry.name)
        srcobj = srcentry if use_srcentry else srcname
        try:
            is_symlink = srcentry.is_symlink()
            if is_symlink and os.name == "nt":
                lstat = srcentry.stat(follow_symlinks=False)
            if is_symlink:
                linkto = os.readlink(srcname)
                if symlinks:
                    os.symlink(linkto, dstname)
                    copystat(srcobj, dstname, follow_symlinks=not symlinks)
                else:
                    if not os.path.exists(linkto) and ignore_dangling_symlinks:
                        continue
                    if srcentry.is_dir():
                        copytree(srcobj, dstname, symlinks, ignore, copy_function,
                                 ignore_dangling_symlinks, dirs_exist_ok)
                    else:
                        copy_function(srcobj, dstname)
            elif srcentry.is_dir():
                copytree(srcobj, dstname, symlinks, ignore, copy_function,
                         ignore_dangling_symlinks, dirs_exist_ok)
            else:
                copy_function(srcobj, dstname)
        except Error as err:
            errors.extend(err.args[0])
        except OSError as why:
            errors.append((srcname, dstname, str(why)))
    try:
        copystat(src, dst)
    except OSError as why:
        errors.append((src, dst, str(why)))
    if errors:
        raise Error(errors)
    return dst


def copytree(src, dst, symlinks=False, ignore=None, copy_function=copy2,
             ignore_dangling_symlinks=False, dirs_exist_ok=False):
    with os.scandir(src) as itr:
        entries = list(itr)
    return _copytree(entries=entries, src=src, dst=dst, symlinks=symlinks,
                     ignore=ignore, copy_function=copy_function,
                     ignore_dangling_symlinks=ignore_dangling_symlinks,
                     dirs_exist_ok=dirs_exist_ok)


def _rmtree_islink(st):
    return stat.S_ISLNK(st.st_mode)


def _rmtree_unsafe(path, onexc):
    try:
        with os.scandir(path) as scandir_it:
            entries = list(scandir_it)
    except OSError as err:
        onexc(os.scandir, path, err)
        entries = []
    for entry in entries:
        fullname = entry.path
        try:
            is_dir = entry.is_dir(follow_symlinks=False)
        except OSError:
            is_dir = False
        if is_dir:
            try:
                if entry.is_symlink():
                    raise OSError("Cannot call rmtree on a symbolic link")
            except OSError as err:
                onexc(os.path.islink, fullname, err)
                continue
            _rmtree_unsafe(fullname, onexc)
        else:
            try:
                os.unlink(fullname)
            except OSError as err:
                onexc(os.unlink, fullname, err)
    try:
        os.rmdir(path)
    except OSError as err:
        onexc(os.rmdir, path, err)


def _raise_exc(*args):
    raise args[2]


def rmtree(path, ignore_errors=False, onerror=None, *, onexc=None, dir_fd=None):
    path = os.fspath(path)
    handler = onexc
    if ignore_errors:
        handler = _ignore_exc
    elif onerror is None and onexc is None:
        handler = _raise_exc
    elif onexc is None:
        handler = _OnError(onerror)
    if dir_fd is not None:
        raise NotImplementedError("dir_fd unavailable on this platform")
    try:
        st = os.lstat(path)
    except OSError as err:
        handler(os.lstat, path, err)
        return
    try:
        if _rmtree_islink(st):
            raise OSError("Cannot call rmtree on a symbolic link")
    except OSError as err:
        handler(os.path.islink, path, err)
        return
    return _rmtree_unsafe(path, handler)


def _ignore_exc(*args):
    pass


class _OnError:
    def __init__(self, onerror):
        self.onerror = onerror

    def __call__(self, func, path, exc):
        return self.onerror(func, path, (type(exc), exc, None))


def _basename(path):
    path = os.fspath(path)
    return os.path.basename(path.rstrip("/"))


def move(src, dst, copy_function=copy2):
    real_dst = dst
    if os.path.isdir(dst):
        if _samefile(src, dst) and not os.path.islink(src):
            os.rename(src, dst)
            return
        real_dst = os.path.join(dst, _basename(src))
        if os.path.exists(real_dst):
            raise Error("Destination path '%s' already exists" % real_dst)
    try:
        os.rename(src, real_dst)
    except OSError:
        if os.path.islink(src):
            linkto = os.readlink(src)
            os.symlink(linkto, real_dst)
            os.unlink(src)
        elif os.path.isdir(src):
            if _destinsrc(src, dst):
                raise Error("Cannot move a directory '%s' into itself '%s'." % (src, dst))
            if (_is_immutable(src) or (not os.access(src, os.W_OK) and os.listdir(src) and
                                       False)):
                raise PermissionError("Cannot move the non-empty directory '%s': Lacking write permission to '%s'." % (src, src))
            copytree(src, real_dst, copy_function=copy_function, symlinks=True)
            rmtree(src)
        else:
            copy_function(src, real_dst)
            os.unlink(src)
    return real_dst


def _destinsrc(src, dst):
    src = os.path.abspath(src)
    dst = os.path.abspath(dst)
    if not src.endswith(os.path.sep):
        src += os.path.sep
    if not dst.endswith(os.path.sep):
        dst += os.path.sep
    return dst.startswith(src)


def _is_immutable(src):
    return False


def disk_usage(path):
    st = os.statvfs(path)
    free = st[4] * st[1]
    total = st[2] * st[1]
    used = (st[2] - st[3]) * st[1]
    return _ntuple_diskusage(total, used, free)


class _ntuple_diskusage:
    def __init__(self, total, used, free):
        self.total = total
        self.used = used
        self.free = free

    def __getitem__(self, i):
        return (self.total, self.used, self.free)[i]

    def __iter__(self):
        return iter((self.total, self.used, self.free))

    def __len__(self):
        return 3

    def __eq__(self, other):
        return tuple(self) == tuple(other)

    def __repr__(self):
        return "usage(total=%r, used=%r, free=%r)" % (self.total, self.used, self.free)


def chown(path, user=None, group=None, *, dir_fd=None, follow_symlinks=True):
    if user is None and group is None:
        raise ValueError("user and/or group must be set")
    uid = -1 if user is None else user
    gid = -1 if group is None else group
    if isinstance(user, str) or isinstance(group, str):
        raise LookupError("user/group names need the pwd and grp modules")
    os.chown(path, uid, gid)


def get_terminal_size(fallback=(80, 24)):
    try:
        columns = int(os.environ["COLUMNS"])
    except (KeyError, ValueError):
        columns = 0
    try:
        lines = int(os.environ["LINES"])
    except (KeyError, ValueError):
        lines = 0
    if columns <= 0 or lines <= 0:
        try:
            size = os.get_terminal_size(1)
        except (AttributeError, ValueError, OSError):
            size = os.terminal_size(fallback[0], fallback[1])
        if columns <= 0:
            columns = size.columns or fallback[0]
        if lines <= 0:
            lines = size.lines or fallback[1]
    return os.terminal_size(columns, lines)


def _access_check(fn, mode):
    return (os.path.exists(fn) and os.access(fn, mode) and not os.path.isdir(fn))


def which(cmd, mode=os.F_OK | os.X_OK, path=None):
    if os.path.dirname(cmd):
        if _access_check(cmd, mode):
            return cmd
        return None
    if path is None:
        path = os.environ.get("PATH", None)
        if path is None:
            path = os.defpath
    if not path:
        return None
    path = os.fsdecode(path)
    path = path.split(os.pathsep)
    seen = set()
    for dir in path:
        normdir = os.path.normcase(dir)
        if normdir not in seen:
            seen.add(normdir)
            name = os.path.join(dir, cmd)
            if _access_check(name, mode):
                return name
    return None


def _make_tarball(base_name, base_dir, compress="gzip", verbose=0, dry_run=0, owner=None, group=None, logger=None, root_dir=None):
    import tarfile
    tar_compression = {"gzip": "gz", None: ""}[compress]
    compress_ext = ".gz" if compress == "gzip" else ""
    archive_name = base_name + ".tar" + compress_ext
    archive_dir = os.path.dirname(archive_name)
    if archive_dir and not os.path.exists(archive_dir):
        os.makedirs(archive_dir)
    if not dry_run:
        tar = tarfile.open(archive_name, "w|" .replace("|", ":") + tar_compression)
        try:
            arcname = base_dir
            src = base_dir
            if root_dir is not None:
                src = os.path.join(root_dir, base_dir)
            tar.add(src, arcname=arcname)
        finally:
            tar.close()
    return archive_name


def _make_zipfile(base_name, base_dir, verbose=0, dry_run=0, logger=None, owner=None, group=None, root_dir=None):
    import zipfile
    zip_filename = base_name + ".zip"
    archive_dir = os.path.dirname(base_name)
    if archive_dir and not os.path.exists(archive_dir):
        os.makedirs(archive_dir)
    if not dry_run:
        with zipfile.ZipFile(zip_filename, "w", compression=zipfile.ZIP_DEFLATED) as zf:
            def full(p):
                return os.path.join(root_dir, p) if root_dir is not None else p
            path = os.path.normpath(base_dir)
            if path != os.curdir:
                zf.write(full(path), path)
            for dirpath, dirnames, filenames in os.walk(full(base_dir)):
                rel = os.path.relpath(dirpath, root_dir) if root_dir is not None else dirpath
                for name in sorted(dirnames):
                    path = os.path.normpath(os.path.join(rel, name))
                    zf.write(full(path), path)
                for name in filenames:
                    path = os.path.normpath(os.path.join(rel, name))
                    if os.path.isfile(full(path)):
                        zf.write(full(path), path)
    return zip_filename


_ARCHIVE_FORMATS = {
    "tar": (_make_tarball, [("compress", None)], "uncompressed tar file"),
    "gztar": (_make_tarball, [("compress", "gzip")], "gzip'ed tar-file"),
    "zip": (_make_zipfile, [], "ZIP file"),
}


def _ensure_directory(path):
    dirname = os.path.dirname(path)
    if not os.path.isdir(dirname):
        os.makedirs(dirname)


def _unpack_zipfile(filename, extract_dir):
    import zipfile
    if not zipfile.is_zipfile(filename):
        raise ReadError("%s is not a zip file" % filename)
    zip = zipfile.ZipFile(filename)
    try:
        for info in zip.infolist():
            name = info.filename
            if name.startswith("/") or ".." in name.split("/"):
                continue
            targetpath = os.path.join(extract_dir, *name.split("/"))
            if not targetpath:
                continue
            _ensure_directory(targetpath)
            if not name.endswith("/"):
                with open(targetpath, "wb") as f:
                    f.write(zip.read(info.filename))
    finally:
        zip.close()


def _unpack_tarfile(filename, extract_dir, *, filter=None):
    import tarfile
    try:
        tarobj = tarfile.open(filename)
    except tarfile.TarError:
        raise ReadError("%s is not a compressed or uncompressed tar file" % filename)
    try:
        tarobj.extractall(extract_dir, filter=filter)
    finally:
        tarobj.close()


_UNPACK_FORMATS = {
    "tar": ([".tar"], _unpack_tarfile, [], "uncompressed tar file"),
    "gztar": ([".tar.gz", ".tgz"], _unpack_tarfile, [], "gzip'ed tar-file"),
    "zip": ([".zip"], _unpack_zipfile, [], "ZIP file"),
}


def get_archive_formats():
    formats = [(name, registry[2]) for name, registry in _ARCHIVE_FORMATS.items()]
    formats.sort()
    return formats


def register_archive_format(name, function, extra_args=None, description=""):
    if extra_args is None:
        extra_args = []
    if not callable(function):
        raise TypeError("The %s object is not callable" % function)
    if not isinstance(extra_args, (tuple, list)):
        raise TypeError("extra_args needs to be a sequence")
    for element in extra_args:
        if not isinstance(element, (tuple, list)) or len(element) != 2:
            raise TypeError("extra_args elements are : (arg_name, value)")
    _ARCHIVE_FORMATS[name] = (function, extra_args, description)


def unregister_archive_format(name):
    del _ARCHIVE_FORMATS[name]


def make_archive(base_name, format, root_dir=None, base_dir=None, verbose=0, dry_run=0, owner=None, group=None, logger=None):
    save_cwd = None
    if root_dir is not None:
        base_name = os.path.abspath(base_name)
    if base_dir is None:
        base_dir = os.curdir
    kwargs = {"dry_run": dry_run, "logger": logger, "owner": owner, "group": group, "root_dir": root_dir}
    try:
        format_info = _ARCHIVE_FORMATS[format]
    except KeyError:
        raise ValueError("unknown archive format '%s'" % format) from None
    func = format_info[0]
    for arg, val in format_info[1]:
        kwargs[arg] = val
    return func(base_name, base_dir, **kwargs)


def get_unpack_formats():
    formats = [(name, info[0], info[3]) for name, info in _UNPACK_FORMATS.items()]
    formats.sort()
    return formats


def register_unpack_format(name, extensions, function, extra_args=None, description=""):
    if extra_args is None:
        extra_args = []
    _UNPACK_FORMATS[name] = (extensions, function, extra_args, description)


def unregister_unpack_format(name):
    del _UNPACK_FORMATS[name]


def _find_unpack_format(filename):
    for name, info in _UNPACK_FORMATS.items():
        for extension in info[0]:
            if filename.endswith(extension):
                return name
    return None


def unpack_archive(filename, extract_dir=None, format=None, *, filter=None):
    filename = os.fspath(filename)
    if extract_dir is None:
        extract_dir = os.getcwd()
    extract_dir = os.fspath(extract_dir)
    if format is not None:
        try:
            format_info = _UNPACK_FORMATS[format]
        except KeyError:
            raise ValueError("Unknown unpack format '{0}'".format(format)) from None
        func = format_info[1]
    else:
        fmt = _find_unpack_format(filename)
        if fmt is None:
            raise ReadError("Unknown archive format '{0}'".format(filename))
        func = _UNPACK_FORMATS[fmt][1]
    if func is _unpack_tarfile:
        func(filename, extract_dir, filter=filter)
    else:
        func(filename, extract_dir)
