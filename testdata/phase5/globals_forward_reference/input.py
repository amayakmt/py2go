def get() -> int:
    return VALUE * 2


def total() -> int:
    return get() + OFFSET


VALUE: int = 21
OFFSET: int = 1
print(get(), total())
