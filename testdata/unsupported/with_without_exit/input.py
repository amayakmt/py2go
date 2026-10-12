class C:
    def __enter__(self) -> int:
        return 1


with C() as x:
    pass
