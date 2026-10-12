from typing import Callable


def greet() -> None:
    print("hello")


def twice(f: Callable[[], None]) -> None:
    f()
    f()


twice(greet)
twice(lambda: print("lambda"))
handlers: list[Callable[[], None]] = [greet, greet]
for h in handlers:
    h()


def pick(flag: bool) -> Callable[[int], int]:
    if flag:
        return lambda x: x + 1
    return lambda x: x - 1


print(pick(True)(10), pick(False)(10))
funcs: dict[bool, Callable[[int], int]] = {True: pick(True), False: pick(False)}
print(funcs[True](1), funcs[False](1))
