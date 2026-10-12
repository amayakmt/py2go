counter: int = 0
log: list[str] = []


def bump() -> int:
    global counter
    counter += 1
    log.append("bump" + str(counter))
    return counter


def outer() -> int:
    def inner() -> int:
        return bump() + bump()
    return inner() + inner()


print(outer(), counter, log)
