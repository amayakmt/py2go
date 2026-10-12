d: dict[str, int] = {"a": 1, "b": 2, "c": 3}
del d["b"]
print(d)
if "z" in d:
    del d["z"]
d.clear()
print(d, len(d))
xs: list[int] = [1, 2, 3, 4, 5, 6]
del xs[0]
del xs[-1]
print(xs)
del xs[1:3]
print(xs)
del xs[::2]
print(xs)
