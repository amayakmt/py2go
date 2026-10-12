class Ctx:
    def __init__(self, tag: str) -> None:
        self.tag = tag

    def __enter__(self) -> "Ctx":
        print("+", self.tag)
        return self

    def __exit__(self, exc_type: object, exc: object, tb: object) -> None:
        print("-", self.tag)


def f(n: int) -> int:
    with Ctx("f"):
        if n > 0:
            return n * 2
    return -1


def g() -> list[int]:
    out: list[int] = []
    for i in range(5):
        with Ctx("loop" + str(i)):
            if i == 1:
                continue
            if i == 3:
                break
            out.append(i)
    return out


print(f(3), f(0))
print(g())
