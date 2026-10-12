class C:
    def __init__(self) -> None:
        self.x = 1


c: C = C()
del c.x
