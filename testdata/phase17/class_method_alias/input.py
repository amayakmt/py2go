class A:
    def __init__(self, v):
        self.v = v

    def __add__(self, o):
        return A(self.v + (o.v if isinstance(o, A) else o))

    __radd__ = __add__

    def show(self):
        return "A(%d)" % self.v

    display = show


print((A(1) + A(2)).show(), (A(3) + 4).display(), (5 + A(6)).show())
print(sum([A(1), A(2), A(3)]).show())
