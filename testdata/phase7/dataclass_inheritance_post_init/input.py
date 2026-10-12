from dataclasses import dataclass


@dataclass
class Base:
    id: int
    name: str = "n"


@dataclass
class Derived(Base):
    extra: float = 1.5

    def __post_init__(self) -> None:
        self.name = self.name.upper()


d: Derived = Derived(7, "abc", 2.5)
print(d, d.id, d.name, d.extra)
e: Derived = Derived(1)
print(e)
print(isinstance(d, Base), d == Derived(7, "ABC", 2.5))
