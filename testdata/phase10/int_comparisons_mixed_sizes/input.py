vals = [-(2 ** 70), -(2 ** 63) - 1, -(2 ** 63), -1, 0, 1, 2 ** 53, 2 ** 53 + 1, 2 ** 63 - 1, 2 ** 63, 2 ** 70]
fl = [float(v) for v in vals] + [1e19, -1e19, 1e30, 0.5, -0.5, float("inf"), float("-inf")]
for a in vals:
    print([a < b for b in vals], [a == b for b in vals])
for a in vals:
    print([a < f for f in fl], [a == f for f in fl], [a >= f for f in fl])
print(sorted(vals + fl[len(vals):]) == sorted(vals + fl[len(vals):], key=lambda x: x), max(vals), min(vals), vals.index(2 ** 63), 2 ** 70 in vals, float(2 ** 70) in vals)
print(2 ** 53 + 1 == 2.0 ** 53, 2 ** 53 == 2.0 ** 53, 2 ** 1024 > 1e308, -2 ** 1024 < -1e308, 2 ** 1024 > float("inf"), 10 ** 400 != float("inf"))
