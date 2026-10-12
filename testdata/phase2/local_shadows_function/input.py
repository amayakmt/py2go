def total() -> int:
    return 5


def use() -> int:
    value: int = total()
    return value + 1


print(use())
total: int = 10
print(total)
