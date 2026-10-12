from typing import Callable


def tag(name: str) -> Callable[[Callable[[str], str]], Callable[[str], str]]:
    def deco(fn: Callable[[str], str]) -> Callable[[str], str]:
        def wrapper(s: str) -> str:
            return "<" + name + ">" + fn(s) + "</" + name + ">"
        return wrapper
    return deco


@tag("b")
@tag("i")
@tag("u")
def text(s: str) -> str:
    return s


print(text("hi"))


def identity(fn: Callable[[int], int]) -> Callable[[int], int]:
    return fn


@identity
@identity
def seven(x: int) -> int:
    return 7


print(seven(0))
