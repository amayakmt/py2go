"""tarfile: read and write tar archives (ustar, GNU and pax; gzip compression)."""
import io
import os
import stat
import struct
import sys
import time
import gzip as _gzip

__all__ = ["TarFile", "TarInfo", "is_tarfile", "TarError", "ReadError", "CompressionError", "StreamError",
           "ExtractError", "HeaderError", "ENCODING", "USTAR_FORMAT", "GNU_FORMAT", "PAX_FORMAT", "DEFAULT_FORMAT",
           "open", "FilterError", "AbsolutePathError", "OutsideDestinationError", "SpecialFileError",
           "AbsoluteLinkError", "LinkOutsideDestinationError", "data_filter", "tar_filter", "fully_trusted_filter"]

NUL = b"\0"
BLOCKSIZE = 512
RECORDSIZE = BLOCKSIZE * 20
GNU_MAGIC = b"ustar  \0"
POSIX_MAGIC = b"ustar\x0000"

LENGTH_NAME = 100
LENGTH_LINK = 100
LENGTH_PREFIX = 155

REGTYPE = b"0"
AREGTYPE = b"\0"
LNKTYPE = b"1"
SYMTYPE = b"2"
CHRTYPE = b"3"
BLKTYPE = b"4"
DIRTYPE = b"5"
FIFOTYPE = b"6"
CONTTYPE = b"7"
GNUTYPE_LONGNAME = b"L"
GNUTYPE_LONGLINK = b"K"
GNUTYPE_SPARSE = b"S"
XHDTYPE = b"x"
XGLTYPE = b"g"
SOLARIS_XHDTYPE = b"X"

USTAR_FORMAT = 0
GNU_FORMAT = 1
PAX_FORMAT = 2
DEFAULT_FORMAT = PAX_FORMAT

SUPPORTED_TYPES = (REGTYPE, AREGTYPE, LNKTYPE, SYMTYPE, DIRTYPE, FIFOTYPE, CONTTYPE, CHRTYPE, BLKTYPE, GNUTYPE_LONGNAME, GNUTYPE_LONGLINK)
REGULAR_TYPES = (REGTYPE, AREGTYPE, CONTTYPE)

ENCODING = "utf-8"


class TarError(Exception):
    pass


class ExtractError(TarError):
    pass


class ReadError(TarError):
    pass


class CompressionError(TarError):
    pass


class StreamError(TarError):
    pass


class HeaderError(TarError):
    pass


class EmptyHeaderError(HeaderError):
    pass


class TruncatedHeaderError(HeaderError):
    pass


class EOFHeaderError(HeaderError):
    pass


class InvalidHeaderError(HeaderError):
    pass


class SubsequentHeaderError(HeaderError):
    pass


class FilterError(TarError):
    def __init__(self, tarinfo):
        self.tarinfo = tarinfo
        super().__init__(tarinfo)


class AbsolutePathError(FilterError):
    def __str__(self):
        return "member " + repr(self.tarinfo.name) + " has an absolute path"


class OutsideDestinationError(FilterError):
    def __init__(self, tarinfo, path):
        self.tarinfo = tarinfo
        self._path = path
        super().__init__(tarinfo)

    def __str__(self):
        return "%r would be extracted to %r, which is outside the destination" % (self.tarinfo.name, self._path)


class SpecialFileError(FilterError):
    def __str__(self):
        return "%r is a special file" % self.tarinfo.name


class AbsoluteLinkError(FilterError):
    def __str__(self):
        return "%r is a link to an absolute path" % self.tarinfo.name


class LinkOutsideDestinationError(FilterError):
    def __init__(self, tarinfo, path):
        self.tarinfo = tarinfo
        self._path = path
        super().__init__(tarinfo)

    def __str__(self):
        return "%r would link to %r, which is outside the destination" % (self.tarinfo.name, self._path)


def stn(s, length, encoding, errors):
    s = s.encode(encoding, errors)
    return s[:length] + (length - len(s)) * NUL


def nts(s, encoding, errors):
    p = s.find(b"\0")
    if p != -1:
        s = s[:p]
    return s.decode(encoding, errors)


def nti(s):
    if s[0] in (0o200, 0o377):
        n = 0
        for i in range(len(s) - 1):
            n <<= 8
            n += s[i + 1]
        if s[0] == 0o377:
            n = -(256 ** (len(s) - 1) - n)
    else:
        try:
            s = nts(s, "ascii", "strict")
            n = int(s.strip() or "0", 8)
        except ValueError:
            raise InvalidHeaderError("invalid header")
    return n


def itn(n, digits=8, format=DEFAULT_FORMAT):
    n = int(n)
    if 0 <= n < 8 ** (digits - 1):
        s = bytes("%0*o" % (digits - 1, n), "ascii") + NUL
    elif format == GNU_FORMAT and -256 ** (digits - 1) <= n < 256 ** (digits - 1):
        if n >= 0:
            s = bytearray([0o200])
        else:
            s = bytearray([0o377])
            n = 256 ** digits + n
        for i in range(digits - 1):
            s.insert(1, n & 0o377)
            n >>= 8
    else:
        raise ValueError("overflow in number field")
    return bytes(s)


def calc_chksums(buf):
    unsigned_chksum = 256 + sum(struct.unpack_from("148B8x356B", buf))
    signed_chksum = 256 + sum(struct.unpack_from("148b8x356b", buf))
    return unsigned_chksum, signed_chksum


class TarInfo:
    def __init__(self, name=""):
        self.name = name
        self.mode = 0o644
        self.uid = 0
        self.gid = 0
        self.size = 0
        self.mtime = 0
        self.chksum = 0
        self.type = REGTYPE
        self.linkname = ""
        self.uname = ""
        self.gname = ""
        self.devmajor = 0
        self.devminor = 0
        self.offset = 0
        self.offset_data = 0
        self.sparse = None
        self.pax_headers = {}
        self._data = b""

    @property
    def path(self):
        return self.name

    @path.setter
    def path(self, name):
        self.name = name

    @property
    def linkpath(self):
        return self.linkname

    @linkpath.setter
    def linkpath(self, linkname):
        self.linkname = linkname

    def __repr__(self):
        return "<%s %r at %#x>" % (self.__class__.__name__, self.name, id(self))

    def replace(self, *, name=None, mtime=None, mode=None, linkname=None, uid=None, gid=None, uname=None, gname=None,
                deep=True):
        new = TarInfo(self.name)
        new.mode = self.mode
        new.uid = self.uid
        new.gid = self.gid
        new.size = self.size
        new.mtime = self.mtime
        new.chksum = self.chksum
        new.type = self.type
        new.linkname = self.linkname
        new.uname = self.uname
        new.gname = self.gname
        new.devmajor = self.devmajor
        new.devminor = self.devminor
        new.offset = self.offset
        new.offset_data = self.offset_data
        new.pax_headers = dict(self.pax_headers)
        new._data = self._data
        if name is not None:
            new.name = name
        if mtime is not None:
            new.mtime = mtime
        if mode is not None:
            new.mode = mode
        if linkname is not None:
            new.linkname = linkname
        if uid is not None:
            new.uid = uid
        if gid is not None:
            new.gid = gid
        if uname is not None:
            new.uname = uname
        if gname is not None:
            new.gname = gname
        return new

    def get_info(self):
        info = {"name": self.name, "mode": self.mode & 0o7777, "uid": self.uid, "gid": self.gid, "size": self.size,
                "mtime": self.mtime, "chksum": self.chksum, "type": self.type, "linkname": self.linkname,
                "uname": self.uname, "gname": self.gname, "devmajor": self.devmajor, "devminor": self.devminor}
        if info["type"] == DIRTYPE and not info["name"].endswith("/"):
            info["name"] += "/"
        return info

    def tobuf(self, format=DEFAULT_FORMAT, encoding=ENCODING, errors="surrogateescape"):
        info = self.get_info()
        for name, value in info.items():
            if value is None:
                raise ValueError("%s may not be None" % name)
        if format == USTAR_FORMAT:
            return self.create_ustar_header(info, encoding, errors)
        elif format == GNU_FORMAT:
            return self.create_gnu_header(info, encoding, errors)
        elif format == PAX_FORMAT:
            return self.create_pax_header(info, encoding)
        else:
            raise ValueError("invalid format")

    def create_ustar_header(self, info, encoding, errors):
        info["magic"] = POSIX_MAGIC
        if len(info["linkname"].encode(encoding, errors)) > LENGTH_LINK:
            raise ValueError("linkname is too long")
        if len(info["name"].encode(encoding, errors)) > LENGTH_NAME:
            info["prefix"], info["name"] = self._posix_split_name(info["name"], encoding, errors)
        return self._create_header(info, USTAR_FORMAT, encoding, errors)

    def create_gnu_header(self, info, encoding, errors):
        info["magic"] = GNU_MAGIC
        buf = b""
        if len(info["linkname"].encode(encoding, errors)) > LENGTH_LINK:
            buf += self._create_gnu_long_header(info["linkname"], GNUTYPE_LONGLINK, encoding, errors)
        if len(info["name"].encode(encoding, errors)) > LENGTH_NAME:
            buf += self._create_gnu_long_header(info["name"], GNUTYPE_LONGNAME, encoding, errors)
        return buf + self._create_header(info, GNU_FORMAT, encoding, errors)

    def create_pax_header(self, info, encoding):
        info["magic"] = POSIX_MAGIC
        pax_headers = dict(self.pax_headers)
        for name, hname, length in (("name", "path", LENGTH_NAME), ("linkname", "linkpath", LENGTH_LINK),
                                    ("uname", "uname", 32), ("gname", "gname", 32)):
            if hname in pax_headers:
                continue
            try:
                info[name].encode("ascii", "strict")
            except UnicodeEncodeError:
                pax_headers[hname] = info[name]
                continue
            if len(info[name]) > length:
                pax_headers[hname] = info[name]
        for name, digits in (("uid", 8), ("gid", 8), ("size", 12), ("mtime", 12)):
            if name in pax_headers:
                info[name] = 0
                continue
            val = info[name]
            if not 0 <= val < 8 ** (digits - 1) or isinstance(val, float):
                pax_headers[name] = str(val)
                info[name] = 0
        if pax_headers:
            buf = self._create_pax_generic_header(pax_headers, XHDTYPE, encoding)
        else:
            buf = b""
        return buf + self._create_header(info, USTAR_FORMAT, "ascii", "replace")

    @classmethod
    def create_pax_global_header(cls, pax_headers):
        return cls._create_pax_generic_header(pax_headers, XGLTYPE, "utf-8")

    def _posix_split_name(self, name, encoding, errors):
        components = name.split("/")
        for i in range(1, len(components)):
            prefix = "/".join(components[:i])
            name = "/".join(components[i:])
            if len(prefix.encode(encoding, errors)) <= LENGTH_PREFIX and len(name.encode(encoding, errors)) <= LENGTH_NAME:
                break
        else:
            raise ValueError("name is too long")
        return prefix, name

    @staticmethod
    def _create_header(info, format, encoding, errors):
        has_device_fields = info.get("type") in (CHRTYPE, BLKTYPE)
        if has_device_fields:
            devmajor = itn(info.get("devmajor", 0), 8, format)
            devminor = itn(info.get("devminor", 0), 8, format)
        else:
            devmajor = stn("", 8, encoding, errors)
            devminor = stn("", 8, encoding, errors)
        filetype = info.get("type", REGTYPE)
        if filetype is None:
            raise ValueError("TarInfo.type must not be None")
        parts = [stn(info.get("name", ""), 100, encoding, errors), itn(info.get("mode", 0) & 0o7777, 8, format),
                 itn(info.get("uid", 0), 8, format), itn(info.get("gid", 0), 8, format),
                 itn(info.get("size", 0), 12, format), itn(info.get("mtime", 0), 12, format), b"        ", filetype,
                 stn(info.get("linkname", ""), 100, encoding, errors), info.get("magic", POSIX_MAGIC),
                 stn(info.get("uname", ""), 32, encoding, errors), stn(info.get("gname", ""), 32, encoding, errors),
                 devmajor, devminor, stn(info.get("prefix", ""), 155, encoding, errors)]
        buf = b"".join(parts)
        buf = buf + NUL * (BLOCKSIZE - len(buf))
        chksum = sum(buf) + 0
        buf = buf[:-364] + bytes("%06o\0" % chksum, "ascii") + buf[-357:]
        return buf

    @staticmethod
    def _create_payload(payload):
        blocks, remainder = divmod(len(payload), BLOCKSIZE)
        if remainder > 0:
            payload += (BLOCKSIZE - remainder) * NUL
        return payload

    @classmethod
    def _create_gnu_long_header(cls, name, type, encoding, errors):
        name = name.encode(encoding, errors) + NUL
        info = {}
        info["name"] = "././@LongLink"
        info["type"] = type
        info["size"] = len(name)
        info["magic"] = GNU_MAGIC
        return cls._create_header(info, USTAR_FORMAT, encoding, errors) + cls._create_payload(name)

    @classmethod
    def _create_pax_generic_header(cls, pax_headers, type, encoding):
        binary = False
        for keyword, value in pax_headers.items():
            try:
                value.encode("utf-8", "strict")
            except UnicodeEncodeError:
                binary = True
                break
        records = b""
        if binary:
            records += b"21 hdrcharset=BINARY\n"
        for keyword, value in pax_headers.items():
            keyword = keyword.encode("utf-8")
            if binary:
                value = value.encode(encoding, "surrogateescape")
            else:
                value = value.encode("utf-8")
            l = len(keyword) + len(value) + 3
            n = p = 0
            while True:
                n = l + len(str(p))
                if n == p:
                    break
                p = n
            records += bytes(str(p), "ascii") + b" " + keyword + b"=" + value + b"\n"
        info = {}
        info["name"] = "././@PaxHeader"
        info["type"] = type
        info["size"] = len(records)
        info["magic"] = POSIX_MAGIC
        return cls._create_header(info, USTAR_FORMAT, "ascii", "replace") + cls._create_payload(records)

    @classmethod
    def frombuf(cls, buf, encoding, errors):
        if len(buf) == 0:
            raise EmptyHeaderError("empty header")
        if len(buf) != BLOCKSIZE:
            raise TruncatedHeaderError("truncated header")
        if buf.count(NUL) == BLOCKSIZE:
            raise EOFHeaderError("end of file header")
        chksum = nti(buf[148:156])
        if chksum not in calc_chksums(buf):
            raise InvalidHeaderError("bad checksum")
        obj = cls()
        obj.name = nts(buf[0:100], encoding, errors)
        obj.mode = nti(buf[100:108])
        obj.uid = nti(buf[108:116])
        obj.gid = nti(buf[116:124])
        obj.size = nti(buf[124:136])
        obj.mtime = nti(buf[136:148])
        obj.chksum = chksum
        obj.type = buf[156:157]
        obj.linkname = nts(buf[157:257], encoding, errors)
        obj.uname = nts(buf[265:297], encoding, errors)
        obj.gname = nts(buf[297:329], encoding, errors)
        obj.devmajor = nti(buf[329:337])
        obj.devminor = nti(buf[337:345])
        prefix = nts(buf[345:500], encoding, errors)
        if obj.type == AREGTYPE and obj.name.endswith("/"):
            obj.type = DIRTYPE
        if obj.type != GNUTYPE_SPARSE and buf[257:265] != GNU_MAGIC:
            if prefix:
                obj.name = prefix + "/" + obj.name
        return obj

    def isreg(self):
        return self.type in REGULAR_TYPES

    def isfile(self):
        return self.isreg()

    def isdir(self):
        return self.type == DIRTYPE

    def issym(self):
        return self.type == SYMTYPE

    def islnk(self):
        return self.type == LNKTYPE

    def ischr(self):
        return self.type == CHRTYPE

    def isblk(self):
        return self.type == BLKTYPE

    def isfifo(self):
        return self.type == FIFOTYPE

    def isdev(self):
        return self.type in (CHRTYPE, BLKTYPE, FIFOTYPE)


def fully_trusted_filter(member, dest_path):
    return member


def tar_filter(member, dest_path):
    name = member.name
    dest_path = os.path.realpath(dest_path)
    if name.startswith(("/", os.sep)):
        name = member.path.lstrip("/" + os.sep)
    if os.path.isabs(name):
        raise AbsolutePathError(member)
    target_path = os.path.realpath(os.path.join(dest_path, name))
    if os.path.commonpath([target_path, dest_path]) != dest_path:
        raise OutsideDestinationError(member, target_path)
    mode = member.mode
    new_attrs = {}
    if mode is not None:
        mode = mode & 0o755
        mode &= ~0o022
        if member.isreg() or member.islnk():
            if not mode & 0o100:
                mode &= ~0o011
        new_attrs["mode"] = mode
    if name != member.name:
        new_attrs["name"] = name
    return member.replace(**new_attrs)


def data_filter(member, dest_path):
    new = tar_filter(member, dest_path)
    if new.type in (CHRTYPE, BLKTYPE, FIFOTYPE):
        raise SpecialFileError(member)
    if new.islnk() or new.issym():
        if os.path.isabs(new.linkname):
            raise AbsoluteLinkError(member)
    mode = new.mode
    if new.isreg():
        mode = mode | 0o600
    return new.replace(uid=None, gid=None, uname=None, gname=None, mode=mode) if False else new.replace(mode=mode)


class TarFile:
    def __init__(self, name=None, mode="r", fileobj=None, format=None, tarinfo=None, dereference=None,
                 ignore_zeros=None, encoding=None, errors="surrogateescape", pax_headers=None, debug=None,
                 errorlevel=None, copybufsize=None, stream=False):
        modes = {"r": "rb", "a": "r+b", "w": "wb", "x": "xb"}
        if mode not in modes:
            raise ValueError("mode must be 'r', 'a', 'w' or 'x'")
        self.mode = mode
        self._mode = modes[mode]
        self._own_file = False
        self._compression = ""
        self._gz_path = ""
        self._gz_level = 9
        self._append_base = 0
        if not fileobj:
            if self.mode == "a" and not os.path.exists(name):
                self.mode = "w"
                self._mode = "wb"
            fileobj = io.open(name, self._mode)
            self._own_file = True
        self.name = os.path.abspath(name) if name else None
        self.fileobj = fileobj
        self.format = DEFAULT_FORMAT if format is None else format
        self.dereference = False if dereference is None else dereference
        self.ignore_zeros = False if ignore_zeros is None else ignore_zeros
        self.encoding = ENCODING if encoding is None else encoding
        self.errors = errors
        self.extraction_filter = None
        if pax_headers is not None and self.format == PAX_FORMAT:
            self.pax_headers = pax_headers
        else:
            self.pax_headers = {}
        self.debug = 0 if debug is None else debug
        self.errorlevel = 1 if errorlevel is None else errorlevel
        self.copybufsize = copybufsize
        self.closed = False
        self.members = []
        self._loaded = False
        self.offset = 0
        self.inodes = {}
        self._outbuf = []
        self._data = b""
        try:
            if self.mode == "r":
                self._load_all()
            elif self.mode == "a":
                self._load_all()
                self._append_base = self._append_offset()
                self.offset = self._append_base
            if self.mode in ("a", "w", "x"):
                self._loaded = True
                if self.pax_headers:
                    buf = self.tarinfo.create_pax_global_header(self.pax_headers.copy())
                    self._write(buf)
        except:
            if self._own_file:
                self.fileobj.close()
            self.closed = True
            raise

    tarinfo = TarInfo

    def _write(self, data):
        self._outbuf.append(data)
        self.offset += len(data)

    def _append_offset(self):
        if not self.members:
            return 0
        last = self.members[-1]
        return last.offset_data + (_blocks(last.size) if last.isreg() or last.type not in SUPPORTED_TYPES else 0)

    @classmethod
    def open(cls, name=None, mode="r", fileobj=None, bufsize=RECORDSIZE, **kwargs):
        if not name and not fileobj:
            raise ValueError("nothing to open")
        if mode in ("r", "r:*"):
            if fileobj is not None:
                saved = fileobj.tell()
                head = fileobj.read(2)
                fileobj.seek(saved)
            else:
                with io.open(name, "rb") as f:
                    head = f.read(2)
            if head == b"\x1f\x8b":
                return cls.gzopen(name, "r", fileobj, **kwargs)
            return cls.taropen(name, "r", fileobj, **kwargs)
        elif ":" in mode:
            filemode, comptype = mode.split(":", 1)
            filemode = filemode or "r"
            comptype = comptype or "tar"
            if comptype == "gz":
                return cls.gzopen(name, filemode, fileobj, **kwargs)
            if comptype == "tar":
                return cls.taropen(name, filemode, fileobj, **kwargs)
            if comptype in ("bz2", "xz"):
                raise CompressionError(comptype + " module is not available")
            raise ValueError("unknown compression type %r" % comptype)
        elif "|" in mode:
            raise ValueError("stream mode '|' is not supported")
        elif mode in ("a", "w", "x"):
            return cls.taropen(name, mode, fileobj, **kwargs)
        raise ValueError("undiscernible mode")

    @classmethod
    def taropen(cls, name, mode="r", fileobj=None, **kwargs):
        if mode not in ("r", "a", "w", "x"):
            raise ValueError("mode must be 'r', 'a', 'w' or 'x'")
        return cls(name, mode, fileobj, **kwargs)

    @classmethod
    def gzopen(cls, name, mode="r", fileobj=None, compresslevel=9, **kwargs):
        if mode not in ("r", "w", "x"):
            raise ValueError("mode must be 'r', 'w' or 'x'")
        if mode == "r":
            if fileobj is None:
                fileobj = io.open(name, "rb")
                own = True
            else:
                own = False
            raw = fileobj.read()
            if own:
                fileobj.close()
            try:
                data = _gzip.decompress(raw)
            except (OSError, EOFError) as e:
                raise ReadError("not a gzip file") from e
            t = cls(name, "r", io.BytesIO(data), **kwargs)
            t._compression = "gz"
            return t
        t = cls(name, mode, io.BytesIO() if fileobj is None else fileobj, **kwargs)
        t._compression = "gz"
        t._gz_level = compresslevel
        t._gz_path = name if fileobj is None else ""
        return t

    def _load_all(self):
        self.fileobj.seek(0)
        data = self.fileobj.read()
        self._data = data
        self.members = []
        pos = 0
        pax_next = {}
        longname = None
        longlink = None
        while pos < len(data):
            buf = data[pos:pos + BLOCKSIZE]
            try:
                tarinfo = TarInfo.frombuf(buf, self.encoding, self.errors)
            except EOFHeaderError:
                if self.ignore_zeros:
                    pos += BLOCKSIZE
                    continue
                break
            except EmptyHeaderError:
                break
            except InvalidHeaderError as e:
                if pos == 0:
                    raise ReadError("file could not be opened successfully:\n- method tar: ReadError('%s')" % str(e)) from e
                break
            except TruncatedHeaderError as e:
                if pos == 0:
                    raise ReadError(str(e)) from e
                break
            tarinfo.offset = pos
            pos += BLOCKSIZE
            tarinfo.offset_data = pos
            if tarinfo.type in (XHDTYPE, XGLTYPE, SOLARIS_XHDTYPE):
                payload = data[pos:pos + tarinfo.size]
                pos += _blocks(tarinfo.size)
                pax = _parse_pax(payload, self.encoding, self.errors)
                if tarinfo.type == XGLTYPE:
                    self.pax_headers.update(pax)
                else:
                    pax_next.update(pax)
                continue
            if tarinfo.type == GNUTYPE_LONGNAME:
                longname = nts(data[pos:pos + tarinfo.size], self.encoding, self.errors)
                pos += _blocks(tarinfo.size)
                continue
            if tarinfo.type == GNUTYPE_LONGLINK:
                longlink = nts(data[pos:pos + tarinfo.size], self.encoding, self.errors)
                pos += _blocks(tarinfo.size)
                continue
            if tarinfo.isdir():
                tarinfo.name = tarinfo.name.rstrip("/")
            if longname is not None:
                tarinfo.name = longname
                longname = None
            if longlink is not None:
                tarinfo.linkname = longlink
                longlink = None
            if tarinfo.isdir():
                tarinfo.name = tarinfo.name.rstrip("/")
            if pax_next:
                _apply_pax(tarinfo, pax_next)
                tarinfo.pax_headers = dict(pax_next)
                pax_next = {}
            if tarinfo.isreg() or tarinfo.type not in SUPPORTED_TYPES:
                pos += _blocks(tarinfo.size)
            elif tarinfo.islnk() or tarinfo.issym() or tarinfo.isdir() or tarinfo.isdev():
                pass
            self.members.append(tarinfo)
        if not self.members and len(data) > 0 and not data.strip(b"\0"):
            pass
        if not self.members and len(data) == 0 and self.mode == "r":
            raise ReadError("empty file")
        if not self.members and len(data) > 0 and data.strip(b"\0") != b"":
            raise ReadError("file could not be opened successfully:\n- method tar: ReadError('invalid header')")
        self._loaded = True

    def getmember(self, name):
        tarinfo = self._getmember(name.rstrip("/"))
        if tarinfo is None:
            raise KeyError("filename %r not found" % name)
        return tarinfo

    def getmembers(self):
        return self.members

    def getnames(self):
        return [tarinfo.name for tarinfo in self.getmembers()]

    def _getmember(self, name, tarinfo=None, normalize=False):
        members = self.members
        for member in reversed(members):
            if name == member.name.rstrip("/") or name == member.name:
                return member
        return None

    def __iter__(self):
        return iter(list(self.members))

    def next(self):
        raise NotImplementedError("next() is not supported; iterate over the TarFile")

    def gettarinfo(self, name=None, arcname=None, fileobj=None):
        if arcname is None:
            arcname = name
        drv, arcname = os.path.splitdrive(arcname)
        arcname = arcname.replace(os.sep, "/")
        arcname = arcname.lstrip("/")
        tarinfo = self.tarinfo()
        if name is not None:
            if self.dereference:
                statres = os.stat(name)
            else:
                statres = os.lstat(name)
        stmd = statres.st_mode
        if stat.S_ISREG(stmd):
            type = REGTYPE
        elif stat.S_ISDIR(stmd):
            type = DIRTYPE
        elif stat.S_ISFIFO(stmd):
            type = FIFOTYPE
        elif stat.S_ISLNK(stmd):
            type = SYMTYPE
            tarinfo.linkname = os.readlink(name)
        elif stat.S_ISCHR(stmd):
            type = CHRTYPE
        elif stat.S_ISBLK(stmd):
            type = BLKTYPE
        else:
            return None
        tarinfo.name = arcname
        tarinfo.mode = stmd
        tarinfo.uid = statres.st_uid
        tarinfo.gid = statres.st_gid
        if type == REGTYPE:
            tarinfo.size = statres.st_size
        else:
            tarinfo.size = 0
        tarinfo.mtime = statres.st_mtime
        tarinfo.type = type
        try:
            import pwd
            tarinfo.uname = pwd.getpwuid(tarinfo.uid)[0]
        except (KeyError, ImportError):
            pass
        try:
            import grp
            tarinfo.gname = grp.getgrgid(tarinfo.gid)[0]
        except (KeyError, ImportError):
            pass
        return tarinfo

    def list(self, verbose=True, *, members=None):
        if members is None:
            members = self
        for tarinfo in members:
            if verbose:
                print(stat.filemode(tarinfo.mode), end=" ")
                print("%s/%s" % (tarinfo.uname or tarinfo.uid, tarinfo.gname or tarinfo.gid), end=" ")
                if tarinfo.ischr() or tarinfo.isblk():
                    print("%10s" % ("%d,%d" % (tarinfo.devmajor, tarinfo.devminor)), end=" ")
                else:
                    print("%10d" % tarinfo.size, end=" ")
                t = time.localtime(tarinfo.mtime)
                print("%d-%02d-%02d %02d:%02d:%02d" % (t[0], t[1], t[2], t[3], t[4], t[5]), end=" ")
            print(tarinfo.name + ("/" if tarinfo.isdir() else ""), end=" ")
            if verbose:
                if tarinfo.issym():
                    print("->", tarinfo.linkname, end="")
                if tarinfo.islnk():
                    print("link to", tarinfo.linkname, end="")
            print()

    def add(self, name, arcname=None, recursive=True, *, filter=None):
        self._check("awx")
        if arcname is None:
            arcname = name
        tarinfo = self.gettarinfo(name, arcname)
        if tarinfo is None:
            return
        if filter is not None:
            tarinfo = filter(tarinfo)
            if tarinfo is None:
                return
        if tarinfo.isreg():
            with io.open(name, "rb") as f:
                self.addfile(tarinfo, f)
        elif tarinfo.isdir():
            self.addfile(tarinfo)
            if recursive:
                for f in sorted(os.listdir(name)):
                    self.add(os.path.join(name, f), os.path.join(arcname, f), recursive, filter=filter)
        else:
            self.addfile(tarinfo)

    def addfile(self, tarinfo, fileobj=None):
        self._check("awx")
        tarinfo = tarinfo.replace()
        buf = tarinfo.tobuf(self.format, self.encoding, self.errors)
        self._write(buf)
        if fileobj is not None:
            data = fileobj.read(tarinfo.size) if tarinfo.size else b""
            if len(data) != tarinfo.size:
                raise OSError("unexpected end of data")
            self._write(data)
            blocks, remainder = divmod(tarinfo.size, BLOCKSIZE)
            if remainder > 0:
                self._write(NUL * (BLOCKSIZE - remainder))
        self.members.append(tarinfo)

    def _check(self, mode=None):
        if self.closed:
            raise OSError("%s is closed" % self.__class__.__name__)
        if mode is not None and self.mode not in mode:
            raise OSError("bad operation for mode %r" % self.mode)

    def extractfile(self, member):
        self._check("r")
        if isinstance(member, str):
            tarinfo = self.getmember(member)
        else:
            tarinfo = member
        if tarinfo.isreg() or tarinfo.type not in SUPPORTED_TYPES:
            start = tarinfo.offset_data
            return io.BytesIO(self._data[start:start + tarinfo.size])
        elif tarinfo.islnk() or tarinfo.issym():
            if tarinfo.islnk():
                linked = self._getmember(tarinfo.linkname)
                if linked is None:
                    raise KeyError("linkname %r not found" % tarinfo.linkname)
                return self.extractfile(linked)
            return None
        return None

    def _get_filter(self, filter):
        if filter is None:
            filter = self.extraction_filter
            if filter is None:
                return fully_trusted_filter
        if isinstance(filter, str):
            if filter == "fully_trusted":
                return fully_trusted_filter
            if filter == "tar":
                return tar_filter
            if filter == "data":
                return data_filter
            raise ValueError("filter %r not found" % filter)
        return filter

    def extractall(self, path=".", members=None, *, numeric_owner=False, filter=None):
        directories = []
        filter_function = self._get_filter(filter)
        if members is None:
            members = self
        for member in members:
            tarinfo = self._get_extract_tarinfo(member, filter_function, path)
            if tarinfo is None:
                continue
            if tarinfo.isdir():
                directories.append(tarinfo)
            self._extract_one(tarinfo, path)
        directories.sort(key=lambda a: a.name, reverse=True)
        for tarinfo in directories:
            dirpath = os.path.join(path, tarinfo.name)
            try:
                os.utime(dirpath, (tarinfo.mtime, tarinfo.mtime))
                os.chmod(dirpath, tarinfo.mode)
            except OSError:
                pass

    def extract(self, member, path="", set_attrs=True, *, numeric_owner=False, filter=None):
        filter_function = self._get_filter(filter)
        tarinfo = self._get_extract_tarinfo(member, filter_function, path)
        if tarinfo is not None:
            self._extract_one(tarinfo, path, set_attrs)

    def _get_extract_tarinfo(self, member, filter_function, path):
        if isinstance(member, str):
            tarinfo = self.getmember(member)
        else:
            tarinfo = member
        unfiltered = tarinfo
        try:
            tarinfo = filter_function(tarinfo, path)
        except FilterError:
            if self.errorlevel > 0:
                raise
            return None
        if tarinfo is None:
            return None
        return tarinfo

    def _extract_one(self, tarinfo, path, set_attrs=True):
        targetpath = os.path.join(path, tarinfo.name) if path else tarinfo.name
        targetpath = targetpath.rstrip("/")
        targetpath = targetpath.replace("/", os.sep)
        upperdirs = os.path.dirname(targetpath)
        if upperdirs and not os.path.exists(upperdirs):
            os.makedirs(upperdirs, exist_ok=True)
        if tarinfo.isreg():
            data = self._data[tarinfo.offset_data:tarinfo.offset_data + tarinfo.size]
            with io.open(targetpath, "wb") as target:
                target.write(data)
        elif tarinfo.isdir():
            try:
                os.mkdir(targetpath, 0o700)
            except FileExistsError:
                if not os.path.isdir(targetpath):
                    raise
        elif tarinfo.issym():
            os.symlink(tarinfo.linkname, targetpath)
        elif tarinfo.islnk():
            os.link(os.path.join(path, tarinfo.linkname) if path else tarinfo.linkname, targetpath)
        elif tarinfo.isfifo():
            os.mkfifo(targetpath)
        else:
            raise ExtractError("tarfile: Unsupported type %r" % tarinfo.type)
        if set_attrs and not tarinfo.issym():
            try:
                os.chmod(targetpath, tarinfo.mode)
                os.utime(targetpath, (tarinfo.mtime, tarinfo.mtime))
            except OSError:
                if self.errorlevel > 1:
                    raise

    def close(self):
        if self.closed:
            return
        self.closed = True
        try:
            if self.mode in ("a", "w", "x"):
                self._write(NUL * (BLOCKSIZE * 2))
                blocks, remainder = divmod(self.offset, RECORDSIZE)
                if remainder > 0:
                    self._write(NUL * (RECORDSIZE - remainder))
                out = b"".join(self._outbuf)
                if self._compression == "gz":
                    out = _gzip.compress(out, self._gz_level, mtime=int(time.time()))
                if self._compression == "gz" and self._gz_path != "":
                    with io.open(self._gz_path, "wb") as gzf:
                        gzf.write(out)
                elif self.mode == "a":
                    self.fileobj.seek(0)
                    self.fileobj.write(self._data[:self._append_base] + out)
                    self.fileobj.truncate()
                else:
                    self.fileobj.write(out)
        finally:
            if self._own_file:
                self.fileobj.close()

    def __enter__(self):
        self._check()
        return self

    def __exit__(self, type, value, traceback):
        if type is None:
            self.close()
        else:
            if not self._loaded or True:
                self.closed = True
                if self._own_file:
                    self.fileobj.close()


def _blocks(size):
    blocks, remainder = divmod(size, BLOCKSIZE)
    if remainder:
        blocks += 1
    return blocks * BLOCKSIZE


def _parse_pax(payload, encoding, errors):
    result = {}
    pos = 0
    binary = False
    while pos < len(payload):
        sp = payload.find(b" ", pos)
        if sp < 0:
            break
        length = int(payload[pos:sp])
        if length == 0:
            break
        record = payload[sp + 1:pos + length - 1]
        eq = record.find(b"=")
        keyword = record[:eq].decode("utf-8", "surrogateescape")
        value = record[eq + 1:]
        if keyword == "hdrcharset":
            binary = value == b"BINARY"
        else:
            if binary:
                value = value.decode(encoding, errors)
            else:
                value = value.decode("utf-8", "surrogateescape")
            result[keyword] = value
        pos += length
    return result


def _apply_pax(tarinfo, pax):
    for keyword, value in pax.items():
        if keyword == "path":
            tarinfo.name = value
        elif keyword == "linkpath":
            tarinfo.linkname = value
        elif keyword in ("uid", "gid", "size"):
            setattr(tarinfo, keyword, int(value))
        elif keyword == "mtime":
            tarinfo.mtime = float(value)
        elif keyword == "uname":
            tarinfo.uname = value
        elif keyword == "gname":
            tarinfo.gname = value


def is_tarfile(name):
    try:
        if hasattr(name, "read"):
            pos = name.tell()
            t = open(fileobj=name)
            name.seek(pos)
        else:
            t = open(name)
        t.close()
        return True
    except TarError:
        return False


def open(name=None, mode="r", fileobj=None, bufsize=RECORDSIZE, **kwargs):
    return TarFile.open(name, mode, fileobj, bufsize, **kwargs)
