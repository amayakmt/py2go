from typing import Callable


def logged(fn: Callable[[int], int]) -> Callable[[int], int]:
    def wrapper(x: int) -> int:
        print("calling with", x)
        r: int = fn(x)
        print("returned", r)
        return r
    return wrapper


@logged
def double(x: int) -> int:
    return x * 2


print(double(4))
print(double(5) + 1)


def counted(fn: Callable[[int], int]) -> Callable[[int], int]:
    calls: list[int] = [0]

    def wrapper(x: int) -> int:
        calls[0] += 1
        print("call number", calls[0])
        return fn(x)
    return wrapper


@counted
def inc(x: int) -> int:
    return x + 1


print(inc(1), inc(2), inc(3))
