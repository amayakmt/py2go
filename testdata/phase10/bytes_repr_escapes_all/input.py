b = bytes(range(0, 256, 5))
print(b)
print(repr(b), len(repr(b)), str(b), bytearray(b[:12]), repr(bytearray(b"\x00'\"")))
for x in (b"", b"'", b'"', b"'\"", b"\\", b"\t\n\r", b"\x7f\x80\xff", b"a\x00b", b"caf\xc3\xa9", bytes([39, 34]), bytes([92, 39])):
    print(repr(x), x, list(x)[:6], len(x), x.hex())
print([b"a", b"b"], (b"x",), {b"k": b"v"}, {b"a"}, [bytearray(b"z")], f"{b'ab'}", f"{b'ab'!r}", "%r %s" % (b"q", b"q"), b"%r" % b"q", str(b"a") == "b'a'", b"abc".decode("ascii") == "abc")
