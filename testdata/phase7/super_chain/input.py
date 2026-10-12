class A:
    def __init__(self) -> None:
        self.trace: list[str] = []
        self.trace.append("A.init")

    def hello(self) -> str:
        return "A"


class B(A):
    def __init__(self) -> None:
        super().__init__()
        self.trace.append("B.init")

    def hello(self) -> str:
        return "B>" + super().hello()


class C(B):
    def __init__(self) -> None:
        super().__init__()
        self.trace.append("C.init")

    def hello(self) -> str:
        return "C>" + super().hello()


c: C = C()
print(c.trace)
print(c.hello())
