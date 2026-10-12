from typing import Callable


class IntBox:
    def __init__(self, items: list[int]) -> None:
        self.items = items

    def map(self, f: Callable[[int], int]) -> "IntBox":
        return IntBox([f(x) for x in self.items])

    def filter(self, p: Callable[[int], bool]) -> "IntBox":
        return IntBox([x for x in self.items if p(x)])

    def reduce(self, f: Callable[[int, int], int], start: int) -> int:
        acc: int = start
        for x in self.items:
            acc = f(acc, x)
        return acc

    def __len__(self) -> int:
        return len(self.items)

    def __iter__(self) -> "IntBoxIter":
        return IntBoxIter(self.items)

    def __repr__(self) -> str:
        return "IntBox" + str(self.items)


class IntBoxIter:
    def __init__(self, items: list[int]) -> None:
        self.items = items
        self.i = 0

    def __iter__(self) -> "IntBoxIter":
        return self

    def __next__(self) -> int:
        if self.i >= len(self.items):
            raise StopIteration
        self.i += 1
        return self.items[self.i - 1]


b: IntBox = IntBox([1, 2, 3, 4, 5, 6])
print(b.map(lambda x: x * x).filter(lambda x: x % 2 == 0))
print(b.reduce(lambda a, x: a + x, 0), len(b), [x + 1 for x in b])
total: int = 0
for x in b.filter(lambda v: v > 3):
    total += x
print(total)
