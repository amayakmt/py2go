print(b"%d items" % 5, b"%s-%s" % (b"a", b"b"), b"%5d|%-5d|%05d" % (1, 2, 3), b"%x %X %o %c" % (255, 255, 8, 65), b"%r" % b"x", b"%a" % "é" if False else b"ok", b"%%" % (), b"%.2f" % 3.14159, b"%s" % bytearray(b"ba"), b"%(k)s" % {b"k": b"v"} if False else b"ok")
for f in (lambda: b"%s" % "str", lambda: b"%d" % "x", lambda: b"%s %s" % (b"a",), lambda: b"%s" % (b"a", b"b"), lambda: b"%y" % 1, lambda: b"%c" % 256, lambda: b"%c" % b"ab"):
    try:
        print(f())
    except (TypeError, ValueError, OverflowError) as e:
        print(type(e).__name__, e)
print(b"%s" % 5 if False else 0, b"%5s|%-5s|" % (b"ab", b"cd"), b"%.1s" % b"xyz", b"%i" % 7, b"%+d" % 3, b"% d" % 3, b"%e" % 12345.678, b"%g" % 1e-5)
