def f() -> int:
    x: int = 1

    def g() -> None:
        x = 2
    g()
    return x
