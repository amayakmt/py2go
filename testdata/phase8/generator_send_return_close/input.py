from typing import Generator


def accumulator() -> Generator[int, int, str]:
    total: int = 0
    while True:
        got: int = yield total
        if got < 0:
            return "done:" + str(total)
        total += got


acc: Generator[int, int, str] = accumulator()
print(next(acc))
print(acc.send(5), acc.send(10))
try:
    acc.send(-1)
except StopIteration as e:
    print("stop", e.args, e.value)


def closing() -> Generator[int, None, None]:
    try:
        yield 1
        yield 2
    finally:
        print("cleanup")


c: Generator[int, None, None] = closing()
print(next(c))
c.close()
try:
    next(c)
except StopIteration:
    print("closed")


def catcher() -> Generator[int, None, None]:
    while True:
        try:
            yield 1
        except ValueError as e:
            print("caught", e)
            yield 99


h: Generator[int, None, None] = catcher()
print(next(h))
print(h.throw(ValueError("injected")))
print(next(h))
