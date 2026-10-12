from typing import Callable

f: Callable[[int, int], int] = lambda a, b: a + b
xs: list[int] = [1, 2]
print(f(*xs))
