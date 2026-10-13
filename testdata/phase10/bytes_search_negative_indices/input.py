b = b"abracadabra"
print(b.find(b"a", -4), b.find(b"a", -4, -1), b.find(b"", 20), b.find(b"", 5, 3), b.count(b"a", -6), b.count(b"a", 2, -2), b.count(b"", 3, 5), b.count(b"", 20), b.index(b"a", 3), b.rfind(b"a", 0, -1), b.rfind(b"a", 5), b.startswith(b"ab", -11), b.startswith(b"", 20), b.endswith(b"bra", -3), b.endswith(b"ra", 0, -1))
for f in (lambda: b.find(b"a", "x"), lambda: b.count(b"a", 1.5), lambda: b.index(b"zz"), lambda: b.rindex(b"zz", 2), lambda: b.find(b"a", 10 ** 30), lambda: b.find(b"a", -10 ** 30)):
    try:
        print(f())
    except (TypeError, ValueError) as e:
        print(type(e).__name__, e)
