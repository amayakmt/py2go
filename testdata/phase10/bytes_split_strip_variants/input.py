b = b"  ab  cd\t\nef  "
print(b.split(), b.split(None, 1), b.rsplit(None, 1), b.split(b" "), b.split(b"  "), b.strip(), b.lstrip(), b.rstrip(), b.strip(b" \t"), b"xxabxx".strip(b"x"), b"abc".strip(b""), b"abc".strip(b"abc"))
print(b"a,b,,c".split(b","), b"a,b,,c".split(b",", 1), b"a,b,,c".rsplit(b",", 1), b",".split(b","), b"".split(b","), b"abc".split(b"abc"), b"a--b".split(b"--"), b"abc".split(b"x"), b"".split())
print(b"line1\nline2\r\nline3\rline4\n".splitlines(), b"line1\nline2\r\nline3\n".splitlines(True), b"".splitlines(), b"\n\n".splitlines(), b"a\x0bb\x0cc\x1cd".splitlines())
print(b"key=value=x".partition(b"="), b"key=value=x".rpartition(b"="), b"nothing".partition(b"="), b"nothing".rpartition(b"="), b"=".partition(b"="))
for f in (lambda: b"abc".split(b""), lambda: b"abc".split("a"), lambda: b"abc".strip("a"), lambda: b"abc".partition(b""), lambda: b"abc".replace("a", b"b"), lambda: b"a".join(["b"]), lambda: b"a".join(5)):
    try:
        print(f())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
