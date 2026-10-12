from typing import Iterator

log: list[str] = []


def resource(name: str) -> Iterator[str]:
    log.append("open " + name)
    try:
        yield name + "-1"
        yield name + "-2"
    finally:
        log.append("close " + name)


for item in resource("a"):
    log.append("saw " + item)
    break
print(log)
log.clear()
g: Iterator[str] = resource("b")
print(next(g), log)
g.close()
print(log)
log.clear()
try:
    for item in resource("c"):
        raise ValueError("stop")
except ValueError as e:
    print("error", e, log)
