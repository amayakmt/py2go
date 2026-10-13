zs = [1 + 1j, 3 - 4j, -2 + 0.5j, 1e10 + 1e-10j, 0.5j, 2.0, complex(0, 0), complex(1, float("inf")) if False else 1j]
for a in zs:
    print([a + b for b in zs[:4]], [a * b for b in zs[:4]], [a - b for b in zs[:4]])
for a in zs[:6]:
    for b in zs[:4]:
        print(a, b, a / b, a ** 2, b ** -1, a == b, abs(a - b) > 0)
print([z ** n for z in (1 + 1j, 2j, -1 + 0j) for n in (0, 1, 2, 3, -1, -2, 10)])
print([(1 + 2j) ** (0.5 + 0j) if False else 0, 2 ** 3j if False else 0, 1j ** 0, 0j ** 0, 0j ** 1, (1j) ** 2, (-1) ** 2 + 0j])
for f in (lambda: 1 / 0j, lambda: 0j ** -1, lambda: (1 + 1j) // 1, lambda: (1 + 1j) % 2, lambda: 1j < 1, lambda: abs("a"), lambda: complex(1, 2) ** (1 << 40) if False else complex("1+2j+3")):
    try:
        print(f())
    except (ZeroDivisionError, TypeError, ValueError, OverflowError) as e:
        print(type(e).__name__, e)
