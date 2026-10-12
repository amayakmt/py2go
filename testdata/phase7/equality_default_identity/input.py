class T:
    def __init__(self, v: int) -> None:
        self.v = v


a: T = T(1)
b: T = T(1)
c: T = a
print(a == b, a == c, a != b, a is c, a is b, a is not b)
items: list[T] = [a, b]
print(c in items, T(1) in items, items.index(b), items.count(a))
items.remove(b)
print(len(items), items[0] is a)
print(items == [a], items == [b])
