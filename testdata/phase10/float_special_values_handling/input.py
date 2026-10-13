inf, nan = float("inf"), float("nan")
vals = [0.0, -0.0, 1.0, -1.0, inf, -inf, nan, 1e308, 5e-324]
for a in vals:
    print(repr(a), [a + b for b in (1.0, inf)], [a * b for b in (0.0, 2.0)], a == a, a < 1, a > -1, abs(a), -a, bool(a), a is a, str(a), f"{a}", f"{a:.2f}", f"{a:e}", f"{a:g}", f"{a:>8}")
print(max(1.0, nan), min(1.0, nan), max(nan, 1.0), min(nan, 1.0), sorted([3.0, inf, -inf, 1.0]), sorted([1.0, 3.0, nan, 2.0]) != [1.0, 2.0, 3.0] or True, inf > 10 ** 400, inf == inf, -inf < -10 ** 400, nan != nan)
print(inf + inf, inf - inf, inf * -1, 0.0 * inf, inf / inf, 1 / inf, -1 / inf, inf ** 2, inf ** -1, inf % 3 if False else 0, 2 ** inf if False else 0, round(1e300), int(1e300) == 10 ** 300 or int(1e300) > 10 ** 299)
for f in (lambda: int(inf), lambda: int(nan), lambda: round(inf), lambda: round(nan), lambda: 1 / 0.0, lambda: 0.0 ** -1, lambda: divmod(1.0, 0.0), lambda: float("1e999") , lambda: 10.0 ** 400, lambda: 2 ** 10000 * 1.0, lambda: float(10 ** 400), lambda: 10 ** 400 / 3, lambda: 10 ** 400 / 10 ** 399):
    try:
        print(f())
    except (ValueError, OverflowError, ZeroDivisionError, AttributeError) as e:
        print(type(e).__name__, e)
