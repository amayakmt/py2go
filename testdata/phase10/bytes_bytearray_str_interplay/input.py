print(str(b"abc", "ascii"), str(bytearray(b"xyz"), "utf-8"), str(b"\xc3\xa9", "utf-8"), bytes("héllo", "latin-1"), bytearray("héllo", "utf-8"), "abc".encode() == b"abc", b"abc".decode() == "abc", "abc" == b"abc", b"abc" == "abc", {b"a": 1, "a": 2}, len({b"a", "a"}))
print(bytes(b"abc") is not None, isinstance(b"a", bytes), isinstance(bytearray(), bytes), isinstance(b"a", bytearray), isinstance(b"a", (str, bytes)), type(b"a").__name__, type(bytearray(b"a")).__name__, bytes.__name__, bytearray.__name__)
s = "naïve café"
b = s.encode("utf-8")
print(len(s), len(b), b.count(b"\xc3"), b.decode("utf-8")[3], b[2:4], b[2:4].decode(), [hex(c) for c in b[:5]], b.upper(), s.upper().encode() == b.upper() or b.upper())
words = b"the quick brown fox".split()
print(words, [w.capitalize() for w in words], b" ".join(w[::-1] for w in words), b"-".join(sorted(words, key=len)), [len(w) for w in words], max(words), min(words, key=lambda w: w[-1]))
for f in (lambda: b"a" < "a", lambda: str(b"a", 5), lambda: "a".encode(5), lambda: b"a".decode(5), lambda: bytes(1.5), lambda: bytes({1, 2}) if False else bytes(b"ok"), lambda: bytearray("abc")):
    try:
        print(f())
    except TypeError as e:
        print(type(e).__name__, e)
