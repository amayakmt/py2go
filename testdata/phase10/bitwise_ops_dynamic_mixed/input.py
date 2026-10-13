vals = [0, 1, -1, 2, -2, 255, -256, 2 ** 31, -2 ** 31, 2 ** 32, 2 ** 63 - 1, -2 ** 63, 2 ** 64, -2 ** 64, 2 ** 100 + 1, -2 ** 100 - 1, True, False]
for a in vals[:12]:
    print(a, [a & b for b in vals[::3]], [a | b for b in vals[::3]], [a ^ b for b in vals[::3]])
for n in (0, 1, 5, 31, 32, 63, 64, 65, 100):
    print(n, [v << n for v in vals[:6]], [v >> n for v in vals[6:12]])
print([~v for v in vals], [-v for v in vals], [v ** 2 for v in vals[:8]], [v // 3 for v in vals], [v % 7 for v in vals], [divmod(v, -5) for v in vals[:8]])
for bad in (lambda: 1 & 1.5, lambda: 1.5 | 1, lambda: 1 ^ "a", lambda: 1 << 1.0, lambda: "a" >> 1, lambda: [1] & [1], lambda: None | 1, lambda: 1 << 10 ** 30, lambda: 2 ** 10 ** 30 if False else 1 >> 10 ** 30):
    try:
        print(bad())
    except (TypeError, ValueError, OverflowError, MemoryError) as e:
        print(type(e).__name__, e)
