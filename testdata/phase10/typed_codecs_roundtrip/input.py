def roundtrip(s: str, enc: str) -> bool:
    try:
        return s.encode(enc).decode(enc) == s
    except UnicodeError:
        return False

samples: list[str] = ["plain", "ünï", "日本語", "emoji 😀", "", "\x00\x01", "ÿ", "Ā", "퟿"]
encs: list[str] = ["utf-8", "ascii", "latin-1", "utf-16", "utf-16-le", "utf-16-be", "utf-32", "utf-8-sig"]
for s in samples:
    print(repr(s), [roundtrip(s, e) for e in encs])
for s in samples[:5]:
    print([len(s.encode(e)) for e in encs if roundtrip(s, e)])
raw: bytes = bytes(range(256))
print(raw.decode("latin-1") == "".join(map(chr, range(256))), len(raw.decode("latin-1").encode("utf-8")), raw[:128].decode("ascii") == "".join(map(chr, range(128))), raw.decode("utf-8", "replace").count("�"), raw.decode("utf-8", "ignore") == raw[:128].decode(), len(raw.decode("utf-8", "backslashreplace")))
