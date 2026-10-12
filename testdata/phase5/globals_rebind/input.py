counter: int = 0
label: str = "start"
history: list[int] = []


def bump(by: int) -> int:
    global counter
    counter += by
    history.append(counter)
    return counter


def rename(name: str) -> None:
    global label
    label = name + "!"


bump(1)
bump(10)
rename("done")
print(counter, label, history)
counter = 100
print(bump(5), history)
