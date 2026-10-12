class Fatal(Exception):
    def __init__(self, what: str) -> None:
        super().__init__(what)


def run(n: int) -> int:
    if n > 2:
        raise Fatal("too big: " + str(n))
    return n


print(run(1))
print(run(2))
print(run(3))
print("unreachable")
