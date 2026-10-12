class C:
    @property
    def x(self) -> int:
        return 1

    @x.deleter
    def x(self) -> None:
        pass
