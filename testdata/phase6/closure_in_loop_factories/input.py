from typing import Callable


def make_adders() -> list[Callable[[int], int]]:
    out: list[Callable[[int], int]] = []
    for i in range(4):
        out.append(lambda x: x + i)
    return out


def make_late() -> list[Callable[[], int]]:
    out: list[Callable[[], int]] = []
    for i in range(3):
        out.append(lambda: i)
    return out


print([f(10) for f in make_adders()])
print([f() for f in make_late()])
