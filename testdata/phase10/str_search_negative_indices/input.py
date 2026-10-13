s = "héllo wörld héllo"
print(s.find("h", -6), s.find("l", -5, -1), s.find("", 40), s.find("", 8, 3), s.count("l", -8), s.count("l", 2, -2), s.count("", 3, 5), s.count("", 40), s.index("w", 3), s.rfind("l", 0, -1), s.rfind("é", 4), s.startswith("hé", -17), s.startswith("", 40), s.endswith("llo", -3), s.endswith("ll", 0, -1))
for f in (lambda: s.find("a", "x"), lambda: s.count("a", 1.5), lambda: s.index("zz"), lambda: s.rindex("zz", 2), lambda: s.find("l", 10 ** 30), lambda: s.find("l", -10 ** 30), lambda: s.startswith("h", None, None), lambda: s.find("l", None, 4)):
    try:
        print(f())
    except (TypeError, ValueError) as e:
        print(type(e).__name__, e)
