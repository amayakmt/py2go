a: dict[str, int] = {"x": 1, "y": 2}
b: dict[str, int] = {"y": 2, "x": 1}
print(a == b, a != b, a == {"x": 1})
c: dict[str, int] = a
c["z"] = 3
print(a, a == c)
d: dict[str, int] = a.copy()
d["w"] = 0
print(a, d)
e: dict[str, int] = dict(a)
e.clear()
print(a, e)
nested: dict[str, list[int]] = {"k": [1]}
shallow: dict[str, list[int]] = nested.copy()
shallow["k"].append(2)
print(nested)
