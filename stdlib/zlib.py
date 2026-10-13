"""zlib: compression compatible with gzip (Go's deflate underneath)."""

from builtins import zlib_pm_error as error

MAX_WBITS = 15
DEFLATED = 8
DEF_MEM_LEVEL = 8
DEF_BUF_SIZE = 16384
Z_NO_COMPRESSION = 0
Z_BEST_SPEED = 1
Z_BEST_COMPRESSION = 9
Z_DEFAULT_COMPRESSION = -1
Z_FILTERED = 1
Z_HUFFMAN_ONLY = 2
Z_RLE = 3
Z_FIXED = 4
Z_DEFAULT_STRATEGY = 0
Z_NO_FLUSH = 0
Z_PARTIAL_FLUSH = 1
Z_SYNC_FLUSH = 2
Z_FULL_FLUSH = 3
Z_FINISH = 4
Z_BLOCK = 5
Z_TREES = 6
ZLIB_VERSION = "1.3.1"
ZLIB_RUNTIME_VERSION = "1.3.1"


def compress(data, /, level=-1, wbits=15):
    return _native("zlib.compress", data, level, wbits)


def decompress(data, /, wbits=15, bufsize=16384):
    return _native("zlib.decompress", data, wbits)


def crc32(data, value=0, /):
    return _native("zlib.crc32", data, value)


def adler32(data, value=1, /):
    return _native("zlib.adler32", data, value)


def compressobj(level=-1, method=8, wbits=15, memLevel=8, strategy=0, zdict=None):
    return _native("zlib.compressobj", level, method, wbits)


def decompressobj(wbits=15, zdict=b""):
    return _native("zlib.decompressobj", wbits)
