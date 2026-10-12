x: int = 1
items: list[int] = [1]


def shadow() -> int:
    x: int = 100
    items: list[int] = [9, 9]
    return x + len(items)


def read() -> int:
    return x + len(items)


print(shadow(), read(), x, items)
