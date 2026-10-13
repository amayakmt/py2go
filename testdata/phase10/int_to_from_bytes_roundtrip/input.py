for n in (0, 1, 127, 128, 255, 256, 65535, 65536, 2 ** 31, 2 ** 32, 2 ** 63 - 1, 2 ** 63, 2 ** 64 - 1, 2 ** 64, 2 ** 100, 10 ** 30):
    ln = max(1, (n.bit_length() + 7) // 8)
    for order in ("big", "little"):
        b = n.to_bytes(ln, order)
        assert int.from_bytes(b, order) == n
    print(n, ln, n.to_bytes(ln, "big").hex(), n.to_bytes(ln + 2, "little").hex(), (n.bit_length() + 8) // 8)
for n in (-1, -128, -129, -255, -256, -2 ** 31, -2 ** 63, -2 ** 63 - 1, -2 ** 100):
    ln = (n.bit_length() + 8) // 8
    b = n.to_bytes(ln, "big", signed=True)
    print(n, ln, b.hex(), int.from_bytes(b, "big", signed=True) == n, int.from_bytes(b, "big"), n.to_bytes(ln + 1, "little", signed=True).hex())
for f in (lambda: (256).to_bytes(1, "big"), lambda: (-1).to_bytes(4, "big"), lambda: (128).to_bytes(1, "big", signed=True), lambda: (-129).to_bytes(1, "big", signed=True), lambda: (1).to_bytes(-1, "big"), lambda: (1).to_bytes(2, "mid"), lambda: int.from_bytes("a", "big"), lambda: int.from_bytes([1, 256], "big"), lambda: (1.5).to_bytes(1, "big")):
    try:
        print(f())
    except (OverflowError, ValueError, TypeError, AttributeError) as e:
        print(type(e).__name__, e)
print(int.from_bytes([1, 2, 3], "big"), int.from_bytes(bytearray(b"\x01\x00"), "little"), int.from_bytes(range(4), "big"), int.from_bytes(b"\xff" * 20, "big") == 2 ** 160 - 1, int.from_bytes(b"\x80" + b"\x00" * 19, "big", signed=True) == -2 ** 159)
