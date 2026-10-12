class NoMsg(Exception):
    pass


class Defaulted(Exception):
    def __init__(self, a: int = 1, b: str = "x") -> None:
        self.a = a
        self.b = b


print(str(NoMsg()), repr(NoMsg()))
print(str(NoMsg("one")), repr(NoMsg("one")))
print(str(NoMsg("one", 2)), repr(NoMsg("one", 2)))
print(str(Defaulted()), repr(Defaulted()))
print(str(Defaulted(5)), repr(Defaulted(5)))
