s = "héllo wörld 日本語 😀!"
print(len(s), s[0], s[-1], s[1], s[-2], s[2:5], s[:5], s[-5:], s[::2], s[::-1], s[3:-3], s[100:], s[:100], s[-100:3], s[5:2], s[::-3], s[-1:-5:-1])
print([s[i] for i in (0, 1, 6, 7, 12, 13, 14, 15, 16)], s.index("w"), s.index("日"), s.find("😀"), s.find("!"), s.count("l"), s.count("ö"), s.rfind("l"), s.encode().find("日".encode()))
for i in (len(s), -len(s) - 1, 10 ** 30, -10 ** 30):
    try:
        print(s[i])
    except IndexError as e:
        print(e)
for bad in (1.5, "a", None, [1]):
    try:
        print(s[bad])
    except TypeError as e:
        print(e)
for sl in ((None, None, 0), (1, 2, 1.5)):
    try:
        print(s[slice(*sl)])
    except (ValueError, TypeError) as e:
        print(e)
t = "".join(chr(c) for c in (0x41, 0xe9, 0x4e2d, 0x1f600, 0x10ffff, 0xd7ff, 0xe000))
print(len(t), [hex(ord(c)) for c in t], len(t.encode("utf-8")), [len(c.encode("utf-8")) for c in t], t.encode("utf-16-le").hex(), t[::-1][0] == "\udfff" or True, repr(t))
print(s.split(" "), s.partition("ö"), s.title(), s.upper(), s.center(30, "*"), s.zfill(25), s.expandtabs(), s.isprintable(), s.isascii(), s.isalnum())
