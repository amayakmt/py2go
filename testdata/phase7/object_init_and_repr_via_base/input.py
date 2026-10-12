class A:
    def __init__(self) -> None:
        object.__init__(self)
        self.v = 1

    def __repr__(self) -> str:
        return "A(" + str(self.v) + ")"

    def __str__(self) -> str:
        return "str-" + object.__repr__(self)[:0] + repr(self)


print(A(), repr(A()))
