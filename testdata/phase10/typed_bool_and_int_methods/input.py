n: int = 255
print(n.bit_length(), n.bit_count(), (n).to_bytes(2, "big"), n.real, n.imag, n.numerator, n.denominator, n.conjugate(), n.as_integer_ratio(), int.from_bytes(b"\x01\x02", "big"), n.__class__.__name__ if False else "int")
m: int = -17
print(m.bit_length(), m.bit_count(), abs(m).bit_length(), m.to_bytes(2, "big", signed=True), (-m).to_bytes(1, "little"), divmod(m, 5), divmod(-m, -5), m // 5, m % 5, pow(m, 3), pow(m, 3, 7), pow(2, 10, 1000))
f: float = 2.75
print(f.is_integer(), f.as_integer_ratio(), f.hex(), (4.0).is_integer(), f.real, f.imag, f.conjugate(), float.fromhex("0x1.6p+1"), int(f), round(f), round(f, 1), f.__floor__() if hasattr(f, "__floor__") else 2)
flag: bool = True
print(flag + flag, flag * 3, flag & False, flag | False, flag ^ True, not flag, int(flag), float(flag), str(flag), repr(flag), flag.bit_length(), flag.real, ~flag, -flag, +flag)
