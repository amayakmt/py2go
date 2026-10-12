from typing import Generator, Iterator


def running_average() -> Generator[float, float, None]:
    total: float = 0.0
    count: int = 0
    average: float = 0.0
    while True:
        value: float = yield average
        total += value
        count += 1
        average = total / count


avg: Generator[float, float, None] = running_average()
next(avg)
print([avg.send(v) for v in [10.0, 20.0, 60.0]])


def chunks(xs: list[int], size: int) -> Iterator[list[int]]:
    for i in range(0, len(xs), size):
        yield xs[i:i + size]


print(list(chunks([1, 2, 3, 4, 5], 2)))


def interleave(a: list[int], b: list[int]) -> Iterator[int]:
    for x, y in zip(a, b):
        yield x
        yield y


print(list(interleave([1, 2, 3], [10, 20, 30])))


def primes() -> Iterator[int]:
    found: list[int] = []
    n: int = 2
    while True:
        if all(n % p != 0 for p in found):
            found.append(n)
            yield n
        n += 1


p: Iterator[int] = primes()
print([next(p) for _ in range(10)])
