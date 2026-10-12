print([x * x for x in range(6)])
print([x for x in range(10) if x % 3 == 0])
print([x * 2 for x in range(10) if x % 2 == 1 if x > 3])
print([c.upper() for c in "abc"])
print([len(w) for w in "the quick brown fox".split()])
xs: list[int] = [3, 1, 2]
print([x + 1 for x in xs], [x for x in xs if x > 1], [str(x) for x in xs])
print([x for x in xs if x > 5])
pairs: list[tuple[int, str]] = [(1, "a"), (2, "b")]
print([n for n, s in pairs], [s * n for n, s in pairs])
