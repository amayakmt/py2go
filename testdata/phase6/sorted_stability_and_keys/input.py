recs: list[tuple[str, int]] = [("a", 2), ("b", 1), ("c", 2), ("d", 1), ("e", 2)]
print(sorted(recs, key=lambda r: r[1]))
print(sorted(recs, key=lambda r: r[1], reverse=True))
print(sorted(["B", "a", "C"], key=lambda s: s.lower()), sorted(["B", "a", "C"]))
fl: list[float] = [2.5, -1.0, 3.25]
print(sorted(fl, key=lambda x: -x), sorted(fl, key=abs))
print(sorted([True, False, True], key=lambda b: not b))
print(sorted(["10", "9", "100"], key=int), sorted(["10", "9", "100"]))
