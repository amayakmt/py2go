def total() -> int:
    return 5


def use() -> int:
    total: int = 100
    return total + 1


print(use())
print(total())
