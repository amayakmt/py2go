from typing import Iterator, Iterable


def naturals() -> Iterator[int]:
    n: int = 0
    while True:
        yield n
        n += 1


def takewhile(pred: int, src: Iterable[int]) -> Iterator[int]:
    for x in src:
        if x >= pred:
            return
        yield x


def pairwise(src: Iterable[int]) -> Iterator[tuple[int, int]]:
    prev: int = -1
    first: bool = True
    for x in src:
        if not first:
            yield (prev, x)
        first = False
        prev = x


print(list(takewhile(6, naturals())))
print(list(pairwise(takewhile(5, naturals()))))
print(list(zip(takewhile(3, naturals()), "abc")))
print(dict(zip("xyz", takewhile(10, naturals()))))
print(list(enumerate(takewhile(3, naturals()), 10)))
print(any(x > 4 for x in takewhile(100, naturals())), sum(takewhile(5, naturals())))
print(sorted(takewhile(5, naturals()), reverse=True), max(takewhile(5, naturals())), len(list(takewhile(5, naturals()))))
