from typing import Callable


def make_acc(start: int) -> Callable[[int], int]:
    total: int = start

    def add(n: int) -> int:
        nonlocal total
        total += n
        return total
    return add


a: Callable[[int], int] = make_acc(10)
b: Callable[[int], int] = make_acc(100)
print(a(1), a(2), b(5), a(3), b(5))
