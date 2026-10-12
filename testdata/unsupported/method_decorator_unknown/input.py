def deco(f: int) -> int:
    return f


class C:
    @deco
    def f(self) -> None:
        pass
