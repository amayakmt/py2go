def level3() -> int:
    xs: list[int] = [1, 2]
    return xs[7]


def level2() -> int:
    return level3() + 1


def level1() -> int:
    try:
        return level2()
    except ValueError:
        return -1


print(level1())
