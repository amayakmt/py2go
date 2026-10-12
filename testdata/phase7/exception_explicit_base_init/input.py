class Base(Exception):
    def __init__(self, msg: str) -> None:
        Exception.__init__(self, msg.upper())


class Sub(Base):
    def __init__(self, n: int) -> None:
        Base.__init__(self, "n=" + str(n))
        self.n = n

    def __str__(self) -> str:
        return "Sub:" + Exception.__str__(self)


try:
    raise Sub(4)
except Base as e:
    print(e, e.args, isinstance(e, Sub))
