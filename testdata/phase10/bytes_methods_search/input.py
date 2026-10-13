b = b"abracadabra"
print(b.find(b"bra"), b.rfind(b"bra"), b.find(b"x"), b.find(b"a", 1), b.find(b"a", 1, 3), b.find(97), b.index(b"cad"), b.rindex(b"a"), b.count(b"a"), b.count(b"abra"), b.count(b""), b.count(b"a", 2, 8), b.startswith(b"abra"), b.endswith(b"bra"), b.startswith((b"x", b"ab")), b.endswith(b"a", 0, -1), b.startswith(b"bra", 1))
for f in (lambda: b.index(b"x"), lambda: b.rindex(b"x"), lambda: b.find("a"), lambda: b.count("a"), lambda: b.startswith("a"), lambda: b.find(300), lambda: b.index(b"z", 2)):
    try:
        print(f())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
print(b.replace(b"a", b"A"), b.replace(b"a", b"", 2), b.replace(b"", b"-"), b.replace(b"abra", b"X", 1), b.partition(b"c"), b.rpartition(b"a"), b.partition(b"z"), b.rpartition(b"z"))
print(b.split(b"a"), b.split(b"a", 2), b.rsplit(b"a", 2), b"  a  b  ".split(), b" a b ".split(b" "), b"a\tb\nc  d".split(), b"a  b".split(None, 1), b"".split(), b"".split(b","), b"abc".split(b"abc"))
print(b"a\nb\r\nc\rd\n".splitlines(), b"a\nb\r\nc".splitlines(True), b"".splitlines(), b"x".splitlines())
print(b.strip(b"ab"), b.lstrip(b"ab"), b.rstrip(b"ab"), b"  x \n".strip(), b"xxhixx".strip(b"x"), b"abc".removeprefix(b"ab"), b"abc".removesuffix(b"bc"), b"abc".removeprefix(b"x"))
print(b.upper(), b"ABC".lower(), b"abc".title(), b"hello world".capitalize(), b"AbC".swapcase(), b"abc".isalpha(), b"a1".isalnum(), b"12".isdigit(), b" \t".isspace(), b"ABC".isupper(), b"abc".islower(), b"Ab Cd".istitle(), b"".isalpha(), b"\xe9".isalpha(), b"caf\xc3\xa9".upper())
print(b"ab".center(6, b"*"), b"ab".ljust(5, b"-"), b"ab".rjust(5), b"42".zfill(5), b"-42".zfill(6), b"+4".zfill(4), b"a\tb".expandtabs(4), b"abc".center(2))
