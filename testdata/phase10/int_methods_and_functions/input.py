n = 2 ** 70 + 5
print(n.bit_length(), (255).bit_length(), (0).bit_length(), (-255).bit_length(), (2 ** 64).bit_length(), (-2 ** 64).bit_length())
print((255).bit_count(), (2 ** 70 - 1).bit_count(), (-7).bit_count(), (0).bit_count())
print((10).as_integer_ratio(), (2 ** 70).as_integer_ratio(), (5).is_integer(), (5).conjugate(), (5).real, (5).imag, (5).numerator, (5).denominator)
print((1024).to_bytes(2, "big"), (1024).to_bytes(2, "little"), (255).to_bytes(1, "big"), (-1).to_bytes(2, "big", signed=True), (2 ** 70).to_bytes(10, "big"))
print(int.from_bytes(b"\x04\x00", "big"), int.from_bytes(b"\x04\x00", "little"), int.from_bytes(b"\xff\xff", "big", signed=True), int.from_bytes(b"\xff\xff", "big"), int.from_bytes(b"", "big"))
for args in ((256, 1, "big"), (-1, 1, "big"), (1, 1, "middle")):
    try:
        print(args[0].to_bytes(args[1], args[2]))
    except (OverflowError, ValueError) as e:
        print(type(e).__name__, e)
print(pow(2, 100), pow(2, 100, 1000), pow(3, -1, 7), pow(-3, 5, 7), pow(2, 5, -7), pow(5, 0, 1), pow(7, 10 ** 20, 13), pow(2, -2), pow(10, 2, 7))
for args in ((2, 3, 0), (2, -1, 4), (2.0, 3, 5)):
    try:
        print(pow(*args))
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
print(abs(-2 ** 70), divmod(-7, 2), divmod(7, -2), round(2 ** 70), round(-2 ** 70, -3), round(1234567, -2), round(1250, -2), round(1350, -2), round(15, -1), round(25, -1))
print(sum(range(10 ** 6)), sum([2 ** 62, 2 ** 62, 2 ** 62]), max(range(10)), min([2 ** 80, 3]), int.__name__, int(True) + int(False))
