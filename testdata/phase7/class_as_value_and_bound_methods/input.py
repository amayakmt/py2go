from typing import Callable


class Item:
    def __init__(self, n: int) -> None:
        self.n = n

    def double(self) -> int:
        return self.n * 2

    def add(self, k: int) -> int:
        return self.n + k


items: list[Item] = list(map(Item, [1, 2, 3]))
print([i.double() for i in items])
f: Callable[[int], int] = items[0].add
print(f(10), f(20))
g: Callable[[], int] = items[2].double
items[2].n = 50
print(g())
make: Callable[[int], Item] = Item
print(make(7).double())
fs: list[Callable[[], int]] = [i.double for i in items]
print([h() for h in fs])
