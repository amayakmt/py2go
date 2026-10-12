def f(a: int, /, b: int, *, c: int, d: int = 4) -> int:
    return a * 1000 + b * 100 + c * 10 + d


print(f(1, 2, c=3))
print(f(1, b=2, c=3, d=5))
print(f(1, c=3, b=9))


def g(x: int, *, scale: int = 2, offset: int = 0) -> int:
    return x * scale + offset


print(g(5), g(5, scale=3), g(5, offset=1), g(5, offset=1, scale=10))
