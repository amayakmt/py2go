from typing import Callable


def run(n: int, cb: Callable[[int], int]) -> int:
    return cb(n + 1)


def bounce(n: int) -> int:
    return run(n, bounce)


try:
    bounce(0)
except RecursionError as e:
    print("caught", e)


def fact(n: int) -> int:
    def walk(k: int) -> int:
        if k <= 1:
            return 1
        return k * walk(k - 1)
    return walk(n)


print(fact(10))


def forever() -> None:
    def loop(n: int) -> int:
        return loop(n + 1)
    loop(0)


try:
    forever()
except RecursionError:
    print("nested function recursion stops too")
