pts: list[tuple[int, int]] = [(3, 1), (1, 2), (2, 0)]
pts.sort()
print(pts)
xs: list[int] = []
ys: list[int] = []
for x, y in pts:
    xs.append(x)
    ys.append(y)
print(xs, ys)
best: tuple[int, int] = pts[0]
for p in pts:
    if p[1] > best[1]:
        best = p
print(best)
pairs: tuple[list[int], str] = ([1, 2], "k")
pairs[0].append(3)
print(pairs, len(pairs[0]))
same: bool = (1, [2]) == (1, [2])
print(same)
