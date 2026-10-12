class Base:
    def who(self) -> str:
        return "Base"


class Left(Base):
    def who(self) -> str:
        return "Left>" + super().who()


class Right(Base):
    def who(self) -> str:
        return "Right>" + super().who()


class Both(Left, Right):
    def who(self) -> str:
        return "Both>" + super().who()


b: Both = Both()
print(b.who())
l: Left = Left()
print(l.who())
r: Right = b
print(r.who())
base: Base = b
print(base.who())
