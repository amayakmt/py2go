from typing import Callable


def make_counter() -> Callable[[], int]:
    count: int = 0

    def inc() -> int:
        nonlocal count
        count += 1
        return count

    return inc


a: Callable[[], int] = make_counter()
b: Callable[[], int] = make_counter()
print(a(), a(), a(), b(), a(), b())
