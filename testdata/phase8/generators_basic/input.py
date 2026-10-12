from typing import Iterator


def count(n: int) -> Iterator[int]:
    i: int = 0
    while i < n:
        yield i
        i += 1


def squares(xs: list[int]) -> Iterator[int]:
    for x in xs:
        yield x * x


def fib() -> Iterator[int]:
    a: int = 0
    b: int = 1
    while True:
        yield a
        a, b = b, a + b


print(list(count(5)), sum(count(4)), max(squares([1, -5, 3])))
for v in squares([1, 2, 3]):
    print(v)
f: Iterator[int] = fib()
print([next(f) for _ in range(10)])
g: Iterator[int] = count(2)
print(next(g), next(g), next(g, -1), next(g, None))
print([x for x in count(6) if x % 2 == 0], sorted(squares([3, 1, 2])))
evens: Iterator[int] = (x for x in count(10) if x % 2 == 0)
print(list(evens))
