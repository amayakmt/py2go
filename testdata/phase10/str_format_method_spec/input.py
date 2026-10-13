print("{} {}".format(1, 2), "{1} {0} {1}".format("a", "b"), "{x}-{y}".format(x=1, y=2), "{0[0]} {0[1]}".format([5, 6]), "{p.real} {p.imag}".format(p=3 + 4j), "{{}} {{{}}}".format(7), "{:>5}|{:<5}|{:^5}|{:*^7}|{:05d}|{:+d}|{: d}".format("a", "b", "c", "d", 42, 5, 5))
print("{:b} {:o} {:x} {:X} {:#b} {:#o} {:#x} {:#X} {:08b} {:_b} {:,} {:_}".format(10, 10, 255, 255, 10, 10, 255, 255, 5, 1 << 20, 1234567, 1234567))
print("{:b} {:x} {:,} {:#x} {:_x}".format(2 ** 70, 2 ** 70, 2 ** 70, -2 ** 70, 2 ** 70), "{:>40}|".format(2 ** 100), "{:^45,}|".format(10 ** 30), "{:+d}".format(2 ** 65), "{:030d}".format(-2 ** 65))
print("{:c} {:c}".format(65, 0x1F600), "{!r} {!s} {!a}".format("é", "é", "é"), "{:10.4}|".format("abcdefgh"), "{:.0}|".format("abc"), "{:s}".format("x"), "{:5}|{:<5}|".format(1, 2), "{:5}|".format("ab"), "{:e}".format(12345), "{:%}".format(0.25), "{:.1%}".format(1))
for f in (lambda: "{}".format(), lambda: "{1}".format(1), lambda: "{x}".format(y=1), lambda: "{".format(1), lambda: "}".format(1), lambda: "{:d}".format("a"), lambda: "{:s}".format(1), lambda: "{:,x}".format(255), lambda: "{:c}".format(-1), lambda: "{:.2d}".format(5), lambda: "{:=s}".format("a"), lambda: "{0.x}".format(1), lambda: "{[0]}".format(1), lambda: "{:z}".format(1)):
    try:
        print(f())
    except (ValueError, TypeError, IndexError, KeyError, AttributeError, OverflowError) as e:
        print(type(e).__name__, e)
print("{:{}}|{:{}.{}}|{:{w}}".format(5, 4, 3.14159, 8, 3, 7, w=3), "{:>{}}".format("x", 4), "{0:{1}}".format("y", 3), "%s and %s" % ("a", "b"), "%(a)s-%(b)d" % {"a": "x", "b": 5}, "%5.1f|%-8s|%+05d|%x|%o|%e|%c|%%|%r" % (3.14159, "ab", 42, 255, 8, 1234.5, 65, "q"))
