from typing import Callable

fs: list[Callable[[int, int], int]] = [lambda a, b: a + b, lambda a, b: a * b]
for f in fs:
    print(f(3, 4))
print([f(2, 5) for f in fs])
print(fs[1](6, 7), len(fs))
