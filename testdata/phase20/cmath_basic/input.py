import cmath

z = complex(3, 4)
print(abs(z), cmath.phase(z), cmath.polar(z))
r = cmath.rect(2, cmath.pi / 2)
print(round(r.real, 12), r.imag)
w = cmath.exp(1j * cmath.pi)
print(cmath.sqrt(-1), cmath.sqrt(z), w.real, round(w.imag, 12))
print(cmath.log(z), cmath.log(8, 2), cmath.log10(100))
for f in (cmath.sin, cmath.cos, cmath.tan, cmath.asin, cmath.acos, cmath.atan,
          cmath.sinh, cmath.cosh, cmath.tanh, cmath.asinh, cmath.acosh, cmath.atanh):
    w = f(0.5 + 0.25j)
    print(f.__name__, round(w.real, 12), round(w.imag, 12))
print(cmath.isfinite(z), cmath.isinf(complex(cmath.inf, 0)), cmath.isnan(cmath.nanj))
print(cmath.isclose(1 + 1j, 1 + 1.0000000001j), cmath.isclose(1j, 2j))
print(cmath.inf, cmath.infj, cmath.nanj, cmath.tau)
print(cmath.sqrt(4), cmath.phase(-1), cmath.phase(complex(-1.0, -0.0)))
for bad in (lambda: cmath.log(0), lambda: cmath.exp(1000), lambda: cmath.atanh(1)):
    try:
        bad()
    except (ValueError, OverflowError) as e:
        print(type(e).__name__, e)
try:
    cmath.sqrt("x")
except TypeError as e:
    print("TypeError", e)
