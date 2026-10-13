import math
import random

random.seed(1)
fns = [
    ("exp", lambda x, y: math.exp(x)), ("log", lambda x, y: math.log(y)), ("sin", lambda x, y: math.sin(x)), ("cos", lambda x, y: math.cos(x)),
    ("tan", lambda x, y: math.tan(x)), ("atan", lambda x, y: math.atan(x)), ("asin", lambda x, y: math.asin(x / 31)), ("acos", lambda x, y: math.acos(x / 31)),
    ("pow", lambda x, y: math.pow(y, x / 7)), ("atan2", lambda x, y: math.atan2(x, y - 5)), ("sinh", lambda x, y: math.sinh(x / 3)), ("cosh", lambda x, y: math.cosh(x / 3)),
    ("tanh", lambda x, y: math.tanh(x / 9)), ("log2", lambda x, y: math.log2(y)), ("log10", lambda x, y: math.log10(y)), ("expm1", lambda x, y: math.expm1(x / 9)),
    ("log1p", lambda x, y: math.log1p(y)), ("asinh", lambda x, y: math.asinh(x)), ("acosh", lambda x, y: math.acosh(y + 1)), ("atanh", lambda x, y: math.atanh(x / 31)),
    ("**", lambda x, y: y ** 0.37), ("exp2", lambda x, y: math.exp2(x / 5)), ("hypot", lambda x, y: math.hypot(x, y)), ("hypot3", lambda x, y: math.hypot(x, y, x * y)),
    ("dist", lambda x, y: math.dist((x, y), (y, x / 2))), ("gamma", lambda x, y: math.gamma(y / 3 + 0.1)), ("lgamma", lambda x, y: math.lgamma(y + 0.1)),
]
for name, f in fns:
    row = []
    for i in range(25):
        x = random.uniform(-30, 30)
        y = random.uniform(0.1, 20)
        row.append(f(x, y))
    print(name, row)
print(math.sin(1e22), math.cos(1e22), math.tan(1e100), math.exp(709.7), math.exp(-745), math.log(1e-320), math.log(2 ** 1000), math.pow(2, 0.5), math.pow(10, -3), 2 ** 0.5, 10 ** -3.0)
print(math.sin(math.pi), math.cos(math.pi / 2), math.atan(1) * 4, math.exp(1), math.log(10), math.log(8, 2), math.log(1000, 10), math.sqrt(2), math.hypot(3, 4), math.hypot(1e300, 1e300), math.hypot(5e-324, 5e-324))
print(math.gamma(5), math.gamma(0.5), math.gamma(-1.5), math.gamma(171), math.lgamma(100), math.lgamma(-0.5), math.erf(0.5))
for f, a in [(math.exp, 1000), (math.log, 0), (math.log, -1), (math.sqrt, -1), (math.acos, 2), (math.gamma, 0), (math.gamma, 172), (math.sin, math.inf), (math.atanh, 1), (math.pow, (0, -1)), (math.lgamma, 0)]:
    try:
        print(f(*a) if isinstance(a, tuple) else f(a))
    except (ValueError, OverflowError) as e:
        print(type(e).__name__, e)
