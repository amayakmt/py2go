def uniq_pairs(xs: list[int]) -> set[frozenset[int]]:
    return {frozenset((a, b)) for a in xs for b in xs if a < b}

pairs = uniq_pairs([1, 2, 3, 4])
print(len(pairs), sorted(sorted(p) for p in pairs), frozenset({1, 2}) in pairs, frozenset({2, 1}) in pairs, frozenset({1, 5}) in pairs)
r: range = range(10, 0, -3)
print(r, list(r), len(r), r[0], r[-1], 7 in r, 8 in r, sum(r), max(r), min(r), r[1:], reversed(r) is not None, sorted(r), [x * 2 for x in r])
xs: list[int] = list(range(20))
sl = slice(2, 15, 4)
print(xs[sl], xs[slice(None, 5)], xs[slice(-3, None)], xs[::5], xs[15:3:-4], "abcdefghij"[sl], (1, 2, 3, 4, 5)[1:4], sl.indices(10), sl.start, sl.stop, sl.step)
ys = xs[:]
ys[sl] = [-1, -2, -3, -4]
del ys[::5]
print(ys, len(ys), ys[-1], ys.index(-2), ys.count(-1))
