s = "  The quick brown fox\tjumps\nover the lazy dog  "
print(s.split(), s.split(" "), s.split(None, 2), s.rsplit(None, 2), s.strip().split("o"), s.strip().rsplit("o", 1), s.splitlines(), s.expandtabs(4)[:30], s.strip().partition(" "), s.strip().rpartition(" "))
print("a b\x85c\x0bd\x0ce\x1cf\x1dg\x1eh\r\ni\rj".splitlines(), "a\n".splitlines(), "\n\n".splitlines(True), "".splitlines(), "ab".splitlines(True), "a\r\nb".splitlines(True))
print("abc".center(10, "*"), "abc".center(2), "abcd".center(7, "-"), "ab".ljust(5, "."), "ab".rjust(5, "."), "-12".zfill(6), "+5".zfill(4), "abc".zfill(2), "".zfill(3), "é".zfill(3))
print("abcabc".find("bc"), "abcabc".rfind("bc"), "abcabc".find("bc", 2), "abcabc".find("x"), "abcabc".index("c"), "abcabc".rindex("a"), "abcabc".count("bc"), "aaaa".count("aa"), "abc".count(""), "héé".find("é"), "héé".rfind("é"))
print("hello".replace("l", "L"), "hello".replace("l", "L", 1), "hello".replace("", "-"), "hello".replace("hello", ""), "aaa".replace("aa", "b"), "hello".translate({104: "J", 101: None, 108: 76}), "abc".translate(str.maketrans("ab", "xy")), "abc".translate(str.maketrans("a", "b", "c")), str.maketrans({"a": "1", 98: "2"}))
print("12".isdigit(), "²".isdigit(), "²".isdecimal(), "²".isnumeric(), "½".isnumeric(), "٣".isdecimal(), "abc".isalpha(), "abc1".isalnum(), "é".isalpha(), " \t\n".isspace(), "".isspace(), "a_b".isidentifier(), "1a".isidentifier(), "".isidentifier(), "é".isidentifier())
print("abc".isprintable(), "a\nb".isprintable(), "".isprintable(), " ".isprintable(), "é".isprintable(), "é".isascii(), "abc".isascii(), "".isascii(), "ǆ".isprintable(), "\x7f".isprintable(), " ".isprintable())
print("a,b;c".split(","), "a,b,".split(","), ",a".split(","), "abc".split(""if False else "b"), "a  b".split(" "), "ab".startswith(("x", "a")), "ab".endswith(("b",)), "abc".startswith("bc", 1), "abc".endswith("ab", 0, 2), "ab".startswith(""), "abc".removeprefix("ab"), "abc".removesuffix("bc"))
for f in (lambda: "a".split(""), lambda: "a".center(), lambda: "a".replace("a"), lambda: "a".index("z"), lambda: "a".startswith(1), lambda: "a".join(1), lambda: "a".join([1]), lambda: "a".zfill("x"), lambda: "a".encode(1), lambda: "a".partition(""), lambda: "abc".translate(5), lambda: str.maketrans("ab", "c"), lambda: "x".expandtabs("a"), lambda: "a".count(1)):
    try:
        print(f())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
