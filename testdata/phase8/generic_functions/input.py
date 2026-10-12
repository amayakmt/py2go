from typing import Callable


def identity[T](x: T) -> T:
    return x


def first[T](xs: list[T]) -> T:
    return xs[0]


def pair[A, B](a: A, b: B) -> tuple[A, B]:
    return (a, b)


def apply[T, R](f: Callable[[T], R], x: T) -> R:
    return f(x)


def largest[T: (int, float)](xs: list[T]) -> T:
    best: T = xs[0]
    for x in xs:
        if x > best:
            best = x
    return best


def swap[T](a: T, b: T) -> tuple[T, T]:
    return b, a


print(identity(5), identity("s"), identity([1, 2]), identity(2.5))
print(first([3, 4]), first(["a", "b"]), first([[1], [2]]))
print(pair(1, "x"), pair("k", [1.5]))
print(apply(lambda n: n * 2, 21), apply(len, "four"), apply(str, 3))
print(largest([3, 9, 2]), largest([1.5, 0.5]))
print(swap(1, 2), swap("a", "b"))
e: list[int] = []
print(len(e), first(["only"]))


def count_up[T](xs: list[T], pred: Callable[[T], bool]) -> int:
    n: int = 0
    for x in xs:
        if pred(x):
            n += 1
    return n


print(count_up([1, 2, 3, 4], lambda v: v % 2 == 0), count_up(["a", "bb"], lambda s: len(s) > 1))
