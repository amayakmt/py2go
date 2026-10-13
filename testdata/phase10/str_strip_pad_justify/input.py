s = "  \t xyz \n "
print(repr(s.strip()), repr(s.lstrip()), repr(s.rstrip()), "xxhixx".strip("x"), "xxhixx".lstrip("x"), "xxhixx".rstrip("x"), "abcba".strip("ab"), "abc".strip(""), "abc".strip("abc"), " x ".strip(), "\x1fx\x1f".strip(), "xyz".strip(None), "...a.b...".strip("."))
print("ab".center(7, "*"), "ab".center(6, "*"), "ab".center(5, "*"), "abc".center(2), "ab".ljust(5, "."), "ab".rjust(5, "."), "ab".zfill(5), "-ab".zfill(5), "+1".zfill(4), "".zfill(3), "12".zfill(1), "é".center(5, "ü"), "x".center(4), "x".center(3, "é"))
for f in (lambda: "x".center(5, ""), lambda: "x".center(5, "ab"), lambda: "x".ljust("5"), lambda: "x".rjust(3, 1), lambda: "x".zfill("3"), lambda: "x".strip(1)):
    try:
        print(f())
    except TypeError as e:
        print(e)
print("a\tb".expandtabs(), "a\tb".expandtabs(4), "ab\tc\td".expandtabs(3), "a\nb\tc".expandtabs(4), "\t".expandtabs(0), "a\tb".expandtabs(-1), "日\tb".expandtabs(4), "a\r\tb".expandtabs(8))
print("hello".removeprefix("he"), "hello".removesuffix("lo"), "hello".removeprefix(""), "hello".removeprefix("hello!"), "hello".removesuffix("x"), "aaa".removeprefix("a"), "abc".removeprefix("abc"))
