s = "abracadabra abra"
print(s.find("abra"), s.find("abra", 1), s.rfind("abra"), s.find("abra", 1, 5), s.find(""), s.find("", 100), s.find("", 5, 3), s.rfind(""), s.index("cad"), s.rindex("a"), s.count("a"), s.count("abra"), s.count(""), s.count("a", -5), s.count("a", 2, -2), s.count("zz"))
for f in (lambda: s.index("zz"), lambda: s.rindex("zz"), lambda: s.find(1), lambda: s.count(1), lambda: s.find("a", "b"), lambda: s.index("a", 100)):
    try:
        print(f())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
print(s.replace("abra", "X"), s.replace("abra", "X", 1), s.replace("abra", "X", 0), s.replace("abra", "X", -1), s.replace("a", ""), s.replace("", "-"), "abc".replace("", "-", 2), s.replace("zz", "y"), "aaa".replace("aa", "b"), "".replace("", "x"), "".replace("a", "x"), "日本日".replace("日", "月"))
print(s.startswith("abra"), s.startswith("abra", 1), s.startswith(("x", "abra")), s.endswith("abra"), s.endswith(("x", "y")), s.endswith("abra", 0, 4), s.startswith(""), s.startswith("", 100), s.endswith("", 0, 0), s.startswith("a", -4), "abc".startswith("abcd"))
for f in (lambda: s.startswith(1), lambda: s.startswith(("a", 1)), lambda: s.endswith(["a"]), lambda: s.replace("a"), lambda: s.replace(1, 2), lambda: s.replace("a", "b", "c")):
    try:
        print(f())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
