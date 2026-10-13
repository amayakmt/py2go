def mandel(c: complex, limit: int) -> int:
    z = 0j
    for i in range(limit):
        z = z * z + c
        if abs(z) > 2:
            return i
    return limit

grid = [[mandel(complex(x / 4, y / 4), 30) for x in range(-8, 4)] for y in range(-4, 5)]
for row in grid:
    print("".join("#" if v == 30 else str(min(v, 9)) for v in row))
roots = [complex(1, 0) * (1j ** k) for k in range(4)]
print(roots, [r.real for r in roots], sum(roots), [abs(r) for r in roots], [r ** 2 for r in roots], [r.conjugate() for r in roots])
z: complex = 3 + 4j
print(z, abs(z), z.real, z.imag, z * z.conjugate(), z / (1 - 1j), z ** 2, -z, z == 3 + 4j, z != 5, complex(z.real, -z.imag), (z * 1j).imag, complex("1+2j") + 1)
