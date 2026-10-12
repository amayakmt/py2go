from typing import Iterator

log: list[str] = []


def noisy(n: int) -> Iterator[int]:
    log.append("start")
    for i in range(n):
        log.append("before " + str(i))
        yield i
        log.append("after " + str(i))
    log.append("end")


g: Iterator[int] = noisy(2)
print(log)
print(next(g))
print(log)
print(next(g))
print(log)
print(list(g), log)
for x in noisy(1):
    log.append("got " + str(x))
print(log[-5:])
