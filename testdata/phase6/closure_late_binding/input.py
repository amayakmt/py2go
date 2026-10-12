from typing import Callable

fs: list[Callable[[], int]] = [lambda: i for i in range(3)]
print([f() for f in fs])

gs: list[Callable[[], int]] = []
for k in range(3):
    gs.append(lambda: k * 10)
print([g() for g in gs])
k = 99
print([g() for g in gs])


def make(i: int) -> Callable[[], int]:
    return lambda: i


hs: list[Callable[[], int]] = [make(j) for j in range(3)]
print([h() for h in hs])


def capture_default() -> list[Callable[[], int]]:
    out: list[Callable[[], int]] = []
    for n in range(3):
        def get(n: int = n) -> int:
            return n
        out.append(get)
    return out


print([f() for f in capture_default()])
