"""The math module: the functions are native, the constants are not."""

pi = 3.141592653589793
e = 2.718281828459045
tau = 6.283185307179586
inf = float("inf")
nan = float("nan")

def sqrt(x, /):
    return _native("math.sqrt", x)

def cbrt(x, /):
    return _native("math.cbrt", x)

def exp(x, /):
    return _native("math.exp", x)

def exp2(x, /):
    return _native("math.exp2", x)

def expm1(x, /):
    return _native("math.expm1", x)

def sin(x, /):
    return _native("math.sin", x)

def cos(x, /):
    return _native("math.cos", x)

def tan(x, /):
    return _native("math.tan", x)

def asin(x, /):
    return _native("math.asin", x)

def acos(x, /):
    return _native("math.acos", x)

def atan(x, /):
    return _native("math.atan", x)

def sinh(x, /):
    return _native("math.sinh", x)

def cosh(x, /):
    return _native("math.cosh", x)

def tanh(x, /):
    return _native("math.tanh", x)

def asinh(x, /):
    return _native("math.asinh", x)

def acosh(x, /):
    return _native("math.acosh", x)

def atanh(x, /):
    return _native("math.atanh", x)

def fabs(x, /):
    return _native("math.fabs", x)

def degrees(x, /):
    return _native("math.degrees", x)

def radians(x, /):
    return _native("math.radians", x)

def log1p(x, /):
    return _native("math.log1p", x)

def erf(x, /):
    return _native("math.erf", x)

def erfc(x, /):
    return _native("math.erfc", x)

def gamma(x, /):
    return _native("math.gamma", x)

def lgamma(x, /):
    return _native("math.lgamma", x)

def isfinite(x, /):
    return _native("math.isfinite", x)

def isinf(x, /):
    return _native("math.isinf", x)

def isnan(x, /):
    return _native("math.isnan", x)

def floor(x, /):
    return _native("math.floor", x)

def ceil(x, /):
    return _native("math.ceil", x)

def trunc(x, /):
    return _native("math.trunc", x)

def log10(x, /):
    return _native("math.log10", x)

def log2(x, /):
    return _native("math.log2", x)

def frexp(x, /):
    return _native("math.frexp", x)

def modf(x, /):
    return _native("math.modf", x)

def fsum(x, /):
    return _native("math.fsum", x)

def nextafter(x, /):
    return _native("math.nextafter", x)

def ulp(x, /):
    return _native("math.ulp", x)

def atan2(x, y, /):
    return _native("math.atan2", x, y)

def copysign(x, y, /):
    return _native("math.copysign", x, y)

def fmod(x, y, /):
    return _native("math.fmod", x, y)

def remainder(x, y, /):
    return _native("math.remainder", x, y)

def pow(x, y, /):
    return _native("math.pow", x, y)

def ldexp(x, y, /):
    return _native("math.ldexp", x, y)

def comb(x, y, /):
    return _native("math.comb", x, y)

def log(x, base=None, /):
    if base is None:
        return _native("math.log", x)
    return _native("math.log", x, base)

def hypot(*coordinates):
    return _native("math.hypot", *coordinates)

def dist(p, q, /):
    return _native("math.dist", p, q)

def isclose(a, b, *, rel_tol=1e-09, abs_tol=0.0):
    return _native("math.isclose", a, b, rel_tol=rel_tol, abs_tol=abs_tol)

def factorial(n, /):
    return _native("math.factorial", n)

def perm(n, k=None, /):
    return _native("math.perm", n, k)

def gcd(*integers):
    return _native("math.gcd", *integers)

def lcm(*integers):
    return _native("math.lcm", *integers)

def isqrt(n, /):
    return _native("math.isqrt", n)

def prod(iterable, /, *, start=1):
    return _native("math.prod", iterable, start=start)

def sumprod(p, q, /):
    return _native("math.sumprod", p, q)

def fma(x, y, z, /):
    return _native("math.fma", x, y, z)
