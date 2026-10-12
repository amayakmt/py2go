from typing import Callable


def deco(fn: Callable[[str], str]) -> Callable[[str], str]:
    return fn


@deco
def f(x: int) -> int:
    return x
