import math
print(math.pi, math.e, math.tau, math.inf, -math.inf, math.isnan(math.nan), math.isinf(-math.inf), math.isfinite(1e308), math.isfinite(math.nan))
print(math.sqrt(2), math.sqrt(0), math.sqrt(1e300), math.sqrt(4), math.sqrt(2 ** 80), math.exp(0), math.exp(1), math.exp2(10), math.expm1(0))
print(math.floor(-0.5), math.ceil(-0.5), math.floor(2.5), math.ceil(2.5), math.trunc(-2.7), math.trunc(2.7), math.floor(7), math.ceil(2 ** 70), math.floor(1e20), math.fabs(-3), math.copysign(3, -0.0), math.copysign(-2, 5))
print(math.fmod(-7, 3), math.fmod(7, -3), math.fmod(7.5, 2), math.remainder(7, 3), math.remainder(8, 3), math.remainder(-5, 3), math.ldexp(0.5, 10), math.frexp(8.0), math.frexp(0.0), math.frexp(-3.5), math.modf(3.75), math.modf(-3.75))
print(math.hypot(3, 4), math.hypot(), math.hypot(1, 1, 1), math.hypot(5), math.dist((0, 0), (3, 4)), math.dist([1, 2, 3], [4, 6, 3]), math.degrees(math.pi), math.radians(180), math.atan2(1, 1), math.atan2(-1, -1))
print(math.isclose(0.1 + 0.2, 0.3), math.isclose(1, 1.1), math.isclose(1, 1.1, rel_tol=0.2), math.isclose(1e-10, 0, abs_tol=1e-9), math.isclose(math.inf, math.inf), math.isclose(math.nan, math.nan))
for f, a in ((math.sqrt, -1), (math.log, 0), (math.log, -1), (math.factorial, -1), (math.exp, 1000), (math.isqrt, -1), (math.acos, 2), (math.asin, -2), (math.sin, math.inf), (math.acosh, 0), (math.atanh, 1), (math.gamma, 0), (math.log2, 0), (math.log10, -5)):
    try:
        print(f(a))
    except (ValueError, OverflowError) as e:
        print(type(e).__name__, e)
for bad in ("x", None, [1], 1j):
    try:
        math.sqrt(bad)
    except TypeError as e:
        print(e)
