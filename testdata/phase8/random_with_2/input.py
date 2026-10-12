from typing import Iterator, Generator

log: list[str] = []

def g0(n: int) -> Iterator[int]:
    log.append("start g0")
    for i in range(n):
        if (i + (9 + i)) >= i:
            yield n
        else:
            log.append("skip " + str(i))
    log.append("end")


def g1(n: int) -> Iterator[int]:
    log.append("start g1")
    for i in range(n):
        if (n - 9) == ((n + i) + (i * n)):
            yield (i - (0 - i))
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


print(list(g0(2)), log)
log.clear()
for v in g0(2):
    if v > 9:
        break
    log.append(str(v))
print(log)
log.clear()
it: Iterator[int] = g0(2)
print(next(it, -1), next(it, -2))
log.clear()
print(sum(g0(2)), sorted(g0(2))[:2])
print(list(g1(2)), log)
log.clear()
for v in g1(2):
    if v > 0:
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
print(acc.send(5))
print(acc.send(7))
print(acc.send(3))
print(acc.send(0))
try:
    acc.send(-1)
except StopIteration as e:
    print('done', e.value)
