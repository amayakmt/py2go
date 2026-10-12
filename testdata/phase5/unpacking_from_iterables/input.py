d: dict[str, int] = {"a": 1, "b": 2}
for (k, v) in d.items():
    print(k, v)
for i, k in enumerate(d):
    print(i, k)
for a, b in zip("xy", [1, 2]):
    print(a, b)
pairs: list[tuple[str, int]] = list(d.items())
print(pairs)
k1, v1 = pairs[0]
print(k1, v1)
first, second = d
print(first, second)
p, q = {5, 6}
print(p, q)
one, two, three = range(3)
print(one, two, three)
