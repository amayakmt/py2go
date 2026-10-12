from typing import Callable


def trace(fn: Callable[[int], int]) -> Callable[[int], int]:
    def wrapper(n: int) -> int:
        print("enter", n)
        return fn(n)
    return wrapper


def outer() -> int:
    @trace
    def helper(x: int) -> int:
        return x + 1
    return helper(1) + helper(2)


print(outer())

cache: dict[int, int] = {}


def memo(fn: Callable[[int], int]) -> Callable[[int], int]:
    def wrapper(n: int) -> int:
        if n in cache:
            return cache[n]
        r: int = fn(n)
        cache[n] = r
        return r
    return wrapper


@memo
def fib(n: int) -> int:
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)


print(fib(30), len(cache))
