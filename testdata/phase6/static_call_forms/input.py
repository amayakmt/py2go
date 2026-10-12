def f(a: int, b: int = 10, *rest: int, k: int = 1, **extra: int) -> int:
    return a + b + sum(rest) + k + sum(extra.values())


print(f(1))
print(f(1, 2))
print(f(1, 2, 3, 4))
print(f(1, k=5))
print(f(1, 2, 3, k=5, z=100, y=1000))
print(f(b=2, a=1))
print(f(1, z=1))
