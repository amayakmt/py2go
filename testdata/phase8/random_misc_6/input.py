from typing import Iterator, Generator

log: list[str] = []

def g0(n: int) -> Iterator[int]:
    log.append("start g0")
    yield from range(n)
    yield n
    log.append("end")


def g1(n: int) -> Iterator[int]:
    log.append("start g1")
    yield from g0(n)
    yield ((n * n) * 0)
    yield from [4, 5]
    log.append("end")


def accumulate() -> Generator[int, int, int]:
    total: int = 0
    while True:
        got: int = yield total
        if got < 0:
            return total
        total += got


print(list(g0(1)), log)
log.clear()
for v in g0(1):
    if v > 0:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g0(1)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g0(1)), sorted(g0(1))[:2])
print(list(g1(2)), log)
log.clear()
for v in g1(2):
    if v > 2:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g1(2)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g1(2)), sorted(g1(2))[:2])
acc: Generator[int, int, int] = accumulate()
next(acc)
print(acc.send(6))
print(acc.send(5))
try:
    acc.send(-1)
except StopIteration as e:
    print('done', e.value)
