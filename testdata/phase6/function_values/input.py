from typing import Callable


def square(x: int) -> int:
    return x * x


def negate(x: int) -> int:
    return -x


def apply(f: Callable[[int], int], x: int) -> int:
    return f(x)


def twice(f: Callable[[int], int], x: int) -> int:
    return f(f(x))


def compose(f: Callable[[int], int], g: Callable[[int], int]) -> Callable[[int], int]:
    def both(x: int) -> int:
        return f(g(x))
    return both


print(apply(square, 7), apply(negate, 7), twice(square, 3))
h: Callable[[int], int] = compose(square, negate)
print(h(4), compose(negate, square)(4))
chosen: Callable[[int], int] = square
print(chosen(5))
chosen = negate
print(chosen(5))
