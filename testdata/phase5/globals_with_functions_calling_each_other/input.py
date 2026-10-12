calls: int = 0


def tick() -> None:
    global calls
    calls += 1


def twice() -> None:
    tick()
    tick()


twice()
twice()
print(calls)
tick()
print(calls)
