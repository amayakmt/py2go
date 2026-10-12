def f(a: int, b: int, /, c: int, *, d: int = 0) -> int:
    return a * 1000 + b * 100 + c * 10 + d


print(f(1, 2, 3), f(1, 2, c=3), f(1, 2, c=3, d=4), f(1, 2, 3, d=4))
xs: list[int] = [1, 2]
print(f(*xs, *[3]))
ks: dict[str, int] = {"c": 3}
print(f(*xs, **ks))
try:
    print(f(*xs, **{"a": 1, "c": 3}))
except TypeError as e:
    print("TypeError:", e)
