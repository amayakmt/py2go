vals = [0, 1, -1, 255, 2 ** 70, -2 ** 70, 0.0, -0.0, 1.5, -1.5, 1e22, 1e-7, 2.5, True, False, "12", " -7 ", "1e3" if False else "3.5"]
for v in vals:
    for fn in (int, float, bool, str, repr, abs if not isinstance(v, str) else len, hex if isinstance(v, int) else oct if False else repr, bin if isinstance(v, int) else str):
        try:
            print(fn.__name__, repr(v)[:12], repr(fn(v))[:40], end=" | ")
        except (TypeError, ValueError) as e:
            print(fn.__name__, "!", type(e).__name__, end=" | ")
    print()
print(complex(1), complex(1.5, -2), complex("1+1j"), float(complex(2, 0)) if False else 0, divmod(7.5, 2), divmod(-7.5, 2), divmod(2 ** 70, 10 ** 10), pow(2.0, 10), pow(2, 0.5), pow(-8, 1 / 3) if False else 0, pow(10, -2), pow(0.0, 0), pow(0, 0.0))
print(int(2.9999999999999999), int(-0.9), int("٣"), float("1_0"), int(1e15), int(1e16), int(1e17), int(float(2 ** 63)), int(float(-2 ** 63)), float(2 ** 53 + 1), float(2 ** 53 + 3), int(float(2 ** 53 + 3)), float(10 ** 22), float(10 ** 23), int(1e23))
