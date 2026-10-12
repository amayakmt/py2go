def f(a: int, b: int) -> int:
    return a + b


xs: list[int] = [1]
try:
    print(f(*xs))
except TypeError as e:
    print("TypeError:", e)
ys: list[int] = [1, 2, 3]
try:
    print(f(*ys))
except TypeError as e:
    print("TypeError:", e)
print(f(*[4, 5]))
