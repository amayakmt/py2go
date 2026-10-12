from typing import Iterator, Generator

log: list[str] = []

def g0(n: int) -> Iterator[int]:
    log.append("start g0")
    i: int = 0
    while i < n:
        yield 3
        i += 3
    log.append("end")


def g1(n: int) -> Iterator[int]:
    log.append("start g1")
    for i in range(n):
        if ((n * n) * n) == ((i + i) + (i + i)):
            yield i
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


print(list(g0(3)), log)
log.clear()
for v in g0(3):
    if v > 7:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g0(3)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g0(3)), sorted(g0(3))[:2])
print(list(g1(5)), log)
log.clear()
for v in g1(5):
    if v > 8:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g1(5)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g1(5)), sorted(g1(5))[:2])
acc: Generator[int, int, int] = accumulate()
next(acc)
print(acc.send(5))
print(acc.send(9))
try:
    acc.send(-1)
except StopIteration as e:
    print('done', e.value)
