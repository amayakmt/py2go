# 0.0 and -0.0 compare equal but print differently, so the output shows
# whether equal items kept their original order.
z: float = 0.0
xs: list[float] = [z, -z, 1.0, z, -z, -1.0]
print(sorted(xs))
print(sorted(xs, reverse=True))
ys: list[float] = xs.copy()
ys.sort()
print(ys)
ys.sort(reverse=True)
print(ys)
pairs: list[tuple[float, int]] = [(z, 1), (-z, 2), (z, 3)]
pairs.sort()
print(pairs)
