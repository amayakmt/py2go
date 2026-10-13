import math
print(math.log(1), math.log(math.e), math.log(8, 2), math.log(100, 10), math.log10(1000), math.log2(1024), math.log1p(0))
print(math.log(2 ** 1000), math.log2(2 ** 2000), math.log10(10 ** 400), math.log(10 ** 500, 10), math.log2(3 ** 1000) > 1584, math.log(2 ** 70, 2), math.log10(2 ** 70))
print(math.floor(2 ** 70 + 0.5), math.ceil(10 ** 20), math.trunc(2 ** 80), math.sqrt(10 ** 20), math.exp(-1000), math.exp(709), math.pow(2, 10), math.pow(2.5, 2), math.pow(-8, 1 / 3) if False else 0, math.pow(0, 0), math.pow(2, -2))
for f in (lambda: math.pow(0, -1), lambda: math.pow(-8, 0.5), lambda: math.pow(10, 400), lambda: math.log(10, 1), lambda: math.log(-1, 10), lambda: math.sqrt(-1e-300), lambda: math.exp(710)):
    try:
        print(f())
    except (ValueError, OverflowError, ZeroDivisionError) as e:
        print(type(e).__name__, e)
print(math.floor(float("inf")) if False else "skip")
for bad in (float("inf"), float("nan")):
    for f in (math.floor, math.ceil, math.trunc):
        try:
            print(f(bad))
        except (OverflowError, ValueError) as e:
            print(f.__name__, type(e).__name__, e)
