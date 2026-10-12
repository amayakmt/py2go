from typing import Iterator, Generator

log: list[str] = []

def g0(n: int) -> Iterator[int]:
    log.append("start g0")
    yield from range(n)
    yield (1 - n)
    log.append("end")


def g1(n: int) -> Iterator[int]:
    log.append("start g1")
    try:
        for i in range(n):
            yield (i - (i + i))
            if i == 3:
                return
    finally:
        log.append("fin g1")
    log.append("end")


def accumulate() -> Generator[int, int, int]:
    total: int = 0
    while True:
        got: int = yield total
        if got < 0:
            return total
        total += got


print(list(g0(3)), log)
log.clear()
for v in g0(3):
    if v > 5:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g0(3)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g0(3)), sorted(g0(3))[:2])
print(list(g1(3)), log)
log.clear()
for v in g1(3):
    if v > 2:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g1(3)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g1(3)), sorted(g1(3))[:2])
acc: Generator[int, int, int] = accumulate()
next(acc)
print(acc.send(5))
print(acc.send(5))
print(acc.send(2))
try:
    acc.send(-1)
except StopIteration as e:
    print('done', e.value)
