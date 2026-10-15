for enc in ["cp437", "cp1252", "latin-1", "iso8859-15", "koi8-r", "cp850", "mac-roman", "windows-1251"]:
    s = "café üß"
    try:
        b = s.encode(enc)
        print(enc, b, b.decode(enc))
    except UnicodeEncodeError as e:
        print(enc, "enc error", e)
print("€".encode("cp1252"), b"\x80".decode("cp1252"), b"\xa4".decode("iso8859-15"), "Ж".encode("cp1251"), "Ж".encode("koi8-r"))
for bad in [lambda: b"\x81".decode("cp1252"), lambda: "中".encode("cp437"), lambda: "中".encode("cp1252", "strict")]:
    try:
        bad()
    except (UnicodeDecodeError, UnicodeEncodeError) as e:
        print(type(e).__name__, e)
print("中 x".encode("cp437", "replace"), "中 x".encode("cp437", "ignore"), "中 x".encode("cp437", "backslashreplace"), "中".encode("cp437", "xmlcharrefreplace"))
print(b"a\x81b".decode("cp1252", "replace"), b"a\x81b".decode("cp1252", "ignore"), b"a\x81b".decode("cp1252", "backslashreplace"))
print(b"caf\xe9".decode("utf-8", "surrogateescape").encode("utf-8", "surrogateescape"))
