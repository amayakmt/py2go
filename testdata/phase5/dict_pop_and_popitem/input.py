d: dict[str, int] = {"a": 1, "b": 2, "c": 3}
print(d.pop("b"), d)
print(d.pop("zz", -1), d)
print(d.popitem(), d)
d["d"] = 4
print(d.popitem(), d.popitem(), d)
