tests = [
    lambda: 1 / 0, lambda: 1 // 0, lambda: 1 % 0, lambda: 1.0 / 0, lambda: 10 ** 30 / 0, lambda: 10 ** 30 // 0, lambda: divmod(1, 0), lambda: 0 ** -1, lambda: 0.0 ** -1, lambda: 1j / 0,
    lambda: int("x"), lambda: int(""), lambda: int("1.0"), lambda: int(None), lambda: int([1]), lambda: float("x"), lambda: float(None), lambda: float([]), lambda: int("1", 99), lambda: int(1, 10),
    lambda: "a" + 1, lambda: 1 + "a", lambda: "a" * "b", lambda: "a" * 1.5, lambda: [1] + (2,), lambda: (1,) + [2], lambda: {} + {}, lambda: None + 1, lambda: abs("a"), lambda: len(5), lambda: len(None),
    lambda: chr(-1), lambda: chr(0x110000), lambda: chr("a"), lambda: ord(""), lambda: ord("ab"), lambda: ord(1), lambda: round("a"), lambda: round(1.5, "a"), lambda: pow(1, 2, 0), lambda: pow("a", 2),
    lambda: bin(1.5), lambda: hex("a"), lambda: oct(None), lambda: sum(["a"]), lambda: sum([1], "a"), lambda: max([]),
    lambda: "abc".index("z"), lambda: [1].index(2), lambda: [1].pop(5), lambda: {}.pop("k"), lambda: {1: 2}[3], lambda: [1][5], lambda: "abc"[5], lambda: b"abc"[5], lambda: range(3)[5], lambda: {1}.remove(2),
    lambda: "abc".encode("bogus"), lambda: b"\xff".decode(), lambda: "x".join(5), lambda: "x".join([1]), lambda: "%d" % "a", lambda: "{}".format(), lambda: "{0}{1}".format(1), lambda: lambda: format(1, "s"), lambda: format("a", "d"),
]
for i, t in enumerate(tests):
    try:
        r = t()
        print(i, "ok", r if not callable(r) else "callable")
    except Exception as e:
        print(i, type(e).__name__, e)
