z = 3 + 4j
w = complex(1, -2)
print(z, w, z + w, z - w, z * w, z / w, -z, +w, abs(z), z.real, z.imag, z.conjugate(), z == 3 + 4j, z != w)
print(1j, -1j, 0j, 1 + 0j, (1 + 2j) * 1j, 1j ** 2, (1 + 1j) ** 2, 2 ** 1j if False else 0, complex(0, 0), complex(1.5), complex("1+2j"), complex(" 3-4j "), complex("5j"), complex("2"))
print(repr(1.5 + 2.5j), repr(1e20 + 1j), repr(complex(-0.0, 1)), repr(complex(1, -0.0)), repr(complex(float("inf"), float("nan"))), str(-1 - 1j), repr(3j), repr(complex(1, 0)))
print(z * 2, 2 * z, z + 1, 1 + z, z - 1.5, 1.5 - z, z / 2, 2 / z, z ** 2, z ** -1, z ** 0, 2 ** z if False else 0)
print(hash(1 + 0j) == hash(1), hash(3j) == hash(3j), hash(1 + 2j) == hash(1 + 2j), 1 + 0j == 1, 0j == 0, 1j == 1j, {1 + 0j: "a"}[1], {1: "b"}[1 + 0j])
print(abs(-3 + 4j), (1 + 2j).conjugate(), complex(2, 3).real + complex(2, 3).imag)
print(isinstance(z, complex), isinstance(1, complex), type(z).__name__, bool(0j), bool(1j), sum([1j, 2, 3.5]), [1j * 1j, 2j ** 2])
for f in (lambda: z < w, lambda: z // 2, lambda: z % 2, lambda: z / 0, lambda: 0j ** -1, lambda: complex("1+2jk"), lambda: complex(1, "2"), lambda: int(z), lambda: float(z), lambda: divmod(z, 2)):
    try:
        print(f())
    except (TypeError, ZeroDivisionError, ValueError) as e:
        print(type(e).__name__, e)
