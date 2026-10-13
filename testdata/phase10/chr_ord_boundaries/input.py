for cp in (0, 1, 31, 32, 126, 127, 128, 255, 256, 0x7ff, 0x800, 0xd7ff, 0xe000, 0xfffe, 0xffff, 0x10000, 0x10ffff):
    c = chr(cp)
    print(cp, ord(c), repr(c), len(c), len(c.encode("utf-8")), len(c.encode("utf-16-le")), len(c.encode("utf-32-le")), c.encode("utf-8").hex(), c.isalpha(), c.isspace(), c.upper() == c)
for bad in (-1, 0x110000, 10 ** 30, 2 ** 64, 1.5, "a", None):
    try:
        print(chr(bad))
    except (ValueError, TypeError, OverflowError) as e:
        print(type(e).__name__, e)
for bad in ("", "ab", 5, None, b"a", b"ab"):
    try:
        print(ord(bad))
    except TypeError as e:
        print(e)
print(ord(b"a"), ord(bytearray(b"z")), [ord(c) for c in "aé日😀"], "".join(map(chr, [72, 105, 0x1F600])), chr(ord("a") + 1), sorted("ba日", key=ord), max("abc", key=ord), list(map(ord, "AZaz")), [chr(i) for i in range(0x3b1, 0x3b1 + 5)], chr(0x41) * 3)
