def log(prefix: str, *values: int) -> None:
    print(prefix, values)


def logged(*values: int) -> None:
    log("call", *values)


logged()
logged(1)
logged(1, 2, 3)


def pick(a: int, b: int = 0, *rest: int) -> int:
    return a + b + len(rest)


def wrapper(*args: int) -> int:
    return pick(*args)


print(wrapper(1), wrapper(1, 2), wrapper(1, 2, 3, 4))
