text = "ab é € 😀 ퟿ end"
for enc in ("ascii", "latin-1", "utf-8", "utf-16", "utf-32"):
    for errs in ("strict", "ignore", "replace", "backslashreplace", "xmlcharrefreplace"):
        try:
            print(enc, errs, text.encode(enc, errs))
        except UnicodeEncodeError as e:
            print(enc, errs, "UnicodeEncodeError", e)
data = b"ab \xe9 \xff\xfe \xc3\xa9 \xe2\x82 \xf0\x9f\x98 end \x80"
for errs in ("strict", "ignore", "replace", "backslashreplace"):
    try:
        print(errs, repr(data.decode("utf-8", errs)))
    except UnicodeDecodeError as e:
        print(errs, "UnicodeDecodeError", e)
print(repr(data.decode("latin-1")), repr(data.decode("ascii", "replace")), repr(b"\xff\xfe\x41\x00".decode("utf-16")), repr(b"\x00\x41".decode("utf-16-be")), repr(b"A\x00\x00\x00".decode("utf-32-le")))
for f in (lambda: "x".encode("utf-8", "bogus"), lambda: "é".encode("ascii", "bogus"), lambda: b"\xff".decode("utf-8", "bogus"), lambda: b"a".decode("utf-8", 5)):
    try:
        print(f())
    except (LookupError, TypeError) as e:
        print(type(e).__name__, e)
