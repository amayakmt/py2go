s = "  alpha  beta\tgamma\n\ndelta  "
print(s.split(), s.split(" "), s.split(None, 2), s.rsplit(None, 2), s.split("\n"), s.strip().split("  "), "a,b,,c".split(","), "a,b,,c".split(",", 1), "a,b,,c".rsplit(",", 1), ",".split(","), "".split(","), "".split(), "abc".split("abc"), "aXXbXXc".split("XX"), "abc".split(""if False else "b"))
for f in (lambda: "abc".split(""), lambda: "abc".split(1), lambda: "abc".rsplit(""), lambda: ",".join([1, 2]), lambda: ",".join(5), lambda: "x".partition(""), lambda: "x".split(maxsplit="a")):
    try:
        print(f())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
print("a-b-c".partition("-"), "a-b-c".rpartition("-"), "abc".partition("x"), "abc".rpartition("x"), "abc".partition("abc"), "".partition("x"), "--".partition("-"))
print("\n".join(["a", "b"]), "".join(["a", "b"]), "-".join("xyz"), ", ".join(map(str, range(4))), "-".join([]), "-".join(["only"]), "ab".join(["1", "2", "3"]), " ".join(("t", "u")), ",".join({"k": 1, "j": 2}), "/".join(sorted({"b", "a"})))
print("a\nb\r\nc\rd\x0be\x0cf\x1cg\x1dh\x1ei\x85j k l".splitlines(), "a\nb\n".splitlines(True), "a\n\nb".splitlines(), "\n".splitlines(), "".splitlines(), "ab\r\n".splitlines(True), "x".splitlines(keepends=True))
print("one two  three".split(" ", 1), "k=v=w".split("=", 1), "k=v=w".rsplit("=", 1), "  lead".split(None, 1), "trail  ".rsplit(None, 1), "a b c d".split(None, 0), "a b c d".rsplit(None, 0), "a b".split(maxsplit=-1))
