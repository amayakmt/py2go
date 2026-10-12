from typing import Iterator, Iterable


def numbers(limit: int) -> Iterator[int]:
    n: int = 0
    while n < limit:
        yield n
        n += 1


def evens(src: Iterable[int]) -> Iterator[int]:
    for x in src:
        if x % 2 == 0:
            yield x


def squared(src: Iterable[int]) -> Iterator[int]:
    for x in src:
        yield x * x


def take(src: Iterable[int], k: int) -> Iterator[int]:
    for i, x in enumerate(src):
        if i >= k:
            return
        yield x


print(list(take(squared(evens(numbers(1000000))), 5)))
print(sum(squared(numbers(5))), list(zip(numbers(3), squared(numbers(3)))))
pairs: list[tuple[int, int]] = [(a, b) for a in numbers(3) for b in numbers(a)]
print(pairs)
print(dict(enumerate(squared(numbers(4)))))
print(any(x > 10 for x in squared(numbers(5))), all(x < 10 for x in squared(numbers(5))))
print(max(squared(numbers(6))), min(squared(numbers(6))), sorted(squared(numbers(4)), reverse=True))
