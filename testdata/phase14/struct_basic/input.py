import struct

print(struct.pack("<hhl", 1, -2, 3), struct.pack(">I", 0xDEADBEEF), struct.pack("!H", 513))
print(struct.pack("@bhiq", 1, 2, 3, 4), struct.calcsize("@bhiq"), struct.calcsize("<bhiq"), struct.calcsize("@ci"), struct.calcsize("=ci"))
print(struct.unpack("<hhl", b"\x01\x00\xfe\xff\x03\x00\x00\x00"), struct.unpack(">I", b"\xde\xad\xbe\xef"))
print(struct.pack("<f", 1.5), struct.pack(">d", -0.1), struct.pack("<e", 3.14159), struct.unpack("<e", b"\x48\x42"), struct.unpack("<f", struct.pack("<f", 0.1)))
print(struct.pack("5s", b"abc"), struct.pack("2s", b"abcdef"), struct.pack("p", b"xy"), struct.pack("5p", b"hello world"), struct.unpack("5p", b"\x03abc\x00"))
print(struct.pack("?c", True, b"z"), struct.unpack("?c", b"\x00y"), struct.pack("3x"), struct.pack("<3B", 1, 2, 255))
print(struct.pack("<q", -(2 ** 63)), struct.pack("<Q", 2 ** 64 - 1), struct.unpack("<Q", b"\xff" * 8), struct.unpack("<q", b"\xff" * 8))
print(struct.pack("<2h", *(1, 2)), struct.unpack("<2h2s", b"\x01\x00\x02\x00ab"), struct.unpack("<3b", b"\xff\x00\x7f"))
print(struct.unpack_from("<h", b"\x00\x01\x02\x03", 2), struct.unpack_from("<h", b"\x00\x01\x02\x03", offset=1))
buf = bytearray(8)
struct.pack_into("<hh", buf, 2, 7, -7)
print(bytes(buf), list(struct.iter_unpack("<h", b"\x01\x00\x02\x00\x03\x00")))
s = struct.Struct("<ii")
print(s.size, s.format, s.pack(1, 2), s.unpack(b"\x01\x00\x00\x00\x02\x00\x00\x00"), repr(s))
for fmt, args in [("<b", (200,)), ("<B", (-1,)), ("<h", (70000,)), ("<H", (-1,)), ("<i", (2 ** 31,)), ("<I", (-1,)), ("<q", (2 ** 63,)), ("<Q", (-1,)), ("<l", (2 ** 40,)), ("@l", (2 ** 63,)),
                  ("<h", (1.5,)), ("<h", ("a",)), ("<f", ("a",)), ("<c", (b"ab",)), ("<s", (1,)), ("<hh", (1,)), ("<h", (1, 2)), ("<z", (1,)), ("3", ()), ("<e", (1e10,))]:
    try:
        print(struct.pack(fmt, *args))
    except (struct.error, OverflowError) as e:
        print(type(e).__name__, e)
for fmt, data in [("<i", b"\x00"), ("<i", b"\x00" * 5)]:
    try:
        struct.unpack(fmt, data)
    except struct.error as e:
        print(e)
try:
    struct.unpack_from("<i", b"\x00\x00\x00\x00", 2)
except struct.error as e:
    print(e)
print(struct.pack("<I", True), struct.pack("<f", 3), struct.pack("<d", 2 ** 70), struct.unpack("<f", b"\x00\x00\xc0\x7f")[0] != struct.unpack("<f", b"\x00\x00\xc0\x7f")[0])
print(struct.unpack("<d", struct.pack("<d", float("inf"))), struct.pack(">e", 65504.0), struct.pack("<e", 5.960464477539063e-08), struct.pack("<e", float("-inf")))
print(struct.error.__name__, issubclass(struct.error, Exception))
