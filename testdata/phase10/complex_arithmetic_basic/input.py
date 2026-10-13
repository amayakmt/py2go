a, b = 3 + 4j, 1 - 2j
print(a, b, a + b, a - b, a * b, a / b, -a, +a, abs(a), a.real, a.imag, a.conjugate(), a == b, a == 3 + 4j, a != b)
print(a + 1, 1 + a, a * 2, 2 * a, a / 2, 2 / a, a - 0.5, 0.5 - a, a ** 2, a ** 0, a ** -1, 2 ** b if False else 0, a ** 0.5 if False else 0, 1j ** 2, 1j * 1j)
print(repr(1j), repr(-1j), repr(0j), repr(1 + 0j), repr(complex(0, -0.0)), repr(complex(-0.0, 0)), repr(1e20j), repr(1.5 + 2.5j), repr(complex(1, float("inf"))), repr(complex(float("nan"), 1)), str(3j), repr(100j), repr(1e-7j))
print(complex(1, 2), complex(1), complex(), complex("1+2j"), complex(" 3j "), complex("-1.5"), complex(1.5, -0.5), complex(2j, 1), complex(a, b), bool(0j), bool(1j), bool(complex(0, 0.0)))
for bad in ("1 + 2j", "j", "1+2", "abc", ""):
    try:
        complex(bad)
    except ValueError as e:
        print(e)
for f in (lambda: 1j < 2j, lambda: a // b, lambda: a % b, lambda: divmod(a, b), lambda: 1 / 0j, lambda: 0j ** -1, lambda: int(a), lambda: float(a), lambda: round(a)):
    try:
        print(f())
    except (TypeError, ZeroDivisionError) as e:
        print(type(e).__name__, e)
print(sum([1j, 2, 3.5]), sum([1j] * 4), [x * x for x in (1j, 2j)], {1j: "a", 1j + 0: "b"}, 1j in [0, 1j], (1 + 1j) in {1 + 1j}, hash(1j), hash(3 + 4j))
