from typing import Generator, Iterator


def inner() -> Generator[int, None, str]:
    yield 1
    yield 2
    return "inner-result"


def outer() -> Generator[int, None, int]:
    r: str = yield from inner()
    print("got", r)
    yield 3
    yield from [10, 20]
    yield from range(2)
    return 7


print(list(outer()))
o: Generator[int, None, int] = outer()
out: list[int] = []
try:
    while True:
        out.append(next(o))
except StopIteration as e:
    print(out, e.args)


def walk(depth: int) -> Iterator[int]:
    yield depth
    if depth > 0:
        yield from walk(depth - 1)
        yield from walk(depth - 1)


print(list(walk(3)))
