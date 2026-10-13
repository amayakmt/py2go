class Money:
    def __init__(self, cents):
        self.cents = cents
    def __int__(self):
        return self.cents // 100
    def __float__(self):
        return self.cents / 100
    def __index__(self):
        return self.cents
    def __bool__(self):
        return self.cents != 0
    def __neg__(self):
        return Money(-self.cents)
    def __pos__(self):
        return self
    def __abs__(self):
        return Money(abs(self.cents))
    def __add__(self, o):
        return Money(self.cents + (o.cents if isinstance(o, Money) else int(o * 100)))
    def __format__(self, spec):
        return format(self.cents / 100, spec or ".2f") + " USD"
    def __repr__(self):
        return f"Money({self.cents})"
    def __round__(self, n=0):
        return Money(round(self.cents, -2 + (n or 0) * 0))
    def __hash__(self):
        return hash(self.cents)
    def __eq__(self, o):
        return isinstance(o, Money) and self.cents == o.cents
    def __lt__(self, o):
        return self.cents < o.cents

m = Money(12345)
print(int(m), float(m), bool(m), bool(Money(0)), -m, +m, abs(Money(-5)), m + Money(55), m + 2, f"{m}", f"{m:.1f}", format(m), repr(m), str(m), round(m))
print(sorted([Money(3), Money(1), Money(2)]), min(Money(3), Money(1)), max([Money(3), Money(1)]), Money(1) == Money(1), Money(1) != Money(2), hash(Money(7)) == hash(7))
for f in (lambda: int(object()), lambda: float(object()), lambda: m + "x", lambda: Money(1) < 3):
    try:
        print(f())
    except (TypeError, IndexError, AttributeError, ValueError) as e:
        print(type(e).__name__, e)
