for v in (0.5, 1.5, 2.5, 3.5, -0.5, -1.5, -2.5, 0.49999999999999994, 2.675, 1.005, 0.285, 1.45, 8.345, 1e15 + 0.5, 4503599627370497.5, 1e300):
    print(v, round(v), round(v, 0), round(v, 1), round(v, 2), round(v, -1), round(v, -2))
for n in (0, 5, 15, 25, -15, -25, 12345, 2 ** 70 + 5):
    print(n, round(n), round(n, -1), round(n, -2), round(n, 1), round(n, -30))
for v in (2.5, -2.5, 0.0, -0.0, 1e20, 1.9999, -1.0001):
    print(int(v // 1), int(-((-v) // 1)), int(v), v // 1, v % 1, divmod(v, 1))
print(round(float("inf") if False else 1.0), type(round(2.5)).__name__, type(round(2.5, 0)).__name__, type(round(5, 1)).__name__)
for bad in (float("inf"), float("nan")):
    for f in (lambda: round(bad), lambda: round(bad, 1)):
        try:
            print(f())
        except (OverflowError, ValueError) as e:
            print(type(e).__name__, e)
