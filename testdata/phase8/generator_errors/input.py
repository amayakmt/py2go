from typing import Iterator


def boom() -> Iterator[int]:
    yield 1
    raise ValueError("inside")


g: Iterator[int] = boom()
print(next(g))
try:
    next(g)
except ValueError as e:
    print("ValueError:", e)
try:
    next(g)
except StopIteration:
    print("exhausted after error")


def leaks_stop() -> Iterator[int]:
    yield 1
    raise StopIteration


h: Iterator[int] = leaks_stop()
next(h)
try:
    next(h)
except RuntimeError as e:
    print("RuntimeError:", e)


def reentrant() -> Iterator[int]:
    yield 1


r: Iterator[int] = reentrant()
print(list(r), list(r))
