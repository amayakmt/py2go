for f in (0.0, -0.0, 0.5, -0.5, 1.5, 2.5, 1e15, 1e16, 1e17, 1e18, 1e19, 2.0 ** 63, 2.0 ** 64, 1e22, 1e30, 1e100, 1.7976931348623157e308, -1e20):
    print(f, int(f), int(f) == f, float(int(f)) == f)
for n in (0, 1, 2 ** 53, 2 ** 53 + 1, 2 ** 53 + 2, 2 ** 63, 2 ** 64, 10 ** 22, 10 ** 100, 10 ** 308, 2 ** 1023, 2 ** 1024 - 1, 2 ** 1024 - 2 ** 970):
    try:
        print(n, float(n))
    except OverflowError as e:
        print("OverflowError", e)
for bad in (float("inf"), float("-inf"), float("nan")):
    try:
        print(int(bad))
    except (OverflowError, ValueError) as e:
        print(type(e).__name__, e)
print(int(True), int(" 12 "), int("+7"), int("-0"), int("0"), int("00012"), int(12.9), int(-12.9), float("12"), float(12), 7 // 2 * 2.0, int(7 / 2), round(7 / 2), int(1e3), 10 ** 3 == 1e3)
