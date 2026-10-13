s = "héllo wörld €😀"
for enc in ("utf-8", "utf8", "UTF-8", "latin-1", "utf-16", "utf-16-le", "utf-16-be", "utf-32", "utf-32-le", "utf-32-be", "unicode-escape"):
    try:
        data = s.encode(enc)
        print(enc, data, len(data), data.decode(enc) == s)
    except (UnicodeEncodeError, UnicodeDecodeError) as e:
        print(enc, type(e).__name__, e)
print("abc".encode(), "abc".encode("ascii"), "é".encode("latin-1"), "é".encode("utf-8"), "".encode(), "\x00\x7f".encode(), "€".encode("utf-8"), "😀".encode("utf-16-le"))
for errors in ("strict", "ignore", "replace", "backslashreplace", "xmlcharrefreplace"):
    try:
        print(errors, "héllo €😀".encode("ascii", errors), "héllo €".encode("latin-1", errors))
    except UnicodeEncodeError as e:
        print(errors, type(e).__name__, e)
for bad in (lambda: "x".encode("nope"), lambda: "x".encode("utf-8", "nope"), lambda: "é€".encode("ascii"), lambda: "€".encode("latin-1"), lambda: "ab€c€d".encode("ascii"), lambda: b"x".decode(1)):
    try:
        print(bad())
    except (UnicodeEncodeError, LookupError, TypeError) as e:
        print(type(e).__name__, e)
print("a".encode("utf-8-sig"), b"\xef\xbb\xbfa".decode("utf-8-sig"), "é\n\t\\".encode("unicode-escape"), b"\\u20ac\\xe9\\n".decode("unicode-escape"), "é€".encode("raw-unicode-escape"))
print(str(b"abc", "utf-8"), str(b"\xc3\xa9", "utf-8"), str(bytearray(b"hi"), "ascii"), bytes("é", "utf-8") == "é".encode(), len("😀".encode()), len("😀"))
