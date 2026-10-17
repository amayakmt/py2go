"""cmath: mathematical functions for complex numbers. The functions are
native (Go's math/cmplx)."""

pi = 3.141592653589793
e = 2.718281828459045
tau = 6.283185307179586
inf = float("inf")
infj = complex(0.0, inf)
nan = float("nan")
nanj = complex(0.0, nan)


def phase(z, /):
    return _native("cmath.phase", z)


def polar(z, /):
    return _native("cmath.polar", z)


def rect(r, phi, /):
    return _native("cmath.rect", r, phi)


def exp(z, /):
    return _native("cmath.exp", z)


def log(x, base=None, /):
    if base is None:
        return _native("cmath.log", x)
    return _native("cmath.log", x, base)


def log10(z, /):
    return _native("cmath.log10", z)


def sqrt(z, /):
    return _native("cmath.sqrt", z)


def acos(z, /):
    return _native("cmath.acos", z)


def asin(z, /):
    return _native("cmath.asin", z)


def atan(z, /):
    return _native("cmath.atan", z)


def cos(z, /):
    return _native("cmath.cos", z)


def sin(z, /):
    return _native("cmath.sin", z)


def tan(z, /):
    return _native("cmath.tan", z)


def acosh(z, /):
    return _native("cmath.acosh", z)


def asinh(z, /):
    return _native("cmath.asinh", z)


def atanh(z, /):
    return _native("cmath.atanh", z)


def cosh(z, /):
    return _native("cmath.cosh", z)


def sinh(z, /):
    return _native("cmath.sinh", z)


def tanh(z, /):
    return _native("cmath.tanh", z)


def isfinite(z, /):
    return _native("cmath.isfinite", z)


def isinf(z, /):
    return _native("cmath.isinf", z)


def isnan(z, /):
    return _native("cmath.isnan", z)


def isclose(a, b, *, rel_tol=1e-09, abs_tol=0.0):
    return _native("cmath.isclose", a, b, rel_tol, abs_tol)
