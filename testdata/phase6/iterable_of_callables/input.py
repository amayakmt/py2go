from typing import Callable, Iterable


def apply(fs: Iterable[Callable[[int], int]], x: int) -> list[int]:
    return [f(x) for f in fs]


def sq(x: int) -> int:
    return x * x


print(apply([sq, sq], -6))
