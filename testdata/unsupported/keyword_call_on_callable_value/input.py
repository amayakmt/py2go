from typing import Callable

f: Callable[[int], int] = lambda a: a
print(f(a=1))
