from typing import Iterator


class Countdown:
    def __init__(self, start: int) -> None:
        self.start = start
        self.calls = 0

    def values(self) -> Iterator[int]:
        self.calls += 1
        n: int = self.start
        while n > 0:
            yield n
            n -= 1

    def pairs(self) -> Iterator[tuple[int, int]]:
        for a in self.values():
            for b in self.values():
                if a < b:
                    yield (a, b)

    def __iter__(self) -> Iterator[int]:
        return self.values()


c: Countdown = Countdown(3)
print(list(c.values()), c.calls)
print(list(c.pairs()), c.calls)
print(sum(c), list(c), c.calls)
total: int = 0
for v in c:
    total += v
print(total, c.calls)
