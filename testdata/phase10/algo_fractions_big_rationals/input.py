def gcd(a, b):
    while b:
        a, b = b, a % b
    return abs(a)

class Q:
    def __init__(self, n, d=1):
        if d == 0:
            raise ZeroDivisionError("zero denominator")
        g = gcd(n, d)
        if d < 0:
            g = -g
        self.n, self.d = n // g, d // g
    def __add__(self, o):
        o = o if isinstance(o, Q) else Q(o)
        return Q(self.n * o.d + o.n * self.d, self.d * o.d)
    def __radd__(self, o):
        return self + o
    def __mul__(self, o):
        o = o if isinstance(o, Q) else Q(o)
        return Q(self.n * o.n, self.d * o.d)
    def __truediv__(self, o):
        o = o if isinstance(o, Q) else Q(o)
        return Q(self.n * o.d, self.d * o.n)
    def __eq__(self, o):
        return isinstance(o, Q) and (self.n, self.d) == (o.n, o.d)
    def __lt__(self, o):
        return self.n * o.d < o.n * self.d
    def __repr__(self):
        return f"{self.n}/{self.d}" if self.d != 1 else str(self.n)
    def __float__(self):
        return self.n / self.d

h = Q(0)
for k in range(1, 40):
    h = h + Q(1, k)
print(h, float(h), len(str(h.d)))
print(Q(10 ** 30, 3 * 10 ** 20), Q(-6, -4), Q(6, -4), Q(2 ** 80, 2 ** 70), Q(1, 3) + Q(1, 6), Q(1, 3) * 3, 1 + Q(1, 2), Q(2, 3) / Q(4, 9), sorted([Q(1, 2), Q(1, 3), Q(2, 3)]), Q(1, 2) == Q(2, 4))
s = Q(0)
for k in range(1, 30):
    s = s + Q(1, 2 ** k)
print(s, s.d == 2 ** 29)
try:
    Q(1, 0)
except ZeroDivisionError as e:
    print(e)
