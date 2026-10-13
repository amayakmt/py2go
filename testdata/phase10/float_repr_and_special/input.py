vals = [0.0, -0.0, 1.0, 0.1, 0.2 + 0.1, 1 / 3, 2 / 3, 1e15, 1e16, 1e-4, 1e-5, 123456789.123456789, 1.5e300, 5e-324, 1.7976931348623157e308, float("inf"), float("-inf"), 100.0, 1e22, 1e21, 12345678901234567890.0]
for v in vals:
    print(repr(v), str(v), f"{v}", v == v, v > 0, int(v) if v == v and abs(v) != float("inf") else None)
nan = float("nan")
print(nan, nan == nan, nan != nan, nan < 1, nan > 1, max(1.0, nan), min(nan, 1.0), -nan, abs(-nan), bool(nan))
inf = float("inf")
print(inf + 1, inf - inf, inf * 0, -inf * -inf, inf / inf, 1 / inf, inf > 10 ** 400, -inf < -(10 ** 400), inf == inf, inf ** 0, 0.0 ** 0)
print(float("1e3"), float(" 1_0.5 "), float("-INF"), float("+nan"), float("infinity"), float(".5"), float("5."), float("1E-2"), float(b"2.5"))
for bad in ("", "abc", "1.2.3", "1e", "--1", "0x10", "1 2"):
    try:
        float(bad)
    except ValueError as e:
        print(e)
