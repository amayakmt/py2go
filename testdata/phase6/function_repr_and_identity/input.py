from typing import Callable


def f() -> None:
    pass


g: Callable[[], None] = f
fs: list[Callable[[], None]] = [f, g]
print(len(fs))
print(f.__name__)
