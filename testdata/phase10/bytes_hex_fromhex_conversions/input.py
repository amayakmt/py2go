b = bytes(range(0, 256, 17))
print(b.hex(), b.hex("-"), b.hex(":", 2), b.hex(" ", -3), bytes.fromhex("48 65 6c6C 6f"), bytes.fromhex(""), bytearray.fromhex("ff 00"), bytes.fromhex(b.hex()) == b)
for bad in ("4", "zz", "4 8 6", "48 6"):
    try:
        print(bytes.fromhex(bad))
    except ValueError as e:
        print(e)
print(int.from_bytes(b"\x01\x00", "big"), int.from_bytes(b"\x01\x00", "little"), int.from_bytes(b"\xff", "big", signed=True), int.from_bytes(b"\xff\xff", "little", signed=True), int.from_bytes(b"", "big"), int.from_bytes(bytes(20), "big"), int.from_bytes(b"\x80" + bytes(15), "big"), int.from_bytes(b"\x80" + bytes(15), "big", signed=True))
print((1024).to_bytes(2, "big"), (1024).to_bytes(4, "little"), (-1).to_bytes(2, "big", signed=True), (255).to_bytes(1, "big"), (0).to_bytes(0, "big"), (2 ** 70).to_bytes(10, "big"), (-2 ** 70).to_bytes(10, "little", signed=True), (2 ** 64 - 1).to_bytes(8, "big"))
for f in (lambda: (256).to_bytes(1, "big"), lambda: (-1).to_bytes(1, "big"), lambda: (128).to_bytes(1, "big", signed=True), lambda: (1).to_bytes(1, "middle"), lambda: int.from_bytes(b"a", "weird")):
    try:
        print(f())
    except (OverflowError, ValueError) as e:
        print(type(e).__name__, e)
print(list(b"\xde\xad\xbe\xef"), [hex(x) for x in b"\xde\xad"], b"\xde\xad\xbe\xef".hex().upper(), bytes(map(lambda x: x ^ 0x55, b"abc")), bytes(reversed(b"abc")), b"abc"[::-1] == bytes(reversed(b"abc")))
