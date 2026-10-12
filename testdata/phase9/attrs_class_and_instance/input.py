class Counter:
    total = 0
    label = "counter"

    def __init__(self, name):
        self.name = name
        self.count = 0
        Counter.total += 1

    def hit(self, n=1):
        self.count += n
        return self

a = Counter("a")
b = Counter("b")
a.hit().hit(5)
b.hit()
print(a.count, b.count, Counter.total, a.total, a.label)
a.label = "overridden"
print(a.label, b.label, Counter.label)
Counter.label = "changed"
print(a.label, b.label, Counter.label)
print(type(a).__name__, type(a) is Counter, isinstance(a, Counter), isinstance(a, object))
print(Counter.__name__, a.__class__.__name__, a.__class__ is b.__class__)
c = a.__class__("made")
print(c.name, Counter.total)
