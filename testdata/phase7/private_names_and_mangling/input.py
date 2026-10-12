class Secret:
    def __init__(self, v: int) -> None:
        self.__v = v
        self._shared = v + 1

    def reveal(self) -> int:
        return self.__v

    def __hidden(self) -> str:
        return "hidden" + str(self.__v)

    def public(self) -> str:
        return self.__hidden()


class Child(Secret):
    def __init__(self) -> None:
        super().__init__(5)
        self.__v = 99

    def own(self) -> int:
        return self.__v


c: Child = Child()
print(c.reveal(), c.own(), c._shared, c.public())
