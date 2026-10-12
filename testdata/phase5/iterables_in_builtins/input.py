d: dict[int, str] = {3: "c", 1: "a", 2: "b"}
s: set[int] = {7, 5, 6}
print(sorted(d), sorted(s), sorted(d.values()), sorted(d.items()))
print(sum(s), min(d), max(d), sum(d), max(d.values()))
print(list(d), list(s), list(d.values()), list(d.items()))
print(len(d.items()), len(s), len(range(7)))
print(any(d), all(d), any(s), any(d.values()), all(s))
print(any([0, 0]), all([1, 2]), all(["", "a"]), any([[1], [1]]))
print(set(d.values()) == {"a", "b", "c"}, sorted(set("hello")))
print(",".join(d.values()), "-".join(sorted(set("banana"))))
print(list("abc"), list(range(3)), set(range(3)))
