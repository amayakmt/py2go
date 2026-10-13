for v in (0.0, -0.0, 1.0, -1.5, 0.1, 3.14, 1e-300, 1e300, 5e-324, 2.0 ** 60, 255.5):
    print(v.hex(), float.fromhex(v.hex()) == v, v.is_integer(), v.as_integer_ratio() if v == v else 0)
    n, d = v.as_integer_ratio()
    print(n / d == v, d & (d - 1) == 0)
print(float.fromhex("0x1.8p1"), float.fromhex("-0X.8P-1"), float.fromhex("  0x10  "), float.fromhex("inf"), float.fromhex("0x1p-1074"), float.fromhex("0x1.fffffffffffffp1023"))
for bad in ("0x", "1.8", "0x1.8p", "zz"):
    try:
        float.fromhex(bad)
    except ValueError as e:
        print(e)
try:
    float.fromhex("0x1p1024")
except OverflowError as e:
    print(e)
for f in (float("inf"), float("nan")):
    for fn in (lambda: f.as_integer_ratio(),):
        try:
            print(fn())
        except (OverflowError, ValueError) as e:
            print(type(e).__name__, e)
print((2.5).conjugate(), (3.0).real, (3.0).imag, 1.5.__trunc__() if hasattr(1.5, "__trunc__") else 1, (0.5).is_integer(), float.__name__)
