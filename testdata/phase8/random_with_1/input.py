from typing import Iterator, Generator

log: list[str] = []

def g0(n: int) -> Iterator[int]:
    log.append("start g0")
    yield from range(n)
    yield 2
    log.append("end")


def g1(n: int) -> Iterator[int]:
    log.append("start g1")
    for i in range(n):
        if n > i:
            yield ((5 - i) - (0 * n))
        else:
            log.append("skip " + str(i))
    log.append("end")


def g2(n: int) -> Iterator[int]:
    log.append("start g2")
    for i in range(n):
        if (n - (i * n)) <= 6:
            yield (4 - i)
        else:
            log.append("skip " + str(i))
    log.append("end")


def accumulate() -> Generator[int, int, int]:
    total: int = 0
    while True:
        got: int = yield total
        if got < 0:
            return total
        total += got


print(list(g0(0)), log)
log.clear()
for v in g0(0):
    if v > 0:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g0(0)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g0(0)), sorted(g0(0))[:2])
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
print(list(g2(6)), log)
log.clear()
for v in g2(6):
    if v > 2:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g2(6)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g2(6)), sorted(g2(6))[:2])
acc: Generator[int, int, int] = accumulate()
next(acc)
print(acc.send(2))
print(acc.send(2))
print(acc.send(3))
try:
    acc.send(-1)
except StopIteration as e:
    print('done', e.value)
