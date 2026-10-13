def poly_eval(coeffs, x):
    r = 0
    for c in coeffs:
        r = r * x + c
    return r

def durand_kerner(coeffs, iters=200):
    n = len(coeffs) - 1
    roots = [(0.4 + 0.9j) ** k for k in range(n)]
    for _ in range(iters):
        new = []
        for i, r in enumerate(roots):
            d = 1
            for j, s in enumerate(roots):
                if i != j:
                    d *= (r - s)
            new.append(r - poly_eval(coeffs, r) / d)
        roots = new
    return roots

rs = durand_kerner([1, 0, 0, -1])
print([f"{r.real:.6f}{r.imag:+.6f}j" for r in sorted(rs, key=lambda c: (round(c.real, 6), round(c.imag, 6)))])
rs = durand_kerner([1, -6, 11, -6])
print(sorted(round(r.real, 6) for r in rs), max(abs(r.imag) for r in rs) < 1e-9, [abs(poly_eval([1, -6, 11, -6], r)) < 1e-9 for r in rs])
w = [1j ** k for k in range(8)]
print(w, sum(w), [abs(v) for v in w], [v.conjugate() * v for v in w][:3])
print(poly_eval([1, 2, 3], 2 + 1j), poly_eval([1, 0, 1], 1j), poly_eval([2, -3j, 4], 1 - 1j), (1 + 2j) * (3 - 4j), (1 + 2j) / (3 - 4j), (1 + 2j) ** 3, (1 + 2j) ** -2, 2 ** (1 + 1j) if False else 0, abs(3 + 4j) ** 2, (1j) ** 1j if False else 0)
