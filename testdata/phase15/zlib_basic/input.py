import zlib

data = b"hello hello hello hello world " * 50 + bytes(range(256))
for level in (-1, 0, 1, 6, 9):
    c = zlib.compress(data, level)
    print(level, zlib.decompress(c) == data, len(c) < len(data) or level == 0, c[:1], zlib.crc32(c) >= 0)
print(zlib.compress(b"", 6)[:2], zlib.decompress(zlib.compress(b"")), zlib.decompress(zlib.compress(b"x" * 100000)) == b"x" * 100000)
raw = zlib.compress(data, 6, -15)
gz = zlib.compress(data, 6, 31)
print(zlib.decompress(raw, -15) == data, zlib.decompress(gz, 31) == data, zlib.decompress(gz, 47) == data, zlib.decompress(zlib.compress(data), 47) == data, gz[:2])
print(zlib.crc32(b"hello"), zlib.crc32(b"world", zlib.crc32(b"hello ")), zlib.crc32(b""), zlib.adler32(b"hello"), zlib.adler32(b"world", zlib.adler32(b"hello ")), zlib.adler32(b""), zlib.crc32(b"abc") == zlib.crc32(b"c", zlib.crc32(b"ab")))
co = zlib.compressobj()
parts = [co.compress(data[:300]), co.compress(data[300:]), co.flush()]
print(zlib.decompress(b"".join(parts)) == data)
do = zlib.decompressobj()
whole = zlib.compress(data)
out = b""
for i in range(0, len(whole), 40):
    out += do.decompress(whole[i:i + 40])
print(out == data, do.eof, do.unused_data, do.flush())
do2 = zlib.decompressobj()
print(do2.decompress(whole + b"trailing") == data, do2.unused_data, do2.eof)
for call in (lambda: zlib.decompress(b"not zlib data"), lambda: zlib.decompress(whole[:20]), lambda: zlib.decompress(b""), lambda: zlib.compress(b"x", 12), lambda: zlib.decompress(whole[:-1] + b"\x00"), lambda: zlib.compress("str")):
    try:
        print(call())
    except (zlib.error, TypeError) as e:
        print(type(e).__name__, e)
print(zlib.MAX_WBITS, zlib.Z_BEST_COMPRESSION, zlib.Z_DEFAULT_COMPRESSION, zlib.DEFLATED, zlib.error.__name__, issubclass(zlib.error, Exception))
