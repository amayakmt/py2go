print("ab" * 3, 3 * "ab", "ab" * 0, "ab" * -1, "" * 10 ** 6, "é" * 4, ("a",) * 3, [0] * 3, [[0]] * 2, b"x" * 3, bytearray(b"y") * 2, [1, 2] * 0, (1, 2) * 2, True * "ab", "ab" * True, "ab" * False)
for f in (lambda: "a" * 1.5, lambda: "a" * "b", lambda: "a" * None, lambda: [1] * 1.5, lambda: "a" * 10 ** 30, lambda: [0] * 10 ** 30, lambda: b"a" * 10 ** 30, lambda: "a" * (10 ** 12) if False else "skip", lambda: (1,) * "a", lambda: {} * 2):
    try:
        r = f()
        print(r if isinstance(r, str) and len(r) < 20 else type(r).__name__)
    except (TypeError, OverflowError, MemoryError) as e:
        print(type(e).__name__, e)
x = [[0] * 3 for _ in range(2)]
x[0][0] = 1
y = [[0] * 3] * 2
y[0][0] = 1
print(x, y, "-" * 20, "=" * 0 + "|", "ab" * 2 + "cd" * 2)
