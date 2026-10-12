def f(a: int, b: int, /, c: int, *, d: int = 0) -> int:
    return a * 1000 + b * 100 + c * 10 + d


def g(x: int, /) -> int:
    return x + 1


print(f(1, 2, 3), f(1, 2, c=3, d=4), g(5))
