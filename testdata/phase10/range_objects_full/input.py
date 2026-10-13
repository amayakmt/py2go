r = range(2, 20, 3)
print(r, len(r), r[0], r[-1], r[1:3], r[::-1], list(r), 5 in r, 6 in r, r.start, r.stop, r.step, range(5), range(0), range(5, 0, -2), list(range(5, 0, -2)), bool(range(0)), bool(range(1)), type(r).__name__, r == range(2, 20, 3), range(0) == range(5, 5), range(3) == [0, 1, 2], repr(range(1, 2)), str(range(3)))
print(range(10 ** 20, 10 ** 20 + 3), len(range(10 ** 20, 10 ** 20 + 3)), list(range(10 ** 20, 10 ** 20 + 3)), 10 ** 20 + 1 in range(10 ** 20, 10 ** 21), range(0, 10 ** 30, 10 ** 29)[5], len(range(0, 10 ** 30, 10 ** 29)), range(-10 ** 20, 0)[-1], sum(range(10 ** 6)), len(range(0, 2 ** 70, 2 ** 69)))
print(sorted(range(5), reverse=True), [i * i for i in range(5)], reversed(range(3)) is not None, list(reversed(range(3))), tuple(range(2)), set(range(3)), max(range(5)), min(range(5)), sum(range(101)), list(zip(range(3), "abc")), dict(zip("ab", range(2))), 2.0 in range(3), True in range(3), "a" in range(3))
for fn in (lambda: range(1, 2, 0), lambda: r[10], lambda: len(range(2 ** 70))):
    try:
        print(fn())
    except (ValueError, TypeError, IndexError, OverflowError, AttributeError) as e:
        print(type(e).__name__, e)
