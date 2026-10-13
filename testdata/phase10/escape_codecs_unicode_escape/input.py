for s in ["plain", "tab\t", "é", "日本", "😀", "\x00\x7f", "a\\b", "quote'\"", " "]:
    e = s.encode("unicode-escape")
    print(repr(s), e, e.decode("unicode-escape") == s, s.encode("raw-unicode-escape"), s.encode("ascii", "backslashreplace"), s.encode("latin-1", "xmlcharrefreplace") if True else "")
print(b"\\u00e9\\n\\x41\\U0001f600".decode("unicode-escape"), repr(b"\\q".decode("unicode-escape")) if False else "skip", b"\\xe9".decode("raw-unicode-escape"), b"\\u00e9".decode("raw-unicode-escape"))
for bad in (b"\\x4", b"\\u12", b"\\U00110000", b"\\N{bogus}" if False else b"\\x"):
    try:
        print(bad.decode("unicode-escape"))
    except UnicodeDecodeError as e:
        print(e)
