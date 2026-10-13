b = b"  Hello, World! hello  "
print(b.strip(), b.lstrip(), b.rstrip(), b.strip(b" H"), b.upper(), b.lower(), b.title(), b.swapcase(), b.strip().capitalize(), b.count(b"l"), b.count(b"l", 5), b.count(b""))
print(b.find(b"o"), b.rfind(b"o"), b.find(b"zzz"), b.index(b"W"), b.find(b"l", 6, 12), b.startswith(b"  H"), b.endswith(b"  "), b.startswith((b"x", b"  ")), b.find(101), b.rindex(b"hello"))
print(b.split(), b.split(b","), b.split(b"l", 2), b.rsplit(b"l", 2), b"a,b,,c".split(b","), b"  a  b ".split(), b"a b c".split(None, 1), b"a b c".rsplit(None, 1), b"abc".split(b"abc"))
print(b"-".join([b"a", b"b", b"c"]), b"".join([b"x", b"y"]), b" ".join([]), b"ab".join([b"1", b"2", b"3"]), b"abc".replace(b"b", b"XX"), b"aaa".replace(b"a", b"b", 2), b"abc".replace(b"", b"-"))
print(b"abc".partition(b"b"), b"abc".partition(b"x"), b"a-b-c".rpartition(b"-"), b"abc".center(7, b"*"), b"abc".ljust(5), b"abc".rjust(5, b"."), b"42".zfill(5), b"-42".zfill(6), b"line1\nline2\r\nline3".splitlines(), b"a\nb".splitlines(True), b"a\tb".expandtabs(4))
print(b"abc".isalpha(), b"abc1".isalnum(), b"123".isdigit(), b" \t".isspace(), b"ABC".isupper(), b"abc".islower(), b"Ab Cd".istitle(), b"abc".isascii(), b"\xff".isascii(), b"".isalpha(), b"Ab".isupper())
print(b"abc".translate(bytes.maketrans(b"abc", b"xyz")), b"hello".translate(None, b"l"), b"abc".removeprefix(b"a"), b"abc".removesuffix(b"c"), b"abc".removeprefix(b"x"))
print(b"caf\xc3\xa9".decode(), b"caf\xc3\xa9".decode("utf-8"), b"caf\xe9".decode("latin-1"), b"abc".decode("ascii"), b"\xff\xfe".decode("utf-16"), b"a\x00b\x00".decode("utf-16-le"), b"\x00a".decode("utf-16-be"))
for bad in (lambda: b"\xff".decode(), lambda: b"caf\xc3".decode(), lambda: b"\xc3\x28".decode(), lambda: b"\xe2\x82".decode(), lambda: b"caf\xc3\xa9".decode("ascii"), lambda: b"a".decode("nope"), lambda: b"a".find("a"), lambda: b"a".split(b""), lambda: b"a".join([1]), lambda: b"a".index(b"z"), lambda: b"abc".partition(b""), lambda: b"a".strip("a")):
    try:
        print(bad())
    except (UnicodeDecodeError, LookupError, TypeError, ValueError) as e:
        print(type(e).__name__, e)
print(b"\xff\xc3\x28ab".decode("utf-8", "replace"), b"\xff\xc3\x28ab".decode("utf-8", "ignore"), b"\xffab".decode("utf-8", "backslashreplace"), b"caf\xc3\xa9\xff".decode("ascii", "replace"))
