b = b"hello world"
print(b, len(b), b[0], b[-1], b[1:4], b[::-1], b[::2], type(b[0]).__name__, type(b[0:1]).__name__, list(b[:3]), tuple(b"ab"), bytes(3), bytes([65, 66, 67]), bytes(b"xyz"), bytes("é", "utf-8"), bytes(range(5)))
print(b == b"hello world", b != b"x", b < b"i", b"a" < b"b", b"abc" < b"abd", b"" < b"a", b"ab" > b"a", max(b"abc"), min(b"abc"), sorted(b"cab"), sum(b"abc"), 104 in b, b"wor" in b, b"xyz" not in b)
print(repr(b"\x00\x01\xff'\"\\\n\t\r"), repr(b"it's"), repr(b'say "hi"'), repr(b"""'"'"""), b"\x7f\x80", str(b"abc"), f"{b!r}", f"{b!s}", [b"a", b"b"], {b"k": b"v"}, (b"x",))
print(b"abc" + b"def", b"ab" * 3, 3 * b"ab", b"ab" * 0, b"ab" * -1, b + bytearray(b"!"), bytearray(b"!") + b)
for f in (lambda: b[100], lambda: b"a" + "b", lambda: "a" + b"b", lambda: b"a" * 1.5, lambda: bytes([256]), lambda: bytes([-1]), lambda: bytes(-1), lambda: bytes("abc"), lambda: bytes(["a"]), lambda: b"abc"[1] + b"x"):
    try:
        print(f())
    except (IndexError, TypeError, ValueError) as e:
        print(type(e).__name__, e)
x = b"abc"
x += b"def"
print(x, hash(b"abc") == hash(b"abc"), {b"a": 1}[b"a"], bool(b""), bool(b"\x00"), b"ab".join([b"1", b"2", b"3"]), b"".join([]), b",".join(map(bytes, [[65], [66]])))
