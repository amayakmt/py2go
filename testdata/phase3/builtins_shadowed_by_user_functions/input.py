def max(a: int, b: int) -> int:
    print("my max")
    if a > b:
        return a
    return b


def sum(a: int, b: int) -> int:
    return a + b + 1000


print(max(1, 2))
print(sum(1, 2))
print(min(5, 3))
