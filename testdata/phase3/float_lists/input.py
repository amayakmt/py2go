xs: list[float] = []
for i in range(1, 6):
    xs.append(1 / i)
print(xs)
print(sum(xs), sum(xs) / len(xs), max(xs), min(xs))
ys: list[float] = []
for x in xs:
    ys.append(round(x * 100) / 100)
print(ys)
print([0.1 + 0.2, 0.1 * 3, 1e16, 1e-7])
zs: list[float] = [3.0, 1.0, 2.0]
zs.sort()
print(zs, zs.index(2.0), 2.0 in zs)
