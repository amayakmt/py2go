b = b"hello \x00\xff world"
print(b, len(b), b[0], b[-1], b[1:4], b[::-1], b[::2], type(b).__name__, repr(b"it's"), repr(b'say "hi"'), repr(b"both ' and \""), b"\n\t\r\\")
print(bytes(), bytes(3), bytes([104, 105]), bytes("héllo", "utf-8"), bytes(b"abc"), bytes(bytearray(b"xyz")), bytes(range(5)), list(b"abc"), tuple(b"\x01\x02"))
print(b"ab" + b"cd", b"ab" * 3, 3 * b"x", b"abc" == b"abc", b"abc" < b"abd", b"b" > b"abc", b"" == b"", b"a" != b"b", b"abc" in b"xabcx", 97 in b"abc", 0 in b"abc")
print(b"abc".hex(), b"\x00\xff".hex(), b"abcdef".hex("-"), b"abcdef".hex(":", 2), bytes.fromhex("48 65 6c6c6f"), bytes.fromhex("00FF"), bytes.fromhex(""))
print(bool(b""), bool(b"\x00"), hash(b"abc") == hash(b"abc"), {b"k": 1}[b"k"], {b"a", b"a", b"b"} == {b"a", b"b"}, sorted([b"b", b"a", b"ab"]), max(b"abc"), min(b"abc"), sum(b"abc"))
print(b"%s-%d-%x" % (b"ab", 5, 255), b"%b" % b"x", b"%5s|%-5s|" % (b"a", b"b"), b"%c" % 65 if False else b"A")
for f in (lambda: b"a" + "b", lambda: "a" + b"b", lambda: b"a" * 1.5, lambda: b"abc"[10], lambda: bytes(-1), lambda: bytes([256]), lambda: bytes("x"), lambda: b"a" < "a", lambda: bytes([1.5]), lambda: 300 in b"a", lambda: bytes.fromhex("xyz")):
    try:
        print(f())
    except (TypeError, IndexError, ValueError) as e:
        print(type(e).__name__, e)
print(b"a" == "a", "a" == b"a", b"a" in {b"a": 1}, str(b"abc"), repr([b"a", b"b"]), f"{b'x'}", "%s" % b"y", "%r" % b"z")
