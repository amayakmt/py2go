from typing import Callable

inc: Callable[[int], int] = lambda x: x + 1
add: Callable[[int, int], int] = lambda a, b: a + b
shout: Callable[[str], str] = lambda s: s.upper() + "!"
nothing: Callable[[], int] = lambda: 42
check: Callable[[int], bool] = lambda n: n % 2 == 0
print(inc(1), add(2, 3), shout("hey"), nothing(), check(4), check(5))
total: int = 10
adder: Callable[[int], int] = lambda x: x + total
print(adder(1))
total = 100
print(adder(1))
print((lambda x, y: x * y)(6, 7))
fs: list[Callable[[int], int]] = [lambda x: x * 2, lambda x: x * 3]
print([f(10) for f in fs])


def run(f: Callable[[int, int], int]) -> int:
    return f(3, 4)


print(run(lambda a, b: a * b), run(lambda a, b: a + b))
