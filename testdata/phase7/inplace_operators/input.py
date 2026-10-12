class Bag:
    def __init__(self) -> None:
        self.items: list[int] = []

    def __iadd__(self, v: int) -> "Bag":
        self.items.append(v)
        return self

    def __add__(self, v: int) -> "Bag":
        out: Bag = Bag()
        out.items = self.items + [v]
        return out


a: Bag = Bag()
b: Bag = a
a += 1
a += 2
print(a.items, b.items, a is b)
c: Bag = a + 3
print(c.items, a.items, c is a)
