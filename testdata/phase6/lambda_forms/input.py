from typing import Callable

f: Callable[[int], int] = lambda x: x + 1
g: Callable[[int, int], int] = lambda a, b: a * b
h: Callable[[], str] = lambda: "const"
print(f(1), g(3, 2), g(3, 4), h())
print((lambda x: x * 3)(4), (lambda a, b: a - b)(9, 4))
s: Callable[[str], str] = lambda t: t.strip().upper()
print(s("  hi "))
tup: Callable[[int], tuple[int, int]] = lambda n: (n, n * n)
print(tup(5))
lst: Callable[[int], list[int]] = lambda n: [i for i in range(n)]
print(lst(4))
