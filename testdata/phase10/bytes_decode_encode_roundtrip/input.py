texts = ["", "abc", "héllo", "日本語", "😀 emoji", "a\u0000b", "ÿĀ", "tab\there"]
for t in texts:
    u8, u16, u32, l1 = t.encode(), t.encode("utf-16"), t.encode("utf-32-le"), None
    try:
        l1 = t.encode("latin-1")
    except UnicodeEncodeError as e:
        l1 = str(e)
    print(repr(t), u8, u16.hex(), u32.hex(), l1, u8.decode() == t, u16.decode("utf-16") == t, len(u8), len(t))
for enc in ("utf-8", "UTF8", "utf_8", "ascii", "latin1", "iso-8859-1", "utf-16-le", "utf-16-be", "utf-32", "utf-8-sig", "US-ASCII", "cp1252" if False else "latin-1"):
    try:
        print(enc, "héé".encode(enc), "abc".encode(enc))
    except UnicodeEncodeError as e:
        print(enc, "ERR", e)
print("é".encode("ascii", "ignore"), "é".encode("ascii", "replace"), "é€".encode("ascii", "backslashreplace"), "é€".encode("ascii", "xmlcharrefreplace"), "é".encode("latin-1", "strict"))
print(b"caf\xc3\xa9".decode(), b"caf\xe9".decode("latin-1"), b"caf\xe9".decode("utf-8", "replace"), b"caf\xe9".decode("utf-8", "ignore"), b"caf\xe9".decode("utf-8", "backslashreplace"), b"\xff\xfe".decode("utf-16"), b"\xef\xbb\xbfhi".decode("utf-8-sig"), b"\xef\xbb\xbfhi".decode("utf-8") == "﻿hi")
for f in (lambda: b"caf\xe9".decode(), lambda: b"\xff".decode("ascii"), lambda: b"\xc3".decode(), lambda: b"\xe2\x82".decode(), lambda: b"a".decode("nope"), lambda: "a".encode("nope"), lambda: b"\xf0\x28\x8c\x28".decode(), lambda: "é".encode("ascii")):
    try:
        print(f())
    except (UnicodeError, LookupError) as e:
        print(type(e).__name__, e)
