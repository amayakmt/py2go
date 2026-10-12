from typing import Callable


class Counter:
    def __init__(self, items: list[int]) -> None:
        self.items = items
        self.seen: dict[int, int] = {}

    def tally(self) -> dict[int, int]:
        for x in self.items:
            self.seen[x] = self.seen.get(x, 0) + 1
        return self.seen

    def squares(self) -> list[int]:
        return [x * x for x in self.items if x % 2 == 1]

    def scaled(self, k: int) -> Callable[[int], int]:
        def inner(x: int) -> int:
            return x * k + len(self.items)
        return inner

    def adders(self) -> list[Callable[[int], int]]:
        return [self.scaled(i) for i in range(3)]

    def total(self) -> int:
        acc: int = 0

        def walk(i: int) -> None:
            nonlocal acc
            if i >= len(self.items):
                return
            acc += self.items[i]
            walk(i + 1)
        walk(0)
        return acc

    def __len__(self) -> int:
        return len(self.items)

    def size(self) -> int:
        return len(self)


c: Counter = Counter([1, 2, 3, 3, 1, 1])
print(c.tally(), c.squares(), c.scaled(3)(4), [f(10) for f in c.adders()], c.total(), c.size())
