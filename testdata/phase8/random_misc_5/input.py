from typing import Iterator, Generator

log: list[str] = []

def g0(n: int) -> Iterator[int]:
    log.append("start g0")
    try:
        for i in range(n):
            yield 7
            if i == 1:
                return
    finally:
        log.append("fin g0")
    log.append("end")


def g1(n: int) -> Iterator[int]:
    log.append("start g1")
    try:
        for i in range(n):
            yield i
            if i == 2:
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


print(list(g0(1)), log)
log.clear()
for v in g0(1):
    if v > 2:
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
    if v > 5:
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
print(acc.send(9))
print(acc.send(1))
print(acc.send(1))
try:
    acc.send(-1)
except StopIteration as e:
    print('done', e.value)
