s = slice(1, 5, 2)
print(s, s.start, s.stop, s.step, slice(3), slice(3).start, slice(3).stop, slice(None), slice(1, None), repr(slice(None, None, -1)), type(s).__name__, s == slice(1, 5, 2), s != slice(1, 5))
xs = list(range(10))
for sl in (slice(2, 5), slice(None, None, 3), slice(None, None, -1), slice(8, 2, -2), slice(-3, None), slice(None, -3), slice(100), slice(-100, 100, 5), slice(5, 2), slice(None, None, None)):
    print(sl, xs[sl], "abcdefghij"[sl], tuple(xs)[sl], bytes(xs)[sl], sl.indices(10), sl.indices(0), sl.indices(3))
ys = list(range(10))
ys[slice(1, 4)] = ["a"]
del ys[slice(0, 3, 2)]
print(ys, ys[slice(None, None, -2)])
for bad in (lambda: slice(), lambda: slice(1, 2, 3, 4), lambda: slice(1, 2, 0).indices(5), lambda: slice(1.5).indices(5), lambda: slice(1).indices(-1), lambda: xs[slice(1, 2, 0)], lambda: [1][slice("a")]):
    try:
        print(bad())
    except (TypeError, ValueError) as e:
        print(type(e).__name__, e)
print(slice(1, 10 ** 30).indices(5), slice(-10 ** 30, 10 ** 30).indices(7), xs[10 ** 30:], xs[:-10 ** 30], xs[::10 ** 30], xs[::-10 ** 30])
