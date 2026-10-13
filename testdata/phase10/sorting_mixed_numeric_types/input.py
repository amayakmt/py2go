xs = [3, 1.5, True, 2 ** 70, -2 ** 70, 0, False, 2.5, 10 ** 20, -0.0, 1, 1.0, float("inf"), -float("inf"), 2 ** 53 + 1, float(2 ** 53)]
print(sorted(xs))
print(sorted(xs, reverse=True))
print(sorted(xs, key=lambda v: -v), sorted(xs, key=abs), sorted([x for x in xs if abs(x) != float("inf")], key=lambda v: (v % 3, v)))
print(min(xs), max(xs), min(xs, key=abs), max(xs, key=lambda v: v % 7), sum(x for x in xs if abs(x) < 100), sum(x for x in xs if isinstance(x, int)))
ys = [2 ** 70 + i % 5 for i in range(20)]
print(sorted(set(ys)), ys.count(2 ** 70 + 3), ys.index(2 ** 70 + 4), max(ys) - min(ys), sorted(ys, reverse=True)[:3], ys == sorted(ys), ys[::-1] == list(reversed(ys)))
d = {2 ** 70: "a", 2 ** 70 + 1: "b", 1: "c", 1.0: "d", 2 ** 63: "e", -2 ** 63: "f", -2 ** 63 - 1: "g"}
print(d, sorted(d), d[2 ** 70], d[float(2 ** 70)], 2 ** 63 in d, 2.0 ** 63 in d, len(d), {k: v for k, v in sorted(d.items())})
s = {2 ** 100, 2 ** 100 + 0, 2 ** 99 * 2, float(2 ** 100), 1 << 100}
print(s, len(s), 2 ** 100 in s, (2 ** 100) + 1 in s)
print([2 ** 70] * 2 == [2 ** 70, 2 ** 70], (2 ** 70, 1) < (2 ** 70, 2), [2 ** 70] < [2 ** 71], "".join(sorted(map(str, [2 ** 70, 3, 10 ** 30]))), sorted(["10", "9", "2"], key=int), sorted(["10", "9", "2"]), sorted([2 ** 70, 3, 5], key=str))
for bad in ([1, "a"], [1j, 2j], [None, 1], [[1], 1]):
    try:
        sorted(bad)
    except TypeError as e:
        print(e)
