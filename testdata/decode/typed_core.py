def add(a: int, b: float, /, c: str, *args, d: bool, **kw) -> int:
    x: int = a + 1
    x += 2
    y = -x
    if x < 3 < 4 and not y:
        return x // 2
    elif x != 0 or y:
        pass
    else:
        return
    while True:
        break
    for i in range(10, 0, -2):
        continue
    return 1.5


@decorator
def deflt(a: int = 1, *, k: int = 2) -> None:
    pass
