from typing import Callable


class Button:
    def __init__(self, label: str) -> None:
        self.label = label
        self.handlers: list[Callable[[str], None]] = []
        self.on_double: Callable[[int], int] = lambda n: n * 2

    def subscribe(self, h: Callable[[str], None]) -> None:
        self.handlers.append(h)

    def click(self) -> None:
        for h in self.handlers:
            h(self.label)


class Recorder:
    def __init__(self) -> None:
        self.seen: list[str] = []

    def record(self, s: str) -> None:
        self.seen.append(s)


b: Button = Button("ok")
r: Recorder = Recorder()
b.subscribe(r.record)
b.subscribe(lambda s: print("clicked", s))
b.click()
b.click()
print(r.seen, b.on_double(21))
b.on_double = lambda n: n + 100
print(b.on_double(1))
