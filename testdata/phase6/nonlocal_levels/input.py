def level1() -> int:
    x: int = 1

    def level2() -> int:
        y: int = 10

        def level3() -> int:
            nonlocal x, y
            x += 1
            y += 1
            return x + y

        r: int = level3()
        return r + y

    out: int = level2()
    return out * 100 + x


print(level1())


def shadow() -> int:
    v: int = 1

    def inner() -> int:
        v: int = 50
        return v

    return inner() + v


print(shadow())
