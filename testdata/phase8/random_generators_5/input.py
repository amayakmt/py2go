from typing import Iterator, Generator

log: list[str] = []

def g0(n: int) -> Iterator[int]:
    log.append("start g0")
    i: int = 0
    while i < n:
        yield ((i + 0) + (3 * i))
        i += 1
    log.append("end")


def g1(n: int) -> Iterator[int]:
    log.append("start g1")
    for i in range(n):
        if ((n * n) * (4 + 7)) <= (n + n):
            yield (n - (3 + 1))
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
    if v > 4:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g0(0)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g0(0)), sorted(g0(0))[:2])
print(list(g1(4)), log)
log.clear()
for v in g1(4):
    if v > 5:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g1(4)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g1(4)), sorted(g1(4))[:2])
acc: Generator[int, int, int] = accumulate()
next(acc)
print(acc.send(3))
print(acc.send(1))
print(acc.send(8))
try:
    acc.send(-1)
except StopIteration as e:
    print('done', e.value)
