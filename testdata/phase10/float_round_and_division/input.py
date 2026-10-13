print(round(0.5), round(1.5), round(2.5), round(-0.5), round(-1.5), round(0.49999999999999994), round(2.675, 2), round(1.005, 2), round(0.125, 2), round(0.375, 2))
print(round(1234.5678, 2), round(1234.5678, -2), round(1234.5678, 0), round(-1234.5678, 1), round(5e-324, 400), round(1e300, -299), round(123.456, 10), round(0.0, 2), round(-0.0, 1))
print(round(2 ** 53 + 1.0), round(1e22), round(-1e22), round(12345.6789e-3, 3), round(3.14159, 3), round(2.5, None), round(3.5), type(round(2.5)).__name__, type(round(2.5, 1)).__name__)
print(7 / 2, -7 / 2, 1 / 3, 2 / 3, 10 / 4, 6 / 3, 1e300 * 10 / 10, 5 / 1e-300, 0.1 + 0.2, 0.1 * 3, 1.1 * 1.1, 1e16 + 1, 1e15 + 0.3)
print(7 // 2.0, -7 // 2.0, 7.5 // 2, 7 % 2.5, -7 % 2.5, 7.5 % -2, 1e20 // 3, 2 ** 70 // 3.0, divmod(7.5, 2), divmod(-7.5, 2), divmod(0.3, 0.1))
print(float("1e400"), float("-1e400"), float("1" * 400), float("1e-400"), float(" 1.5\n"), float("1_0.5"), float("+.5"), float("5."), float("INFINITY"), float("-nan"))
for bad in ("1,5", "1e", "0x10", "", "1 2", "--1", "1__0", "_1"):
    try:
        print(float(bad))
    except ValueError as e:
        print("ValueError", e)
print(2 ** 0.5, 10 ** -2, 10.0 ** 20, 0.0 ** 0, 4 ** 0.5, 2 ** -1074, 2.0 ** 1023)
for a, b in ((0.0, -1), (0, -1), (10.0, 400), (-0.0, -1)):
    try:
        print(a ** b)
    except (ZeroDivisionError, OverflowError) as e:
        print(type(e).__name__, e)
