print("start")


def first() -> None:
    print("first")


first()
count: int = 1


def second(n: int) -> int:
    return n * 2


print(second(count))
first()
