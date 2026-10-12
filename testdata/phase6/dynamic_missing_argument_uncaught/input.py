def f(a: int, b: int) -> int:
    return a + b


xs: list[int] = [1]
print("before")
print(f(*xs))
print("unreachable")
