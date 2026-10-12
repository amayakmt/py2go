a: dict[str, int] = {"x": 1, "y": 2}
b: dict[str, int] = {"y": 20, "z": 30}
a.update(b)
print(a)
a.update([("w", 0), ("x", 100)])
print(a)
a.update(zip(["p", "q"], [7, 8]))
print(a)
c: dict[str, int] = {}
c.update({"only": 1})
print(c)
