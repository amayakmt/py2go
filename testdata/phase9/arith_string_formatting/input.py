def show(v):
    return "%s|%r|%d" % (v, v, 7)

print(show("a"), show(3), show([1, "b"]))
print("%5d|%-5d|%05d|%+d" % (42, 42, 42, 42))
print("%.2f|%8.3f|%e|%g" % (3.14159, 2.71828, 12345.678, 0.0001))
print("%s and %s" % ("x", "y"), "%d%%" % 50, "%c" % 65 if False else "A")
print("%x %X %o %#x" % (255, 255, 8, 255))
name, n = "pi", 3.14159
print(f"{name}={n:.2f} {n!r} {name!s:>6} {n:>10.3f} {42:05d} {42:x} {1234567:,}")
print(f"{'nested ' + name} {n * 2:.1f} {[1, 2][0]} {{literal}}")
print("{} {} {}".format(1, "two", 3.0), "{1} {0}".format("a", "b"), "{x}".format(x=5))
print(str(1.0), str(1e20), str(1e-5), repr("it's"), repr('say "hi"'), str(None), str(True))
