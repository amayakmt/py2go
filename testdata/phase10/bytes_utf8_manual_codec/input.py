def enc(cp):
    if cp < 0x80:
        return bytes([cp])
    if cp < 0x800:
        return bytes([0xC0 | cp >> 6, 0x80 | cp & 0x3F])
    if cp < 0x10000:
        return bytes([0xE0 | cp >> 12, 0x80 | (cp >> 6) & 0x3F, 0x80 | cp & 0x3F])
    return bytes([0xF0 | cp >> 18, 0x80 | (cp >> 12) & 0x3F, 0x80 | (cp >> 6) & 0x3F, 0x80 | cp & 0x3F])

def dec(b):
    out, i = [], 0
    while i < len(b):
        c = b[i]
        n = 0 if c < 0x80 else 1 if c >> 5 == 6 else 2 if c >> 4 == 14 else 3 if c >> 3 == 30 else -1
        if n < 0 or i + n >= len(b) + (0 if n else 1) - (0 if n else 1) and i + n > len(b) - 1 + 0 and n:
            raise ValueError("bad byte at %d" % i)
        cp = c if n == 0 else c & (0x3F >> n)
        for k in range(1, n + 1):
            cp = (cp << 6) | (b[i + k] & 0x3F)
        out.append(cp)
        i += n + 1
    return out

s = "aé€😀日本語\u0000~߿ࠀ￿\U0010ffff"
mine = b"".join(enc(ord(c)) for c in s)
print(mine == s.encode("utf-8"), mine.hex(), [hex(c) for c in dec(mine)], "".join(map(chr, dec(mine))) == s, len(mine), len(s))
for cp in (0x7f, 0x80, 0x7ff, 0x800, 0xffff, 0x10000, 0x10ffff):
    print(hex(cp), enc(cp).hex(), chr(cp).encode().hex(), len(enc(cp)))
for bad in (b"\xff", b"\xc3", b"\xe2\x82", b"\x80", b"\xf0\x9f\x98"):
    try:
        print(bad.decode("utf-8"))
    except UnicodeDecodeError as e:
        print("builtin", bad, e)
