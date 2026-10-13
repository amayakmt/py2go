for f in (0.0, -0.0, 1.0, -1.5, 0.1, 3.14159, 1e100, 1e-300, 5e-324, 2.5e-310, 1.7976931348623157e308, 255.0, 0.5, float("inf"), float("-inf"), float("nan")):
    print(repr(f), f.hex() if f == f else "nan", f.is_integer() if f - f == 0 else None)
print(float.fromhex("0x1.8p+0"), float.fromhex("0x1.8"), float.fromhex("-0x0.1p4"), float.fromhex("  0x1p-2 "), float.fromhex("inf"), float.fromhex("0xA"), float.fromhex("0x.8p1"))
for f in (0.5, 0.1, -2.75, 3.0, 1e20, 1.5e-5):
    n, d = f.as_integer_ratio()
    print(f, n, d, n / d == f)
for bad in ("0xg", "", "1.0", "0x1p"):
    try:
        print(float.fromhex(bad))
    except ValueError as e:
        print("ValueError", e)
try:
    float("inf").as_integer_ratio()
except OverflowError as e:
    print("OverflowError", e)
try:
    float("nan").as_integer_ratio()
except ValueError as e:
    print("ValueError", e)
print((2.5).conjugate(), (2.5).real, (2.5).imag, (3.0).is_integer(), (3.5).is_integer())
