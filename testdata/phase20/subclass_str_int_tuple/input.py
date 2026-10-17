class Color(str):
    def shout(self):
        return self.upper() + "!"


class Point(tuple):
    def __new__(cls, x, y):
        return super().__new__(cls, (x, y))

    @property
    def x(self):
        return self[0]

    @property
    def y(self):
        return self[1]

    def __repr__(self):
        return "Point(%r, %r)" % (self.x, self.y)


class Meters(float):
    def __str__(self):
        return "%gm" % float(self)


class Flags(int):
    def __new__(cls, value, label="f"):
        obj = super().__new__(cls, value * 2)
        obj.label = label
        return obj


c = Color("red")
print(c, repr(c), c.shout(), c + "dish", len(c), c == "red", isinstance(c, str), c.startswith("r"), "%s!" % c, f"<{c}>")
d = {"red": 1}
print(d.get(c), {c: 2}[c], hash(c) == hash("red"), sorted([Color("b"), Color("a")]), "-".join([c, Color("x")]))
p = Point(1, 2)
print(p, p.x, p.y, len(p), p[0], tuple(p), p == (1, 2), isinstance(p, tuple), p + (3,), list(p), hash(p) == hash((1, 2)))
a, b = p
print(a, b)
m = Meters(2.5)
print(m, m * 2, m + 1, float(m), int(m), round(m), m > 2, isinstance(m, float), "%.1f" % m)
f = Flags(3, "x")
print(f, f.label, f + 1, f * 3, f << 1, f & 4, int(f), isinstance(f, int), [0, 1, 2, 3, 4, 5, 6][f], bin(f), f == 6, -f, abs(-f))


class Version(tuple):
    def __new__(cls, text):
        parts = tuple(int(p) for p in text.split("."))
        return super().__new__(cls, parts)

    def __str__(self):
        return ".".join(str(p) for p in self)


v1 = Version("1.10.2")
v2 = Version("1.9")
print(v1, repr(v1), v1 > v2, sorted([v1, v2]) == [v2, v1], max(v1), v1.count(1), v1.index(10))


class Name(str):
    def __new__(cls, first, last):
        return super().__new__(cls, first + " " + last)

    def initials(self):
        return "".join(w[0] for w in self.split())


n = Name("Ada", "Lovelace")
print(n, n.initials(), n.lower(), n.split(), int(Color("42")) + 1, float(Color("1.5")))
print(isinstance(n, Name), isinstance(n, str), type(n).__name__, issubclass(Name, str))
