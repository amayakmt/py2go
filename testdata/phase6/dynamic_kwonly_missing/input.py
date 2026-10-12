def f(a: int, *, key: int, other: int = 0) -> int:
    return a + key + other


empty: dict[str, int] = {}
try:
    print(f(1, **empty))
except TypeError as e:
    print("TypeError:", e)
print(f(1, **{"key": 10}))
print(f(1, **{"key": 10, "other": 5}))
