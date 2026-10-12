from typing import Generator

out: list[str] = []


def printer() -> Generator[None, str, None]:
    while True:
        line: str = yield
        out.append("got " + line)


def doubler(target: Generator[None, str, None]) -> Generator[None, str, None]:
    while True:
        text: str = yield
        target.send(text + text)


sink: Generator[None, str, None] = printer()
next(sink)
mid: Generator[None, str, None] = doubler(sink)
next(mid)
for word in ["ab", "cd"]:
    mid.send(word)
print(out)
mid.close()
sink.close()
