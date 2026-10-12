from typing import Generator


def g() -> Generator[int, int, None]:
    x: int = yield 1


gen: Generator[int, int, None] = g()
next(gen)
gen.send("text")
