from typing import Callable


def compose(f: Callable[[int], int], g: Callable[[int], int]) -> Callable[[int], int]:
    return lambda x: f(g(x))


def curry(f: Callable[[int, int], int]) -> Callable[[int], Callable[[int], int]]:
    return lambda a: lambda b: f(a, b)


inc: Callable[[int], int] = lambda x: x + 1
dbl: Callable[[int], int] = lambda x: x * 2
print(compose(inc, dbl)(5), compose(dbl, inc)(5))
add: Callable[[int], Callable[[int], int]] = curry(lambda a, b: a + b)
print(add(3)(4))
add5: Callable[[int], int] = add(5)
print([add5(i) for i in range(3)])


def pipeline(fs: list[Callable[[int], int]], x: int) -> int:
    for f in fs:
        x = f(x)
    return x


print(pipeline([inc, dbl, inc, dbl], 1))
