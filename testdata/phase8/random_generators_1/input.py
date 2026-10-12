from typing import Iterator, Generator

log: list[str] = []

def g0(n: int) -> Iterator[int]:
    log.append("start g0")
    for i in range(n):
        if i > ((2 * i) + (n - 5)):
            yield 7
        else:
            log.append("skip " + str(i))
    log.append("end")


def g1(n: int) -> Iterator[int]:
    log.append("start g1")
    yield from g0(n)
    yield ((3 * n) - (3 + n))
    yield from [5, (8 - n)]
    log.append("end")


def accumulate() -> Generator[int, int, int]:
    total: int = 0
    while True:
        got: int = yield total
        if got < 0:
            return total
        total += got


print(list(g0(5)), log)
log.clear()
for v in g0(5):
    if v > 3:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g0(5)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g0(5)), sorted(g0(5))[:2])
print(list(g1(1)), log)
log.clear()
for v in g1(1):
    if v > 5:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g1(1)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g1(1)), sorted(g1(1))[:2])
acc: Generator[int, int, int] = accumulate()
next(acc)
print(acc.send(1))
print(acc.send(4))
print(acc.send(6))
print(acc.send(7))
try:
    acc.send(-1)
except StopIteration as e:
    print('done', e.value)
