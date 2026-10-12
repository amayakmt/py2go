def f(a: int, b: int = 2) -> int:
    return a * b


opts: dict[str, int] = {"a": 3, "c": 4}
try:
    print(f(**opts))
except TypeError as e:
    print("TypeError:", e)
good: dict[str, int] = {"a": 3, "b": 5}
print(f(**good))
try:
    print(f(1, **good))
except TypeError as e:
    print("TypeError:", e)
