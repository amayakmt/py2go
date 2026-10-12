class Base:
    created: int = 0

    def __init__(self, tag: str) -> None:
        self.tag = tag
        Base.created += 1

    @classmethod
    def make(cls, tag: str) -> "Base":
        return cls(tag + "!")

    @classmethod
    def kind(cls) -> str:
        return cls.__name__

    @staticmethod
    def helper(x: int) -> int:
        return x * 2


class Child(Base):
    @classmethod
    def kind(cls) -> str:
        return "child of " + super().kind()


b: Base = Base.make("b")
c: Base = Child.make("c")
print(b.tag, c.tag, type(b).__name__, type(c).__name__)
print(Base.kind(), Child.kind(), c.kind(), b.kind())
print(Base.helper(4), Child.helper(5), c.helper(6), Base.created)
