"""zipfile: read and write ZIP files (deflate and store; no ZIP64, bzip2 or lzma)."""
import binascii
import io
import os
import shutil
import stat
import struct
import time
import zlib

__all__ = ["BadZipFile", "BadZipfile", "error", "ZIP_STORED", "ZIP_DEFLATED", "ZIP_BZIP2", "ZIP_LZMA",
           "is_zipfile", "ZipInfo", "ZipFile", "LargeZipFile"]


def _obj(x: object) -> object:
    return x


class BadZipFile(Exception):
    pass


class LargeZipFile(Exception):
    pass


error = BadZipfile = BadZipFile

ZIP64_LIMIT = (1 << 31) - 1
ZIP_FILECOUNT_LIMIT = (1 << 16) - 1
ZIP_MAX_COMMENT = (1 << 16) - 1

ZIP_STORED = 0
ZIP_DEFLATED = 8
ZIP_BZIP2 = 12
ZIP_LZMA = 14

DEFAULT_VERSION = 20
ZIP64_VERSION = 45

structEndArchive = b"<4s4H2LH"
stringEndArchive = b"PK\005\006"
sizeEndCentDir = struct.calcsize(structEndArchive)

structCentralDir = "<4s4B4HL2L5H2L"
stringCentralDir = b"PK\001\002"
sizeCentralDir = struct.calcsize(structCentralDir)

structFileHeader = "<4s2B4HL2L2H"
stringFileHeader = b"PK\003\004"
sizeFileHeader = struct.calcsize(structFileHeader)

_DD_SIGNATURE = 0x08074b50


def _check_zipfile_end(data):
    pos = data.rfind(stringEndArchive)
    return pos


def is_zipfile(filename):
    try:
        if hasattr(filename, "read"):
            pos = filename.tell()
            data = filename.read()
            filename.seek(pos)
        else:
            with open(filename, "rb") as fp:
                data = fp.read()
        return _find_eocd(data) >= 0
    except OSError:
        return False


def _find_eocd(data):
    start = len(data) - sizeEndCentDir
    stop = max(0, start - ZIP_MAX_COMMENT - 1)
    pos = start
    while pos >= stop:
        if data[pos:pos + 4] == stringEndArchive:
            if pos + sizeEndCentDir <= len(data):
                comment_len = struct.unpack("<H", data[pos + 20:pos + 22])[0]
                if pos + sizeEndCentDir + comment_len <= len(data):
                    return pos
        pos -= 1
    return -1


class ZipInfo:
    def __init__(self, filename="NoName", date_time=(1980, 1, 1, 0, 0, 0)):
        self.orig_filename = filename
        null_byte = filename.find(chr(0))
        if null_byte >= 0:
            filename = filename[0:null_byte]
        if os.sep != "/" and os.sep in filename:
            filename = filename.replace(os.sep, "/")
        self.filename = filename
        self.date_time = date_time
        if date_time[0] < 1980:
            raise ValueError("ZIP does not support timestamps before 1980")
        self.compress_type = ZIP_STORED
        self.comment = b""
        self.extra = b""
        self.create_system = 3
        self.create_version = DEFAULT_VERSION
        self.extract_version = DEFAULT_VERSION
        self.reserved = 0
        self.flag_bits = 0
        self.volume = 0
        self.internal_attr = 0
        self.external_attr = 0
        self.compress_level = None
        self.header_offset = 0
        self.CRC = 0
        self.compress_size = 0
        self.file_size = 0

    def __repr__(self):
        result = ["<%s filename=%r" % (self.__class__.__name__, self.filename)]
        if self.compress_type != ZIP_STORED:
            result.append(" compress_type=%s" % _compressor_name(self.compress_type))
        hi = self.external_attr >> 16
        lo = self.external_attr & 0xFFFF
        if hi:
            result.append(" filemode=%r" % stat.filemode(hi))
        if lo:
            result.append(" external_attr=%#x" % lo)
        isdir = self.is_dir()
        if not isdir or self.file_size:
            result.append(" file_size=%r" % self.file_size)
        if (not isdir or self.compress_size) and (self.compress_type != ZIP_STORED or self.file_size != self.compress_size):
            result.append(" compress_size=%r" % self.compress_size)
        result.append(">")
        return "".join(result)

    def FileHeader(self, zip64=None):
        dt = self.date_time
        dosdate = (dt[0] - 1980) << 9 | dt[1] << 5 | dt[2]
        dostime = dt[3] << 11 | dt[4] << 5 | (dt[5] // 2)
        if self.flag_bits & 0x08:
            CRC = compress_size = file_size = 0
        else:
            CRC = self.CRC
            compress_size = self.compress_size
            file_size = self.file_size
        extra = self.extra
        min_version = 0
        self.extract_version = max(min_version, self.extract_version)
        self.create_version = max(min_version, self.create_version)
        filename, flag_bits = self._encodeFilenameFlags()
        header = struct.pack(structFileHeader, stringFileHeader, self.extract_version, self.reserved, flag_bits,
                             self.compress_type, dostime, dosdate, CRC, compress_size, file_size, len(filename), len(extra))
        return header + filename + extra

    def _encodeFilenameFlags(self):
        try:
            return self.filename.encode("ascii"), self.flag_bits
        except UnicodeEncodeError:
            return self.filename.encode("utf-8"), self.flag_bits | 0x800

    def _decodeExtra(self):
        pass

    def is_dir(self):
        return self.filename[-1:] == "/"

    @classmethod
    def from_file(cls, filename, arcname=None, *, strict_timestamps=True):
        filename = os.fspath(filename)
        st = os.stat(filename)
        isdir = stat.S_ISDIR(st.st_mode)
        mtime = time.localtime(st.st_mtime)
        date_time = (mtime[0], mtime[1], mtime[2], mtime[3], mtime[4], mtime[5])
        if not strict_timestamps and date_time[0] < 1980:
            date_time = (1980, 1, 1, 0, 0, 0)
        elif not strict_timestamps and date_time[0] > 2107:
            date_time = (2107, 12, 31, 23, 59, 59)
        if arcname is None:
            arcname = filename
        arcname = os.path.normpath(os.path.splitdrive(arcname)[1])
        while arcname[0] in (os.sep,):
            arcname = arcname[1:]
        if isdir:
            arcname += "/"
        zinfo = cls(arcname, date_time)
        zinfo.external_attr = (st.st_mode & 0xFFFF) << 16
        if isdir:
            zinfo.file_size = 0
            zinfo.external_attr |= 0x10
        else:
            zinfo.file_size = st.st_size
        return zinfo


def _compressor_name(t):
    names = {0: "store", 8: "deflate", 12: "bzip2", 14: "lzma"}
    return names.get(t, str(t))


class _ZipFileReader:
    pass


class ZipFile:
    def __init__(self, file, mode="r", compression=ZIP_STORED, allowZip64=True, compresslevel=None, *,
                 strict_timestamps=True, metadata_encoding=None):
        if mode not in ("r", "w", "x", "a"):
            raise ValueError("ZipFile requires mode 'r', 'w', 'x', or 'a'")
        if compression not in (ZIP_STORED, ZIP_DEFLATED):
            if compression in (ZIP_BZIP2, ZIP_LZMA):
                raise RuntimeError("Compression requires the (missing) %s module" % ("bz2" if compression == ZIP_BZIP2 else "lzma"))
            raise NotImplementedError("That compression method is not supported")
        self._allowZip64 = allowZip64
        self._didModify = False
        self.debug = 0
        self.NameToInfo = {}
        self.filelist = []
        self.compression = compression
        self.compresslevel = compresslevel
        self.mode = mode
        self.pwd = None
        self._comment = b""
        self._strict_timestamps = strict_timestamps
        self.metadata_encoding = metadata_encoding
        self._filePassed = 0
        self._closed = False
        self._start_disk = 0
        self.start_dir = 0
        self._writing = False
        self.fp = _obj(None)
        if isinstance(file, os.PathLike):
            file = os.fspath(file)
        if isinstance(file, str):
            self._filePassed = 0
            self.filename = file
            modeDict = {"r": "rb", "w": "w+b", "x": "x+b", "a": "r+b", "r+b": "w+b", "w+b": "wb", "x+b": "xb"}
            filemode = modeDict[mode]
            while True:
                try:
                    self.fp = _obj(io.open(file, filemode))
                except OSError:
                    if filemode in modeDict:
                        filemode = modeDict[filemode]
                        continue
                    raise
                break
        else:
            self._filePassed = 1
            self.fp = _obj(file)
            self.filename = getattr(file, "name", None)
        self._fileRefCnt = 1
        try:
            if mode == "r":
                self._RealGetContents()
            elif mode in ("w", "x"):
                self._didModify = True
                try:
                    self.start_dir = self.fp.tell()
                except (AttributeError, OSError):
                    self.start_dir = 0
            elif mode == "a":
                try:
                    self._RealGetContents()
                    self.fp.seek(self.start_dir)
                except BadZipFile:
                    self.fp.seek(0, 2)
                    self._didModify = True
                    self.start_dir = self.fp.tell()
            else:
                raise ValueError("Mode must be 'r', 'w', 'x', or 'a'")
        except:
            fp = self.fp
            self.fp = _obj(None)
            if not self._filePassed:
                fp.close()
            raise

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        self.close()

    def __repr__(self):
        result = ["<zipfile.%s" % self.__class__.__name__]
        if self.fp is not None:
            if self._filePassed:
                result.append(" file=%r" % self.fp)
            elif self.filename is not None:
                result.append(" filename=%r" % self.filename)
            result.append(" mode=%r" % self.mode)
        else:
            result.append(" [closed]")
        result.append(">")
        return "".join(result)

    def _RealGetContents(self):
        fp = self.fp
        fp.seek(0, 2)
        data_len = fp.tell()
        fp.seek(0)
        data = fp.read()
        pos = _find_eocd(data)
        if pos < 0:
            raise BadZipFile("File is not a zip file")
        endrec = struct.unpack(structEndArchive, data[pos:pos + sizeEndCentDir])
        size_cd = endrec[5]
        offset_cd = endrec[6]
        self._comment = data[pos + sizeEndCentDir:pos + sizeEndCentDir + endrec[7]]
        concat = pos - size_cd - offset_cd
        self.start_dir = offset_cd + concat
        if self.start_dir < 0:
            raise BadZipFile("Bad offset for central directory")
        total = endrec[4]
        p = self.start_dir
        end = p + size_cd
        for i in range(total):
            centdir = data[p:p + sizeCentralDir]
            if len(centdir) != sizeCentralDir:
                raise BadZipFile("Truncated central directory")
            c = struct.unpack(structCentralDir, centdir)
            if c[0] != stringCentralDir:
                raise BadZipFile("Bad magic number for central directory")
            p += sizeCentralDir
            filename = data[p:p + c[12]]
            p += c[12]
            flags = c[5]
            if flags & 0x800:
                filename = filename.decode("utf-8")
            else:
                filename = filename.decode(self.metadata_encoding or "cp437")
            x = ZipInfo(filename)
            x.extra = data[p:p + c[13]]
            p += c[13]
            x.comment = data[p:p + c[14]]
            p += c[14]
            x.create_version = c[1]
            x.create_system = c[2]
            x.extract_version = c[3]
            x.reserved = c[4]
            x.flag_bits = c[5]
            x.compress_type = c[6]
            t = c[7]
            d = c[8]
            x.CRC = c[9]
            x.compress_size = c[10]
            x.file_size = c[11]
            x.volume = c[15]
            x.internal_attr = c[16]
            x.external_attr = c[17]
            x.header_offset = c[18]
            x.date_time = ((d >> 9) + 1980, (d >> 5) & 0xF, d & 0x1F, t >> 11, (t >> 5) & 0x3F, (t & 0x1F) * 2)
            x._decodeExtra()
            x.header_offset = x.header_offset + concat
            self.filelist.append(x)
            self.NameToInfo[x.filename] = x

    def namelist(self):
        return [data.filename for data in self.filelist]

    def infolist(self):
        return self.filelist

    def printdir(self, file=None):
        import sys
        if file is None:
            file = sys.stdout
        print("%-46s %19s %12s" % ("File Name", "Modified    ", "Size"), file=file)
        for zinfo in self.filelist:
            date = "%d-%02d-%02d %02d:%02d:%02d" % zinfo.date_time[:6]
            print("%-46s %s %12d" % (zinfo.filename, date, zinfo.file_size), file=file)

    def testzip(self):
        chunk_size = 1 << 20
        for zinfo in self.filelist:
            try:
                with self.open(zinfo.filename, "r") as f:
                    while f.read(chunk_size):
                        pass
            except BadZipFile:
                return zinfo.filename
        return None

    def getinfo(self, name):
        info = self.NameToInfo.get(name)
        if info is None:
            raise KeyError("There is no item named %r in the archive" % name)
        return info

    def setpassword(self, pwd):
        if pwd and not isinstance(pwd, bytes):
            raise TypeError("pwd: expected bytes, got %s" % type(pwd).__name__)
        self.pwd = pwd or None

    @property
    def comment(self):
        return self._comment

    @comment.setter
    def comment(self, comment):
        if not isinstance(comment, bytes):
            raise TypeError("comment: expected bytes, got %s" % type(comment).__name__)
        if len(comment) > ZIP_MAX_COMMENT:
            comment = comment[:ZIP_MAX_COMMENT]
        self._comment = comment
        self._didModify = True

    def read(self, name, pwd=None):
        with self.open(name, "r", pwd) as fp:
            return fp.read()

    def _read_member(self, zinfo):
        fp = self.fp
        fp.seek(zinfo.header_offset)
        fheader = fp.read(sizeFileHeader)
        if len(fheader) != sizeFileHeader:
            raise BadZipFile("Truncated file header")
        fheader = struct.unpack(structFileHeader, fheader)
        if fheader[0] != stringFileHeader:
            raise BadZipFile("Bad magic number for file header")
        fname = fp.read(fheader[10])
        if fheader[11]:
            fp.read(fheader[11])
        if zinfo.flag_bits & 0x800:
            fname_str = fname.decode("utf-8")
        else:
            fname_str = fname.decode("cp437")
        if fname_str != zinfo.orig_filename:
            raise BadZipFile("File name in directory %r and header %r differ." % (zinfo.orig_filename, fname))
        if zinfo.flag_bits & 0x1:
            raise NotImplementedError("File is encrypted; encryption is not supported")
        raw = fp.read(zinfo.compress_size)
        if zinfo.compress_type == ZIP_STORED:
            data = raw
        elif zinfo.compress_type == ZIP_DEFLATED:
            try:
                data = zlib.decompress(raw, -15)
            except zlib.error as e:
                raise BadZipFile("Error -3 while decompressing data: invalid stored block lengths") from e
        else:
            raise NotImplementedError("That compression method is not supported")
        if (binascii.crc32(data) & 0xFFFFFFFF) != zinfo.CRC:
            raise BadZipFile("Bad CRC-32 for file %r" % zinfo.orig_filename)
        return data

    def open(self, name, mode="r", pwd=None, *, force_zip64=False):
        if mode not in ("r", "w"):
            raise ValueError('open() requires mode "r" or "w"')
        if pwd and mode == "w":
            raise ValueError("pwd is only supported for reading files")
        if not self.fp:
            raise ValueError("Attempt to use ZIP archive that was already closed")
        if mode == "w":
            if isinstance(name, ZipInfo):
                zinfo = name
            else:
                zinfo = ZipInfo(name)
                dt = time.localtime(time.time())
                zinfo.date_time = (dt[0], dt[1], dt[2], dt[3], dt[4], dt[5])
                zinfo.compress_type = self.compression
                zinfo.compress_level = self.compresslevel
            return _ZipWriteFile(self, zinfo)
        if isinstance(name, ZipInfo):
            zinfo = name
        elif mode == "w":
            zinfo = ZipInfo(name)
            zinfo.compress_type = self.compression
        else:
            zinfo = self.getinfo(name)
        if self._writing:
            raise ValueError("Can't read from the ZIP file while there is an open writing handle on it. Close the writing handle before trying to read.")
        data = self._read_member(zinfo)
        return _ZipReadFile(data, zinfo)

    def extract(self, member, path=None, pwd=None):
        if path is None:
            path = os.getcwd()
        else:
            path = os.fspath(path)
        return self._extract_member(member, path, pwd)

    def extractall(self, path=None, members=None, pwd=None):
        if members is None:
            members = self.namelist()
        if path is None:
            path = os.getcwd()
        else:
            path = os.fspath(path)
        for zipinfo in members:
            self._extract_member(zipinfo, path, pwd)

    def _extract_member(self, member, targetpath, pwd):
        if not isinstance(member, ZipInfo):
            member = self.getinfo(member)
        arcname = member.filename.replace("/", os.path.sep)
        if os.path.altsep:
            arcname = arcname.replace(os.path.altsep, os.path.sep)
        arcname = os.path.splitdrive(arcname)[1]
        invalid_path_parts = ("", os.path.curdir, os.path.pardir)
        arcname = os.path.sep.join([x for x in arcname.split(os.path.sep) if x not in invalid_path_parts])
        targetpath = os.path.join(targetpath, arcname)
        targetpath = os.path.normpath(targetpath)
        upperdirs = os.path.dirname(targetpath)
        if upperdirs and not os.path.exists(upperdirs):
            os.makedirs(upperdirs, exist_ok=True)
        if member.is_dir():
            if not os.path.isdir(targetpath):
                os.mkdir(targetpath)
            return targetpath
        data = self._read_member(member)
        with open(targetpath, "wb") as target:
            target.write(data)
        return targetpath

    def _writecheck(self, zinfo):
        if zinfo.filename in self.NameToInfo:
            import warnings
            warnings.warn("Duplicate name: %r" % zinfo.filename, UserWarning, stacklevel=3)
        if self.mode not in ("w", "x", "a"):
            raise ValueError("write() requires mode 'w', 'x', or 'a'")
        if not self.fp:
            raise ValueError("Attempt to write ZIP archive that was already closed")
        if zinfo.compress_type not in (ZIP_STORED, ZIP_DEFLATED):
            raise NotImplementedError("That compression method is not supported")

    def write(self, filename, arcname=None, compress_type=None, compresslevel=None):
        if not self.fp:
            raise ValueError("Attempt to write to ZIP archive that was already closed")
        if self._writing:
            raise ValueError("Can't write to ZIP archive while an open writing handle exists")
        zinfo = ZipInfo.from_file(filename, arcname, strict_timestamps=self._strict_timestamps)
        if zinfo.is_dir():
            zinfo.compress_size = 0
            zinfo.CRC = 0
            self.mkdir(zinfo)
        else:
            if compress_type is not None:
                zinfo.compress_type = compress_type
            else:
                zinfo.compress_type = self.compression
            self._writecheck(zinfo)
            with open(filename, "rb") as src:
                data = src.read()
            self._write_data(zinfo, data)

    def writestr(self, zinfo_or_arcname, data, compress_type=None, compresslevel=None):
        if isinstance(data, str):
            data = data.encode("utf-8")
        if not isinstance(zinfo_or_arcname, ZipInfo):
            zinfo = ZipInfo(filename=zinfo_or_arcname, date_time=_now())
            zinfo.compress_type = self.compression
            zinfo.compress_level = self.compresslevel
            if zinfo.filename[-1] == "/":
                zinfo.external_attr = 0o40775 << 16
                zinfo.external_attr |= 0x10
            else:
                zinfo.external_attr = 0o600 << 16
        else:
            zinfo = zinfo_or_arcname
        if not self.fp:
            raise ValueError("Attempt to write to ZIP archive that was already closed")
        if self._writing:
            raise ValueError("Can't write to ZIP archive while an open writing handle exists.")
        if compress_type is not None:
            zinfo.compress_type = compress_type
        self._writecheck(zinfo)
        self._write_data(zinfo, bytes(data))

    def _write_data(self, zinfo, data):
        self._didModify = True
        if not zinfo.external_attr:
            zinfo.external_attr = 0o600 << 16
        zinfo.file_size = len(data)
        zinfo.CRC = binascii.crc32(data) & 0xFFFFFFFF
        if zinfo.compress_type == ZIP_DEFLATED:
            level = zinfo.compress_level if zinfo.compress_level is not None else -1
            c = zlib.compressobj(level, zlib.DEFLATED, -15)
            comp = c.compress(data) + c.flush()
        else:
            comp = data
        zinfo.compress_size = len(comp)
        if zinfo.compress_size > ZIP64_LIMIT or zinfo.file_size > ZIP64_LIMIT:
            raise LargeZipFile("Filesize would require ZIP64 extensions")
        zinfo.header_offset = self.fp.tell()
        zinfo.flag_bits &= ~0x08
        self.fp.write(zinfo.FileHeader())
        self.fp.write(comp)
        self.filelist.append(zinfo)
        self.NameToInfo[zinfo.filename] = zinfo

    def mkdir(self, zinfo_or_directory_name, mode=511):
        if isinstance(zinfo_or_directory_name, ZipInfo):
            zinfo = zinfo_or_directory_name
            if not zinfo.is_dir():
                raise ValueError("The given ZipInfo does not describe a directory")
        elif isinstance(zinfo_or_directory_name, str):
            directory_name = zinfo_or_directory_name
            if not directory_name.endswith("/"):
                directory_name += "/"
            zinfo = ZipInfo(directory_name)
            zinfo.compress_size = 0
            zinfo.CRC = 0
            zinfo.external_attr = ((0o40000 | mode) & 0xFFFF) << 16
            zinfo.file_size = 0
            zinfo.external_attr |= 0x10
        else:
            raise TypeError("Expected type str or ZipInfo")
        self._write_data(zinfo, b"")

    def close(self):
        if self.fp is None:
            return
        if self._writing:
            raise ValueError("Can't close the ZIP file while there is an open writing handle on it. Close the writing handle before closing the zip.")
        try:
            if self.mode in ("w", "x", "a") and self._didModify:
                self._write_end_record()
        finally:
            fp = self.fp
            self.fp = _obj(None)
            if not self._filePassed:
                fp.close()

    def _write_end_record(self):
        pos1 = self.fp.tell()
        for zinfo in self.filelist:
            dt = zinfo.date_time
            dosdate = (dt[0] - 1980) << 9 | dt[1] << 5 | dt[2]
            dostime = dt[3] << 11 | dt[4] << 5 | (dt[5] // 2)
            extra_data = zinfo.extra
            filename, flag_bits = zinfo._encodeFilenameFlags()
            centdir = struct.pack(structCentralDir, stringCentralDir, zinfo.create_version, zinfo.create_system,
                                  zinfo.extract_version, zinfo.reserved, flag_bits, zinfo.compress_type, dostime,
                                  dosdate, zinfo.CRC, zinfo.compress_size, zinfo.file_size, len(filename),
                                  len(extra_data), len(zinfo.comment), 0, zinfo.internal_attr, zinfo.external_attr,
                                  zinfo.header_offset)
            self.fp.write(centdir)
            self.fp.write(filename)
            self.fp.write(extra_data)
            self.fp.write(zinfo.comment)
        pos2 = self.fp.tell()
        centDirCount = len(self.filelist)
        centDirSize = pos2 - pos1
        centDirOffset = pos1
        if centDirCount > ZIP_FILECOUNT_LIMIT or centDirOffset > ZIP64_LIMIT or centDirSize > ZIP64_LIMIT:
            raise LargeZipFile("Zip64 extensions are not supported")
        endrec = struct.pack(structEndArchive, stringEndArchive, 0, 0, centDirCount, centDirCount, centDirSize,
                             centDirOffset, len(self._comment))
        self.fp.write(endrec)
        self.fp.write(self._comment)
        if self.mode == "a":
            self.fp.truncate()
        self.fp.flush()


def _decode_extra_stub(self):
    pass


def _now():
    t = time.localtime(time.time())
    return (t[0], t[1], t[2], t[3], t[4], t[5])


class _ZipReadFile:
    def __init__(self, data, zinfo):
        self._buf = io.BytesIO(data)
        self._data = data
        self.name = zinfo.filename
        self._zinfo = zinfo
        self.mode = "r"

    def read(self, n=-1):
        return self._buf.read(n)

    def read1(self, n=-1):
        return self._buf.read(n)

    def readline(self, limit=-1):
        return self._buf.readline(limit)

    def readlines(self, hint=-1):
        return self._buf.readlines(hint)

    def peek(self, n=1):
        pos = self._buf.tell()
        data = self._buf.read(n)
        self._buf.seek(pos)
        return data

    def seek(self, offset, whence=0):
        return self._buf.seek(offset, whence)

    def tell(self):
        return self._buf.tell()

    def seekable(self):
        return True

    def readable(self):
        return True

    def writable(self):
        return False

    def close(self):
        self._buf.close()

    @property
    def closed(self):
        return self._buf.closed

    def __iter__(self):
        return iter(self._buf.readlines())

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        self.close()


class _ZipWriteFile:
    def __init__(self, zf, zinfo):
        self._zf = zf
        self._zinfo = zinfo
        self._parts = []
        self._closed = False
        self.name = zinfo.filename
        zf._writing = True

    def writable(self):
        return True

    def write(self, data):
        if self._closed:
            raise ValueError("I/O operation on closed file.")
        data = bytes(data)
        self._parts.append(data)
        return len(data)

    def close(self):
        if self._closed:
            return
        self._closed = True
        self._zf._writing = False
        self._zf._writecheck(self._zinfo)
        self._zf._write_data(self._zinfo, b"".join(self._parts))

    @property
    def closed(self):
        return self._closed

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        self.close()
