print("%d %i %u" % (3.9, -3.9, 7), "%s" % None, "%s" % [1, "a"], "%r" % "it's", "%s" % (1,), "%s %s" % ((1, 2), 3), "%d" % True, "%x" % -255, "%#x %#o %#b" % (255, 8, 5) if False else "%#x %#o" % (255, 8), "%5s|%-5s|%.2s" % ("ab", "cd", "xyz"), "%*d|%-*d|" % (5, 1, 4, 2), "%.*f" % (2, 3.14159))
print("%c%c" % (72, "i"), "%s" % "é", "%a" % "é", "%10.3s|" % "abcdef", "%05.1f %+.2e %G" % (-2.55, 12345.6789, 1e-10), "%d%%" % 50, "%3d|%-3d|%03d" % (1, 2, 3), "% d % d" % (5, -5), "%+s" % "x" if False else "ok", "%.0f %.0f %.0f" % (0.5, 1.5, 2.5), "%d" % 2 ** 70, "%x" % 2 ** 70, "%e" % 2 ** 70, "%s" % 2 ** 70, "%.3s" % 12345)
for f in (lambda: "%d" % "x", lambda: "%d %d" % (1,), lambda: "%d" % (1, 2), lambda: "%s %s" % "ab", lambda: "%(a)s" % (1,), lambda: "%(a)s" % {"b": 1}, lambda: "%z" % 1, lambda: "%" % (), lambda: "%c" % -1, lambda: "%c" % "ab", lambda: "%c" % 1.5, lambda: "%s" % {1: 2} if False else "%.f" % "a", lambda: "%(a)s %s" % {"a": 1}, lambda: "%s" % {"a": 1}, lambda: "%x" % 1.5, lambda: "%x" % "a"):
    try:
        print(f())
    except (ValueError, TypeError, KeyError, OverflowError) as e:
        print(type(e).__name__, e)
