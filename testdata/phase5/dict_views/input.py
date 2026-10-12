d: dict[str, int] = {"a": 1, "b": 2}
print(d.keys(), d.values(), d.items())
d["c"] = 3
print(d.keys(), len(d.keys()), len(d.values()), len(d.items()))
print("a" in d.keys(), 3 in d.values(), 9 in d.values(), ("a", 1) in d.items(), ("a", 2) in d.items())
print(list(d.keys()), list(d.values()), list(d.items()))
print(sorted(d.items()), sum(d.values()), max(d.values()), min(d.keys()))
print(sorted(d.keys() & {"a", "z"}), sorted(d.keys() | {"q"}), sorted(d.keys() - {"a"}), sorted(d.keys() ^ {"a", "q"}))
n: dict[int, str] = {1: "a", 2: "b", 3: "c"}
print(n.keys() & {3, 2, 9}, n.keys() | [7, 1], n.keys() - {2}, n.keys() ^ {3, 4})
print(set(d.keys()) == {"a", "b", "c"}, d.keys() == {"a", "b", "c"})
