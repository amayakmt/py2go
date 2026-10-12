from typing import Optional


def f(a: Optional[int], b: Optional[int]) -> int:
    if a is None:
        return -1
    if b is None:
        return a
    return a + b


def g(s: Optional[str]) -> str:
    if s:
        return s.upper()
    else:
        return "empty"


def h(a: Optional[int]) -> int:
    if a is not None and a > 3:
        return a
    return 0


def k(a: Optional[int]) -> str:
    return "none" if a is None else "v" + str(a)


print(f(None, 1), f(2, None), f(2, 3))
print(g(None), g(""), g("x"))
print(h(None), h(2), h(9))
print(k(None), k(4))
n: Optional[int] = 5
while n is not None and n > 0:
    print(n)
    if n == 3:
        n = None
    else:
        n -= 1
print(n)
