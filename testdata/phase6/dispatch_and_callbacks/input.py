from typing import Callable

events: list[str] = []


def on_start() -> None:
    events.append("start")


def on_stop() -> None:
    events.append("stop")


handlers: dict[str, list[Callable[[], None]]] = {"start": [on_start], "stop": [on_stop, on_stop]}


def fire(name: str) -> None:
    for h in handlers[name]:
        h()


fire("start")
fire("stop")
fire("start")
print(events)


def apply_all(fs: list[Callable[[int], int]], x: int) -> list[int]:
    return [f(x) for f in fs]


print(apply_all([lambda v: v + 1, lambda v: v * 2, abs], -4))
acc: list[int] = []
collect: Callable[[int], None] = lambda v: acc.append(v)
for i in range(3):
    collect(i)
print(acc)
