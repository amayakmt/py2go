from typing import Callable


def repeat(times: int) -> Callable[[Callable[[str], str]], Callable[[str], str]]:
    def decorator(fn: Callable[[str], str]) -> Callable[[str], str]:
        def wrapper(s: str) -> str:
            out: str = ""
            for _ in range(times):
                out += fn(s)
            return out
        return wrapper
    return decorator


@repeat(3)
def shout(s: str) -> str:
    return s.upper() + "!"


print(shout("hey"))


def add(n: int) -> Callable[[Callable[[int], int]], Callable[[int], int]]:
    def deco(fn: Callable[[int], int]) -> Callable[[int], int]:
        return lambda x: fn(x) + n
    return deco


@add(100)
def base(x: int) -> int:
    return x


print(base(1), base(5))
